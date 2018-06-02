package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/element/agent"
	"github.com/ehazlett/element/server"
	"github.com/ehazlett/element/version"
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
			Name:  "agent-addr, a",
			Usage: "agent grpc addr",
			Value: "127.0.0.1:9000",
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
			Name:  "peer,p",
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
	agentConfig := &agent.Config{
		NodeName:       c.String("node-name"),
		AgentAddr:      c.String("agent-addr"),
		ConnectionType: c.String("connection-type"),
		BindAddr:       c.String("bind-addr"),
		BindPort:       c.Int("bind-port"),
		AdvertiseAddr:  c.String("advertise-addr"),
		AdvertisePort:  c.Int("advertise-port"),
		Peers:          c.StringSlice("peer"),
	}

	containerdAddr := c.String("containerd-addr")
	namespace := c.String("namespace")

	srv, err := server.NewServer(&server.Config{
		AgentConfig:    agentConfig,
		ContainerdAddr: containerdAddr,
		Namespace:      namespace,
	})
	if err != nil {
		return err
	}

	return srv.Run()
}
