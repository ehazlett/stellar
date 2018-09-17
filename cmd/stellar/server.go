package main

import (
	"net"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	"github.com/ehazlett/stellar/server"
)

var serverCommand = cli.Command{
	Name:   "server",
	Usage:  "run the stellar daemon",
	Action: serverAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "node-name, n",
			Usage: "agent node name",
			Value: getHostname(),
		},
		cli.StringFlag{
			Name:  "data-dir, d",
			Usage: "stellar agent data dir",
			Value: "/var/lib/stellar",
		},
		cli.StringFlag{
			Name:  "state-dir, s",
			Usage: "stellar agent state dir",
			Value: "/run/stellar",
		},
		cli.StringFlag{
			Name:  "bridge",
			Usage: "bridge name for networking",
			Value: "stellar0",
		},
		cli.StringFlag{
			Name:  "agent-addr, a",
			Usage: "agent grpc addr",
		},
		cli.IntFlag{
			Name:  "agent-port, p",
			Usage: "agent grpc port",
			Value: 9000,
		},
		cli.StringFlag{
			Name:  "containerd-addr, c",
			Usage: "containerd socket address",
			Value: "/run/containerd/containerd.sock",
		},
		cli.StringFlag{
			Name:  "namespace",
			Usage: "containerd namespace to manage",
			Value: "default",
		},
		cli.StringFlag{
			Name:  "subnet",
			Usage: "network subnet to use for containers",
			Value: "172.16.0.0/12",
		},
		cli.StringFlag{
			Name:  "connection-type, t",
			Usage: "connection type (lan, wan, local)",
			Value: "local",
		},
		cli.StringFlag{
			Name:  "bind-addr",
			Usage: "bind address",
			Value: "127.0.0.1",
		},
		cli.IntFlag{
			Name:  "bind-port",
			Usage: "bind port",
			Value: 7946,
		},
		cli.StringFlag{
			Name:  "advertise-addr",
			Usage: "advertise address",
			Value: "127.0.0.1",
		},
		cli.IntFlag{
			Name:  "advertise-port",
			Usage: "advertise port",
			Value: 7946,
		},
		cli.IntFlag{
			Name:  "proxy-http-port",
			Usage: "https port for the proxy service",
			Value: 80,
		},
		cli.IntFlag{
			Name:  "proxy-https-port",
			Usage: "https port for the proxy service",
			Value: 443,
		},
		cli.StringFlag{
			Name:  "proxy-tls-email",
			Usage: "email for the auto TLS proxy service",
			Value: "",
		},
		cli.DurationFlag{
			Name:  "proxy-healthcheck-interval",
			Usage: "proxy backend healthcheck interval",
			Value: time.Second * 5,
		},
		cli.StringFlag{
			Name:  "upstream-dns-addr",
			Usage: "address to forward non-cluster dns lookups",
			Value: "1.1.1.1:53",
		},
		cli.StringSliceFlag{
			Name:  "peer",
			Usage: "one or more peers for agent to join",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "cniBinPath",
			Usage: "path to look for cni binaries",
			Value: "",
		},
	},
}

func serverAction(c *cli.Context) error {
	agentAddr := c.String("agent-addr")
	bindAddr := c.String("bind-addr")
	if agentAddr == "" {
		agentAddr = bindAddr
	}
	agentConfig := &element.Config{
		NodeName:       c.String("node-name"),
		AgentAddr:      agentAddr,
		AgentPort:      c.Int("agent-port"),
		ConnectionType: c.String("connection-type"),
		BindAddr:       bindAddr,
		BindPort:       c.Int("bind-port"),
		AdvertiseAddr:  c.String("advertise-addr"),
		AdvertisePort:  c.Int("advertise-port"),
		Peers:          c.StringSlice("peer"),
	}

	containerdAddr := c.String("containerd-addr")
	namespace := c.String("namespace")

	_, subnet, err := net.ParseCIDR(c.String("subnet"))
	if err != nil {
		return err
	}
	srv, err := server.NewServer(&stellar.Config{
		AgentConfig:              agentConfig,
		ContainerdAddr:           containerdAddr,
		Namespace:                namespace,
		Subnet:                   subnet,
		DataDir:                  c.String("data-dir"),
		StateDir:                 c.String("state-dir"),
		Bridge:                   c.String("bridge"),
		UpstreamDNSAddr:          c.String("upstream-dns-addr"),
		ProxyHTTPPort:            c.Int("proxy-http-port"),
		ProxyHTTPSPort:           c.Int("proxy-https-port"),
		ProxyTLSEmail:            c.String("proxy-tls-email"),
		ProxyHealthcheckInterval: c.Duration("proxy-healthcheck-interval"),
		CNIBinPath:               c.String("cniBinPath"),
	})
	if err != nil {
		return err
	}

	return srv.Run()
}
