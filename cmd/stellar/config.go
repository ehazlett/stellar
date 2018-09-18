package main

import (
	"encoding/json"
	"os"

	"github.com/codegangsta/cli"
)

var configCommand = cli.Command{
	Name:  "config",
	Usage: "output default config",
	Action: func(ctx *cli.Context) error {
		cfg, err := defaultConfig()
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
