package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/element/agent"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "start element agent",
	Flags: []cli.Flag{
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
	},
	Action: runAction,
}

func getHostname() string {
	if h, _ := os.Hostname(); h != "" {
		return h
	}

	return "unknown"
}

func runAction(c *cli.Context) error {
	cfg := &agent.Config{
		NodeName:       c.String("node-name"),
		AgentAddr:      c.String("agent-addr"),
		ConnectionType: c.String("connection-type"),
		BindAddr:       c.String("bind-addr"),
		BindPort:       c.Int("bind-port"),
		AdvertiseAddr:  c.String("advertise-addr"),
		AdvertisePort:  c.Int("advertise-port"),
	}

	a, err := agent.NewAgent(cfg)
	if err != nil {
		return err
	}

	peers := c.StringSlice("peer")
	if len(peers) > 0 {
		if err := a.Join(peers); err != nil {
			return err
		}
	}

	signals := make(chan os.Signal, 32)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	return a.Start(signals)
}
