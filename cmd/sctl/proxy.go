package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	ptypes "github.com/gogo/protobuf/types"
)

var proxyCommand = cli.Command{
	Name:  "proxy",
	Usage: "manage the cluster proxy",
	Subcommands: []cli.Command{
		proxyListBackendsCommand,
		proxyReloadCommand,
	},
}

var proxyReloadCommand = cli.Command{
	Name:  "reload",
	Usage: "reload proxy service",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		if err := client.Proxy().Reload(); err != nil {
			return err
		}

		return nil
	},
}
var proxyListBackendsCommand = cli.Command{
	Name:  "list",
	Usage: "list backends",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		backends, err := client.Proxy().Backends()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "HOST\tUPSTREAMS\n")
		for _, b := range backends {
			upstreams := []string{}
			for _, up := range b.Upstreams {
				latency, err := ptypes.DurationFromProto(up.Latency)
				if err != nil {
					return err
				}

				status := up.Status
				if latency > time.Millisecond*0 {
					status = fmt.Sprintf("%s: %s", up.Status, latency)
				}
				upstreams = append(upstreams, fmt.Sprintf("%s (%s)", up.Address, status))
			}
			fmt.Fprintf(w, "%s\t%s\n", b.Host, strings.Join(upstreams, ", "))
		}
		w.Flush()

		return nil
	},
}
