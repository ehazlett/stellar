package main

import (
	"time"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/stellar/server"
)

const (
	localhost = "127.0.0.1"
)

var serverCommand = cli.Command{
	Name:   "server",
	Usage:  "run the stellar daemon",
	Action: serverAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "path to config file (overrides all other options)",
			Value: "",
		},
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
			Name:  "containerd-addr",
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
	},
}

func serverAction(ctx *cli.Context) error {
	cfg, err := getConfig(ctx)
	if err != nil {
		return err
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		return err
	}

	return srv.Run()
}
