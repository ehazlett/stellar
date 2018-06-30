package server

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	// TODO: make configurable
	bridgeName = "stellar0"
)

func (s *Server) initNetworking() error {
	c, err := s.client()
	if err != nil {
		return err
	}
	defer c.Close()

	subnetCIDR, err := c.Network().AllocateSubnet(s.NodeName())
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

	// configure container networking
	if err := s.initContainerNetworking(subnetCIDR, gw); err != nil {
		return err
	}

	return nil
}

func (s *Server) initContainerNetworking(subnetCIDR string, gw net.IP) error {
	client, err := s.client()
	if err != nil {
		return err
	}
	defer client.Close()

	containers, err := client.Node().Containers()
	if err != nil {
		return err
	}

	for i, container := range containers {
		if !container.Running() {
			logrus.Debugf("container %s not running; skipping networking", container.ID)
			continue
		}
		logrus.WithFields(logrus.Fields{
			"id":     container.ID,
			"labels": container.Labels,
			"pid":    container.Task.Pid,
		}).Debug("configuring container network")
		// TODO: allocate IP from network service
		ip := fmt.Sprintf("172.16.0.%d", i+10)
		if err := client.Node().SetupContainerNetwork(container.ID, ip, subnetCIDR, gw.String(), bridgeName); err != nil {
			logrus.Errorf("error setting up networking for container %s: %s", container.ID, err)
			continue
		}
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

	bindInterface, err := s.getBindDeviceName()
	if err != nil {
		return err
	}

	logrus.Debugf("bind interface: %s", bindInterface)

	target := fmt.Sprintf("%s/%d", ip.String(), mask)
	brAddr, err := netlink.ParseAddr(target)
	if err != nil {
		return err
	}

	brLink, err := netlink.LinkByName(bridgeName)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return err
		}

		// setup
		brLink = &netlink.Bridge{
			LinkAttrs: netlink.LinkAttrs{
				Name: bridgeName,
			},
		}
		if err := netlink.LinkAdd(brLink); err != nil {
			return err
		}

		if err := netlink.LinkSetUp(brLink); err != nil {
			return err
		}
	}
	if err := netlink.AddrReplace(brLink, brAddr); err != nil {
		return err
	}

	bindIP := s.getBindIP()

	// add route
	_, ipnet, err := net.ParseCIDR(target)
	if err != nil {
		return err
	}
	networkCIDR := fmt.Sprintf("%s/%d", ipnet.IP.String(), mask)
	if err := c.Network().AddRoute(networkCIDR, bindIP.String()); err != nil {
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

	deviceName, err := s.getBindDeviceName()
	if err != nil {
		return err
	}

	dev, err := netlink.LinkByName(deviceName)
	if err != nil {
		return err
	}

	bindIP := s.getBindIP()

	routes, err := c.Network().Routes()
	for _, r := range routes {
		_, ipnet, err := net.ParseCIDR(r.CIDR)
		if err != nil {
			logrus.Warnf("error setting up route %s", r.CIDR)
			continue
		}

		gw := net.ParseIP(r.Target)
		if err != nil {
			logrus.Errorf("error parsing target CIDR: %s", err)
			continue
		}

		// check for route
		exists, err := routeExists(dev, ipnet, gw)
		if err != nil {
			logrus.Errorf("error checking route %s: %s", r, err)
			continue
		}
		if exists {
			logrus.Debugf("route %s exists", r.CIDR)
			continue
		}

		if bindIP.Equal(gw) {
			logrus.Debugf("skipping local route %s", r.CIDR)
			continue
		}

		logrus.Debugf("configuring peer route %s via %s", r.CIDR, r.Target)
		route := &netlink.Route{
			LinkIndex: dev.Attrs().Index,
			Dst:       ipnet,
			Gw:        gw,
		}
		if err := netlink.RouteReplace(route); err != nil {
			return err
		}
	}

	return nil
}

func routeExists(link netlink.Link, network *net.IPNet, gateway net.IP) (bool, error) {
	routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
	if err != nil {
		return false, err
	}

	for _, route := range routes {
		if route.Dst == nil {
			continue
		}
		if route.Dst.IP.Equal(network.IP) && route.Gw.Equal(gateway) && route.LinkIndex == link.Attrs().Index {
			return true, nil
		}
	}

	return false, nil
}
