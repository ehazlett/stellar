package main

import (
	"encoding/json"
	"os"

	"github.com/codegangsta/cli"
)

var configCommand = cli.Command{
	Name:  "config",
	Usage: "output default config",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "nic, n",
			Usage: "network interface to use for detecting IP (default: first non-local)",
		},
	},
	Action: func(ctx *cli.Context) error {
		cfg, err := defaultConfig(ctx)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		if err := enc.Encode(cfg); err != nil {
			return err
		}
		return nil
	},
}
