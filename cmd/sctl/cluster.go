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
	healthapi "github.com/ehazlett/stellar/api/services/health/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/sirupsen/logrus"
)

var clusterCommand = cli.Command{
	Name:    "cluster",
	Aliases: []string{"c"},
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

		resp, err := client.ClusterService().Containers(context.Background(), &clusterapi.ContainersRequest{})
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "ID\tIMAGE\tRUNTIME\n")
		for _, c := range resp.Containers {
			fmt.Fprintf(w, "%s\t%s\t%s\n", c.ID, c.Image, c.Runtime)
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
		cl, err := getClient(c)
		if err != nil {
			return err
		}

		nodes, err := cl.Cluster().Nodes()
		if err != nil {
			return err
		}

		info := map[*clusterapi.Node]*healthapi.HealthResponse{}
		for _, node := range nodes {
			nc, err := client.NewClient(node.Addr)
			if err != nil {
				return err
			}

			resp, err := nc.HealthService().Health(context.Background(), nil)
			if err != nil {
				return err
			}
			info[node] = resp
			nc.Close()
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "NAME\tADDR\tOS\tUPTIME\tCPUS\tMEMORY (USED)\n")
		for node, health := range info {
			started, err := health.Started()
			if err != nil {
				logrus.Error(err)
				continue
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				node.Name,
				node.Addr,
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
