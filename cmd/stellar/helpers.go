package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
)

func defaultConfig() (*stellar.Config, error) {
	agentConfig := &element.Config{
		NodeName:       getHostname(),
		AgentAddr:      localhost,
		AgentPort:      9000,
		ConnectionType: "local",
		BindAddr:       localhost,
		BindPort:       7946,
		AdvertiseAddr:  localhost,
		AdvertisePort:  7946,
		Peers:          []string{},
	}
	_, subnet, err := net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		return nil, err
	}
	return &stellar.Config{
		AgentConfig:              agentConfig,
		ContainerdAddr:           "/run/containerd/containerd.sock",
		Namespace:                "default",
		Subnet:                   subnet,
		DataDir:                  "/var/lib/stellar",
		StateDir:                 "/run/stellar",
		Bridge:                   "stellar0",
		UpstreamDNSAddr:          "1.1.1.1:53",
		ProxyHTTPPort:            80,
		ProxyHTTPSPort:           443,
		ProxyTLSEmail:            "",
		ProxyHealthcheckInterval: time.Second * 5,
		GatewayAddr:              localhost,
		GatewayPort:              9001,
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

func getConfig(ctx *cli.Context) (*stellar.Config, error) {
	if p := ctx.String("config"); p != "" {
		return loadConfigFromFile(p)
	}

	agentAddr := ctx.String("agent-addr")
	bindAddr := ctx.String("bind-addr")
	if agentAddr == "" {
		agentAddr = bindAddr
	}
	gatewayAddr := ctx.String("gateway-addr")
	agentConfig := &element.Config{
		NodeName:       ctx.String("node-name"),
		AgentAddr:      agentAddr,
		AgentPort:      ctx.Int("agent-port"),
		ConnectionType: ctx.String("connection-type"),
		BindAddr:       bindAddr,
		BindPort:       ctx.Int("bind-port"),
		AdvertiseAddr:  ctx.String("advertise-addr"),
		AdvertisePort:  ctx.Int("advertise-port"),
		Peers:          ctx.StringSlice("peer"),
	}
	containerdAddr := ctx.String("containerd-addr")
	namespace := ctx.String("namespace")

	_, subnet, err := net.ParseCIDR(ctx.String("subnet"))
	if err != nil {
		return nil, err
	}
	return &stellar.Config{
		AgentConfig:              agentConfig,
		ContainerdAddr:           containerdAddr,
		Namespace:                namespace,
		Subnet:                   subnet,
		DataDir:                  ctx.String("data-dir"),
		StateDir:                 ctx.String("state-dir"),
		Bridge:                   ctx.String("bridge"),
		UpstreamDNSAddr:          ctx.String("upstream-dns-addr"),
		ProxyHTTPPort:            ctx.Int("proxy-http-port"),
		ProxyHTTPSPort:           ctx.Int("proxy-https-port"),
		ProxyTLSEmail:            ctx.String("proxy-tls-email"),
		ProxyHealthcheckInterval: ctx.Duration("proxy-healthcheck-interval"),
		GatewayAddr:              gatewayAddr,
		GatewayPort:              ctx.Int("gateway-port"),
		CNIBinPaths:              ctx.StringSlice("cni-bin-path"),
	}, nil
}
