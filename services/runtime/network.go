package runtime

import (
	"context"
	"fmt"
	"net"
	"runtime"

	api "github.com/ehazlett/stellar/api/services/runtime/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	vethPeerNameContainer = "eth0"
)

var (
	empty = &ptypes.Empty{}
)

func (s *service) SetupContainerNetwork(ctx context.Context, req *api.ContainerNetworkRequest) (*ptypes.Empty, error) {
	_, ipnet, err := net.ParseCIDR(req.Network)
	if err != nil {
		return empty, err
	}

	gw := net.ParseIP(req.Gateway)

	ip := net.ParseIP(req.IP)
	ipnet.IP = ip

	veth, err := s.createVeth(ctx, req.ID, ipnet, gw)
	if err != nil {
		return empty, err
	}

	logrus.WithFields(logrus.Fields{
		"name":    veth.Attrs().Name,
		"id":      req.ID,
		"ip":      req.IP,
		"network": req.Network,
		"gateway": req.Gateway,
	}).Debug("configured veth")

	return empty, nil
}

func (s *service) createVeth(ctx context.Context, id string, ipnet *net.IPNet, gw net.IP) (netlink.Link, error) {
	resp, err := s.Container(ctx, &api.ContainerRequest{
		ID: id,
	})
	if err != nil {
		return nil, err
	}
	container := resp.Container
	pid := container.Task.Pid

	br, err := netlink.LinkByName(s.bridge)
	if err != nil {
		return nil, errors.Wrapf(err, "%s not found", s.bridge)
	}
	attrs := netlink.NewLinkAttrs()
	// a deterministic name is generated as ethernet device names have a limitation on length
	devName := getName(container.ID)
	vethName := fmt.Sprintf("v%s", devName)
	vethPeerName := fmt.Sprintf("p%s", devName)
	// check for existing
	v, _ := netlink.LinkByName(vethName)
	if v != nil {
		logrus.Debugf("veth for %s exists", id)
		return v, nil
	}

	attrs.Name = vethName
	attrs.MasterIndex = br.Attrs().Index
	veth := &netlink.Veth{
		LinkAttrs: attrs,
		PeerName:  vethPeerName,
	}
	if err := netlink.LinkAdd(veth); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(veth); err != nil {
		return nil, err
	}

	vethPeer, err := netlink.LinkByName(vethPeerName)
	if err != nil {
		return nil, err
	}
	if err := configureVeth(vethPeer, ipnet, gw, int(pid)); err != nil {
		return nil, err
	}

	return veth, nil
}

func configureVeth(link netlink.Link, ipnet *net.IPNet, gateway net.IP, pid int) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	oNS, err := netns.Get()
	if err != nil {
		return err
	}
	defer oNS.Close()

	cNS, err := netns.GetFromPid(pid)
	if err != nil {
		return err
	}
	defer cNS.Close()

	if err := netlink.LinkSetNsPid(link, pid); err != nil {
		return err
	}

	// change to ns
	if err := netns.Set(cNS); err != nil {
		return err
	}
	defer netns.Set(oNS)

	// add addr
	addr := &netlink.Addr{
		IPNet: ipnet,
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return err
	}

	// rename veth in container
	if err := netlink.LinkSetName(link, vethPeerNameContainer); err != nil {
		return err
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	// add route
	_, defaultNet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		return err
	}
	if err := netlink.RouteReplace(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       defaultNet,
		Gw:        gateway,
	}); err != nil {
		return err
	}

	return nil
}
