package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	cniversion "github.com/containernetworking/cni/pkg/version"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/version"
)

type Net struct {
	Name       string      `json:"name"`
	CNIVersion string      `json:"cniVersion"`
	IPAM       *IPAMConfig `json:"ipam"`
}

type IPAMConfig struct {
	Name     string
	Type     string `json:"type"`
	NodeName string `json:"node_name"`
	PeerAddr string `json:"peer_addr"`
}

func main() {
	ver := version.BuildVersion()
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, cniversion.All, ver)
}

func cmdGet(args *skel.CmdArgs) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

func loadConfig(bytes []byte, envArgs string) (*IPAMConfig, string, error) {
	n := Net{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, "", err
	}

	if n.IPAM == nil {
		return nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	}

	if n.IPAM.NodeName == "" || n.IPAM.PeerAddr == "" {
		return nil, "", fmt.Errorf("IPAM config must have node_name and peer_addr")
	}

	n.IPAM.Name = n.Name

	return n.IPAM, n.CNIVersion, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	cfg, confVersion, err := loadConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	id := args.ContainerID
	ip, subnet, err := allocateIP(id, cfg.NodeName, cfg.PeerAddr)
	if err != nil {
		return err
	}

	gw := ip.Mask(subnet.Mask)
	gw[3]++

	result := &current.Result{}
	result.IPs = []*current.IPConfig{
		{
			Version: "4",
			Address: net.IPNet{IP: ip, Mask: subnet.Mask},
			Gateway: gw,
		},
	}
	//result.DNS = types.DNS{
	//	Nameservers: []string{gw.String()},
	//}
	result.Routes = []*types.Route{
		{
			Dst: net.IPNet{IP: net.IP{0, 0, 0, 0}, Mask: net.IPv4Mask(0, 0, 0, 0)},
			GW:  gw,
		},
	}

	return types.PrintResult(result, confVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	cfg, _, err := loadConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	id := args.ContainerID
	if err := releaseIP(id, cfg.NodeName, cfg.PeerAddr); err != nil {
		return err
	}
	return nil
}

func allocateIP(id, nodeName, peerAddr string) (net.IP, *net.IPNet, error) {
	c, err := client.NewClient(peerAddr)
	if err != nil {
		return nil, nil, err
	}
	defer c.Close()

	subnetCIDR, err := c.Network().GetSubnet(nodeName)
	if err != nil {
		return nil, nil, err
	}
	_, subnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return nil, nil, err
	}

	ip, err := c.Network().AllocateIP(id, nodeName, subnetCIDR)
	if err != nil {
		return nil, nil, err
	}

	return ip, subnet, nil
}

func releaseIP(id, nodeName, peerAddr string) error {
	c, err := client.NewClient(peerAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	ip, err := c.Network().GetIP(id, nodeName)
	if err != nil {
		return err
	}
	if _, err := c.Network().ReleaseIP(id, ip.String(), nodeName); err != nil {
		return err
	}

	return nil
}
