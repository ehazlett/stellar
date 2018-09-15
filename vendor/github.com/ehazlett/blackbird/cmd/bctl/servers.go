package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/blackbird"
)

var serversCommand = cli.Command{
	Name:  "servers",
	Usage: "manage servers",
	Subcommands: []cli.Command{
		addServerCommand,
		removeServerCommand,
		listServersCommand,
	},
}

var addServerCommand = cli.Command{
	Name:   "add",
	Usage:  "add server",
	Action: addServer,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Usage: "host name of server",
		},
		cli.StringSliceFlag{
			Name:  "upstream",
			Usage: "server upstreams",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "path",
			Usage: "path prefix for server",
			Value: "/",
		},
		cli.DurationFlag{
			Name:  "timeouts",
			Usage: "server timeouts",
		},
		cli.BoolFlag{
			Name:  "tls",
			Usage: "enable tls",
		},
	},
}

func addServer(ctx *cli.Context) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	host := ctx.String("host")
	upstreams := ctx.StringSlice("upstream")
	opts := []blackbird.AddOpts{
		blackbird.WithPath(ctx.String("path")),
		blackbird.WithUpstreams(upstreams...),
		blackbird.WithTimeouts(ctx.Duration("timeout")),
	}
	if ctx.Bool("tls") {
		opts = append(opts, blackbird.WithTLS)
	}

	if err := client.AddServer(host, opts...); err != nil {
		return err
	}

	return nil
}

var removeServerCommand = cli.Command{
	Name:      "remove",
	Usage:     "remove server",
	Action:    removeServer,
	ArgsUsage: "[HOST]",
}

func removeServer(ctx *cli.Context) error {
	host := ctx.Args().First()
	if host == "" {
		return fmt.Errorf("you must specify a host")
	}

	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.RemoveServer(host); err != nil {
		return err
	}

	return nil
}

var listServersCommand = cli.Command{
	Name:   "list",
	Usage:  "list servers",
	Action: listServers,
}

func listServers(ctx *cli.Context) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	servers, err := client.Servers()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprintf(w, "HOST\tPATH\tTIMEOUTS\tUPSTREAMS\n")
	for _, s := range servers {
		ups := strings.Join(s.Upstreams, ", ")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Host, s.Path, s.Timeouts, ups)

	}
	w.Flush()

	return nil
}
