package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
)

func defaultConfig() (*stellar.Config, error) {
	agentConfig := &element.Config{
		ConnectionType:   "local",
		ClusterAddress:   fmt.Sprintf("%s:%d", localhost, 7946),
		AdvertiseAddress: fmt.Sprintf("%s:%d", localhost, 7946),
		Peers:            []string{},
	}
	_, subnet, err := net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		return nil, err
	}
	return &stellar.Config{
		NodeID:                   getHostname(),
		GRPCAddress:              fmt.Sprintf("%s:%d", localhost, 9000),
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
		GatewayAddress:           fmt.Sprintf("%s:%d", localhost, 9001),
		EventsAddress:            fmt.Sprintf("%s:%d", localhost, 4222),
		EventsHTTPAddress:        fmt.Sprintf("%s:%d", localhost, 4322),
		EventsClusterAddress:     fmt.Sprintf("%s:%d", localhost, 5222),
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
