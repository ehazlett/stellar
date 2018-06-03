package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	clusterapi "github.com/ehazlett/element/api/services/cluster/v1"
)

var clusterCommand = cli.Command{
	Name:    "cluster",
	Aliases: []string{"n"},
	Usage:   "interact with cluster",
	Subcommands: []cli.Command{
		clusterContainersCommand,
		clusterNodesCommand,
	},
}

var clusterContainersCommand = cli.Command{
	Name:    "containers",
	Aliases: []string{"c"},
	Usage:   "container management",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}

		resp, err := client.ClusterService.Containers(context.Background(), &clusterapi.ContainersRequest{})
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "ID\tIMAGE\n")
		for _, c := range resp.Containers {
			fmt.Fprintf(w, "%s\t%s\n", c.ID, c.Image)
		}
		w.Flush()

		return nil
	},
}

var clusterNodesCommand = cli.Command{
	Name:    "nodes",
	Aliases: []string{"n"},
	Usage:   "cluster node management",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}

		nodes, err := client.Nodes()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "NAME\tADDR\n")
		for _, n := range nodes {
			fmt.Fprintf(w, "%s\t%s\n", n.Name, n.Addr)
		}
		w.Flush()

		return nil
	},
}
