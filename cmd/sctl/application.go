package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/codegangsta/cli"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	"github.com/pkg/errors"
)

var appCommand = cli.Command{
	Name:  "application",
	Usage: "manage applications",
	Subcommands: []cli.Command{
		appCreateCommand,
	},
}

var appCreateCommand = cli.Command{
	Name:  "create",
	Usage: "create an application",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "path to application config",
			Value: "",
		},
	},
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		configPath := c.String("file")
		if configPath == "" {
			return cli.ShowSubcommandHelp(c)
		}
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			return errors.Wrapf(err, "error accessing config %s", configPath)
		}

		var req *api.CreateRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return errors.Wrap(err, "error loading config")
		}

		if err := client.Application().Create(req); err != nil {
			return err
		}

		return nil
	},
}
