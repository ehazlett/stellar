package server

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/containerd/containerd/errdefs"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (s *Server) initNetworking() error {
	c, err := s.client()
	if err != nil {
		return err
	}
	defer c.Close()

	// check for existing assigned subnet; if not, allocate
	localSubnetKey := fmt.Sprintf(dsSubnetsKey, s.NodeName())

	subnets, err := c.Network().Subnets()
	if err != nil {
		return err
	}

	if len(subnets) == 0 {
		return fmt.Errorf("no available subnets in network configuration")
	}

	bSubnetCIDR, err := c.Datastore().Get(dsNetworkBucketName, localSubnetKey)
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return err
		}
	}

	if bytes.Equal(bSubnetCIDR, []byte("")) {
		logrus.Debug("local subnet key not found; assigning new subnet")

		searchKey := fmt.Sprintf(dsSubnetsKey, "")
		existingSubnets, err := c.Datastore().Search(dsNetworkBucketName, searchKey)
		if err != nil {
			err = errdefs.FromGRPC(err)
			if !errdefs.IsNotFound(err) {
				return err
			}
		}

		assigned := len(existingSubnets)
		if len(subnets) < assigned {
			return fmt.Errorf("no available subnet for current node; need %d subnets", assigned)
		}

		bSubnetCIDR = []byte(subnets[assigned].CIDR)
		if err := c.Datastore().Set(dsNetworkBucketName, localSubnetKey, bSubnetCIDR, true); err != nil {
			return err
		}
	}

	subnetCIDR := string(bSubnetCIDR)
	logrus.Infof("setting up subnet %s", subnetCIDR)
	ip, ipnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return err
	}

	gw := ip.Mask(ipnet.Mask)
	gw[3]++

	mask, _ := ipnet.Mask.Size()

	logrus.Debugf("setting up local gateway %s", gw.String())
	if err := s.setupGateway(gw, mask); err != nil {
		return err
	}

	return nil
}

func (s *Server) getBindDeviceName() (string, error) {
	bindAddr := s.config.AgentConfig.BindAddr
	bindIP := net.ParseIP(bindAddr)
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logrus.Warnf("error getting addresses for interface %s", iface.Name)
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				logrus.Warnf("error parsing address %s", addr)
				continue
			}
			if bindIP.Equal(ip) {
				return iface.Name, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find interface for bind addr %s", bindAddr)
}

func (s *Server) setupGateway(ip net.IP, mask int) error {
	c, err := s.client()
	if err != nil {
		return err
	}
	defer c.Close()

	bindIP := s.getBindIP()

	bindInterface, err := s.getBindDeviceName()
	if err != nil {
		return err
	}

	logrus.Debugf("bind interface: %s", bindInterface)
	dev, err := netlink.LinkByName(bindInterface)
	if err != nil {
		return err
	}

	target := fmt.Sprintf("%s/%d", ip.String(), mask)
	aliasAddr, err := netlink.ParseAddr(target)
	if err != nil {
		return err
	}
	aliasAddr.Label = dev.Attrs().Name + ".stellar-gw"

	logrus.Debugf("adding address %s to device %s", aliasAddr, dev.Attrs().Name)
	if err := netlink.AddrReplace(dev, aliasAddr); err != nil {
		return err
	}

	// add route
	_, ipnet, err := net.ParseCIDR(target)
	if err != nil {
		return err
	}
	networkCIDR := fmt.Sprintf("%s/%d", ipnet.IP.String(), mask)
	localRouteKey := fmt.Sprintf(dsRoutesKey, s.NodeName())
	routeData := []byte(networkCIDR + ":" + bindIP.String())
	if err := c.Datastore().Set(dsNetworkBucketName, localRouteKey, routeData, true); err != nil {
		return err
	}

	return nil
}

func (s *Server) getBindIP() net.IP {
	bindAddr := s.config.AgentConfig.BindAddr
	bindIP := net.ParseIP(bindAddr)

	return bindIP
}

func (s *Server) setupRoutes() error {
	c, err := s.client()
	if err != nil {
		return err
	}
	defer c.Close()

	bindIP := s.getBindIP()

	searchKey := fmt.Sprintf(dsRoutesKey, "")
	routes, err := c.Datastore().Search(dsNetworkBucketName, searchKey)
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return err
		}
	}

	deviceName, err := s.getBindDeviceName()
	if err != nil {
		return err
	}

	dev, err := netlink.LinkByName(deviceName)
	if err != nil {
		return err
	}

	for _, routeSpec := range routes {
		v := string(routeSpec.Value)
		rt := strings.Split(v, ":")
		if len(rt) != 2 {
			logrus.Errorf("invalid route format: %s", v)
			continue
		}

		network := rt[0]
		gw := rt[1]

		_, ipnet, err := net.ParseCIDR(network)
		if err != nil {
			logrus.Warnf("error setting up route %s", network)
			continue
		}

		ip := net.ParseIP(gw)

		if ip.Equal(bindIP) {
			logrus.Debugf("skipping local route %s", network)
			continue
		}

		logrus.Debugf("configuring route %s via %s", network, gw)
		route := &netlink.Route{
			LinkIndex: dev.Attrs().Index,
			Dst:       ipnet,
			Gw:        ip,
		}
		if err := netlink.RouteReplace(route); err != nil {
			return err
		}
	}

	return nil
}
