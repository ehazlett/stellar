package server

import (
	"bytes"
	"fmt"
	"html/template"
	"net"
	"os"
	"path/filepath"

	"github.com/ehazlett/stellar"
	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	resolvConfTemplate = `nameserver {{.}}
ndots 0
`
)

func (s *Server) initNetworking() error {
	logrus.Debug("network init")
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return err
	}
	defer c.Close()

	logrus.Debugf("allocating network subnet for node %s", s.NodeID())
	subnetCIDR, err := c.Network().AllocateSubnet(s.NodeID())
	if err != nil {
		return err
	}
	ip, ipnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return errors.Wrapf(err, "error parsing subnet %q (%s)", subnetCIDR, ip)
	}
	logrus.Debugf("setting up subnet %s", subnetCIDR)

	gw := ip.Mask(ipnet.Mask)
	gw[3]++

	mask, _ := ipnet.Mask.Size()

	logrus.Debugf("setting up local gateway %s", gw.String())
	if err := s.setupGateway(gw, mask); err != nil {
		return err
	}

	if err := s.generateResolvConf(gw.String()); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"subnet":  subnetCIDR,
		"gateway": gw.String(),
	}).Info("network initialized")

	return nil
}

func (s *Server) generateResolvConf(gw string) error {
	resolvConfPath := filepath.Join(s.config.DataDir, "resolv.conf")
	os.Remove(resolvConfPath)
	f, err := os.Create(resolvConfPath)
	if err != nil {
		return err
	}
	defer f.Close()
	t := template.New("resolvconf")
	tmpl, err := t.Parse(resolvConfTemplate)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, gw); err != nil {
		return err
	}
	f.Write(b.Bytes())
	return nil
}

func (s *Server) initContainerNetworking(subnetCIDR string, gw net.IP) error {
	client, err := s.client(s.agent.Self().Address)
	if err != nil {
		return err
	}
	defer client.Close()

	containers, err := client.Node().Containers()
	if err != nil {
		return err
	}

	for _, container := range containers {
		if !networkEnabled(container) {
			logrus.Debugf("container %s does not have stellar networking enabled", container.ID)
			continue
		}
		if !container.Running() {
			logrus.Debugf("container %s not running; skipping networking", container.ID)
			continue
		}
		logrus.WithFields(logrus.Fields{
			"id":     container.ID,
			"labels": container.Labels,
			"pid":    container.Task.Pid,
		}).Debug("configuring container network")
		// allocate IP from network service
		ip, err := client.Network().AllocateIP(container.ID, s.NodeID(), subnetCIDR)
		if err != nil {
			logrus.Errorf("error allocating IP for container %s: %s", container.ID, err)
			continue
		}
		if err := client.Node().SetupContainerNetwork(container.ID, ip.String(), subnetCIDR, gw.String()); err != nil {
			logrus.Errorf("error setting up networking for container %s: %s", container.ID, err)
			continue
		}
	}

	return nil
}

func (s *Server) getBindDeviceName() (string, error) {
	bindHost, _, err := net.SplitHostPort(s.config.AgentConfig.ClusterAddress)
	if err != nil {
		return "", err
	}
	bindIP := net.ParseIP(bindHost)
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

	return "", fmt.Errorf("unable to find interface for bind addr %s", bindHost)
}

func (s *Server) setupGateway(ip net.IP, mask int) error {
	logrus.Debug("setting up gateway")
	c, err := s.client(s.agent.Self().Address)
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

	brLink, err := netlink.LinkByName(s.config.Bridge)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return err
		}

		// setup
		brLink = &netlink.Bridge{
			LinkAttrs: netlink.LinkAttrs{
				Name: s.config.Bridge,
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

	bindIP, err := s.getBindIP()
	if err != nil {
		return err
	}

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

func (s *Server) getBindIP() (net.IP, error) {
	bindHost, _, err := net.SplitHostPort(s.config.AgentConfig.ClusterAddress)
	if err != nil {
		return nil, err
	}
	bindIP := net.ParseIP(bindHost)

	return bindIP, nil
}

func (s *Server) setupRoutes() error {
	c, err := s.client(s.agent.Self().Address)
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

	bindIP, err := s.getBindIP()
	if err != nil {
		return err
	}

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

func networkEnabled(container *runtimeapi.Container) bool {
	_, exists := container.Labels[stellar.StellarNetworkLabel]
	return exists
}
