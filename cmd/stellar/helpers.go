package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	"github.com/sirupsen/logrus"
)

func getIP(ctx *cli.Context) string {
	ip := "127.0.0.1"
	devName := ctx.String("nic")
	ifaces, err := net.Interfaces()
	if err != nil {
		logrus.Warnf("unable to detect network interfaces")
		return ip
	}
	fmt.Println(ifaces)

	for _, i := range ifaces {
		if devName == "" || i.Name == devName {
			a := getInterfaceIP(i)
			if a != "" {
				return a
			}
		}
	}

	logrus.Warnf("unable to find interface %s", devName)
	return ip
}

func getInterfaceIP(iface net.Interface) string {
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		return ip.To4().String()
	}

	return ""
}

func defaultConfig(ctx *cli.Context) (*stellar.Config, error) {
	ip := getIP(ctx)
	agentConfig := &element.Config{
		ConnectionType:   "local",
		ClusterAddress:   fmt.Sprintf("%s:%d", ip, 7946),
		AdvertiseAddress: fmt.Sprintf("%s:%d", ip, 7946),
		Peers:            []string{},
	}
	_, subnet, err := net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		return nil, err
	}
	return &stellar.Config{
		NodeID:                   getHostname(),
		GRPCAddress:              fmt.Sprintf("%s:%d", ip, 9000),
		AgentConfig:              agentConfig,
		ContainerdAddr:           "/run/containerd/containerd.sock",
		Namespace:                "default",
		Subnet:                   subnet,
		DataDir:                  "/var/lib/stellar",
		StateDir:                 "/run/stellar",
		Bridge:                   "stellar0",
		UpstreamDNSAddr:          "8.8.8.8:53",
		ProxyHTTPPort:            80,
		ProxyHTTPSPort:           443,
		ProxyTLSEmail:            "",
		ProxyHealthcheckInterval: time.Second * 5,
		GatewayAddress:           fmt.Sprintf("%s:%d", ip, 9001),
		EventsAddress:            fmt.Sprintf("%s:%d", ip, 4222),
		EventsHTTPAddress:        fmt.Sprintf("%s:%d", ip, 4322),
		EventsClusterAddress:     fmt.Sprintf("%s:%d", ip, 5222),
		CNIBinPaths:              []string{"/opt/containerd/bin", "/opt/cni/bin"},
	}, nil
}

func loadConfigFromFile(path string) (*stellar.Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg *stellar.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
