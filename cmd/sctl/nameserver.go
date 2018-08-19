package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/containerd/typeurl"
	"github.com/ehazlett/stellar/api/types"
)

var nameserverCommand = cli.Command{
	Name:  "nameserver",
	Usage: "manage the cluster nameserver",
	Subcommands: []cli.Command{
		nameserverListRecordsCommand,
		nameserverCreateRecordCommand,
		nameserverDeleteRecordCommand,
	},
}

var nameserverListRecordsCommand = cli.Command{
	Name:  "list",
	Usage: "list nameserver records",
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		records, err := client.Nameserver().List()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "NAME\tTYPE\tVALUE\tOPTIONS\n")
		for _, r := range records {
			opts := ""
			if r.Options != nil {
				v, err := typeurl.UnmarshalAny(r.Options)
				if err != nil {
					return err
				}
				if o, ok := v.(types.NameserverOption); ok {
					opts = o.String()
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.Type, r.Value, opts)
		}
		w.Flush()

		return nil
	},
}

var nameserverCreateRecordCommand = cli.Command{
	Name:  "create",
	Usage: "create nameserver record",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type, t",
			Usage: "resource record type (A, CNAME, TXT, SRV, MX)",
			Value: "A",
		},
		// TODO: handle resource record options
	},
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		t := c.String("type")
		name := c.Args().First()
		value := c.Args().Get(1)

		if name == "" || value == "" {
			return fmt.Errorf("you must enter a name and value")
		}

		if err := client.Nameserver().Create(t, name, value, nil); err != nil {
			return err
		}

		fmt.Printf("added %s=%s (%s)\n", name, value, t)

		return nil
	},
}

var nameserverDeleteRecordCommand = cli.Command{
	Name:  "delete",
	Usage: "delete nameserver record",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type, t",
			Usage: "resource record type (A, CNAME, TXT, SRV, MX)",
			Value: "A",
		},
	},
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		t := c.String("type")
		name := c.Args().First()

		if name == "" {
			return fmt.Errorf("you must enter a name")
		}

		if err := client.Nameserver().Delete(t, name); err != nil {
			return err
		}

		fmt.Printf("removed %s (%s)\n", name, t)

		return nil
	},
}
