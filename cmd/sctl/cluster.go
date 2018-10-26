package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	humanize "github.com/dustin/go-humanize"
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	"github.com/sirupsen/logrus"
)

var clusterCommand = cli.Command{
	Name:  "cluster",
	Usage: "interact with cluster",
	Subcommands: []cli.Command{
		clusterContainersCommand,
		clusterNodesCommand,
	},
}

var clusterContainersCommand = cli.Command{
	Name:  "containers",
	Usage: "container management",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		resp, err := client.ClusterService().Containers(context.Background(), &clusterapi.ContainersRequest{})
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "ID\tIMAGE\tRUNTIME\tNODE\n")
		for _, c := range resp.Containers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Container.ID, c.Container.Image, c.Container.Runtime, c.Node.ID)
		}
		w.Flush()

		return nil
	},
}

var clusterNodesCommand = cli.Command{
	Name:  "nodes",
	Usage: "cluster node management",
	Action: func(c *cli.Context) error {
		cl, err := getClient(c)
		if err != nil {
			return err
		}
		defer cl.Close()

		nodes, err := cl.Cluster().Health()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "NAME\tADDR\tOS\tUPTIME\tCPUS\tMEMORY\n")
		for _, nodeHealth := range nodes {
			node := nodeHealth.Node
			health := nodeHealth.Health
			started, err := health.Started()
			if err != nil {
				logrus.Error(err)
				continue
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				node.ID,
				node.Address,
				health.OSName+" ("+health.OSVersion+")",
				humanize.RelTime(started, time.Now(), "", ""),
				health.Cpus,
				fmt.Sprintf("%s / %s", humanize.Bytes(uint64(health.MemoryUsed)), humanize.Bytes(uint64(health.MemoryTotal))),
			)
		}
		w.Flush()

		return nil
	},
}
