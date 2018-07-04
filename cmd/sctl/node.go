package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
)

var nodeCommand = cli.Command{
	Name:    "node",
	Aliases: []string{"n"},
	Usage:   "interact with nodes",
	Subcommands: []cli.Command{
		nodeContainersCommand,
	},
}

var nodeContainersCommand = cli.Command{
	Name:    "containers",
	Aliases: []string{"c"},
	Usage:   "container management",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}

		containers, err := client.Node().Containers()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "ID\tIMAGE\n")
		for _, c := range containers {
			fmt.Fprintf(w, "%s\t%s\n", c.ID, c.Image)
		}
		w.Flush()

		return nil
	},
}
