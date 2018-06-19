package main

import (
	"net"
	"os"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar/server"
	"github.com/ehazlett/stellar/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := cli.NewApp()
	app.Name = version.Name + " daemon"
	app.Version = version.BuildVersion()
	app.Author = "@ehazlett"
	app.Email = ""
	app.Usage = version.Description
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug logging",
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
		cli.StringSliceFlag{
			Name:  "peer",
			Usage: "one or more peers for agent to join",
			Value: &cli.StringSlice{},
		},
	}
	app.Action = action
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getHostname() string {
	if h, _ := os.Hostname(); h != "" {
		return h
	}

	return "unknown"
}

func action(c *cli.Context) error {
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
	srv, err := server.NewServer(&server.Config{
		AgentConfig:    agentConfig,
		ContainerdAddr: containerdAddr,
		Namespace:      namespace,
		Subnet:         subnet,
		DataDir:        c.String("data-dir"),
	})
	if err != nil {
		return err
	}

	return srv.Run()
}
