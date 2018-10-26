package main

import (
	"fmt"

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
			Usage: "path to config file",
			Value: "",
		},
	},
}

func serverAction(ctx *cli.Context) error {
	p := ctx.String("config")
	if p == "" {
		return fmt.Errorf("config file not specified")
	}

	cfg, err := loadConfigFromFile(p)
	if err != nil {
		return err
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		return err
	}

	return srv.Run()
}
