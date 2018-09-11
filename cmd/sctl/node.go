package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
)

var nodeCommand = cli.Command{
	Name:  "node",
	Usage: "interact with nodes",
	Subcommands: []cli.Command{
		nodeContainersCommand,
	},
}

var nodeContainersCommand = cli.Command{
	Name:  "containers",
	Usage: "container management",
	Subcommands: []cli.Command{
		nodeContainersListCommand,
		nodeDeleteContainerCommand,
	},
}

var nodeContainersListCommand = cli.Command{
	Name:  "list",
	Usage: "list node containers",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		containers, err := client.Node().Containers()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "ID\tIMAGE\tRUNTIME\n")
		for _, c := range containers {
			fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Image, c.Runtime)
		}
		w.Flush()

		return nil
	},
}

var nodeDeleteContainerCommand = cli.Command{
	Name:      "delete",
	Usage:     "delete container",
	ArgsUsage: "<NAME>",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		id := c.Args().First()
		if id == "" {
			return fmt.Errorf("you must specify an id")
		}

		if err := client.Node().DeleteContainer(id); err != nil {
			return err
		}

		fmt.Printf("%s deleted", id)

		return nil
	},
}
