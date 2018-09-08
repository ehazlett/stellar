package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	"github.com/pkg/errors"
)

var appInspectTemplate = `Name: {{ .Name }}

Services:{{ range .Services }}
  Name: {{ .Name }}
  Image: {{ .Image }}
  Runtime: {{ .Runtime }}
  Snapshotter: {{ .Snapshotter }}
  Labels:{{ range .Labels }}
    {{.}}{{ end }}
  Endpoints:{{ range .Endpoints }}
    - Service: {{.Service}}
      Protocol: {{.Protocol}}
      Host: {{.Host}}
      Port: {{.Port}}
      TLS:  {{.TLS}}{{ end }}
{{end}}
`

var appCommand = cli.Command{
	Name:    "applications",
	Aliases: []string{"apps"},
	Usage:   "manage applications",
	Subcommands: []cli.Command{
		appListCommand,
		appCreateCommand,
		appDeleteCommand,
		appInspectCommand,
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

var appDeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete an application",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		name := c.Args().First()
		if name == "" {
			return fmt.Errorf("you must specify an application name")
		}

		if err := client.Application().Delete(name); err != nil {
			return err
		}

		fmt.Printf("%s deleted\n", name)

		return nil
	},
}

var appListCommand = cli.Command{
	Name:  "list",
	Usage: "list applications",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		apps, err := client.Application().List()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
		fmt.Fprintf(w, "NAME\tSERVICES\n")
		for _, app := range apps {
			fmt.Fprintf(w, "%s\t%d\n", app.Name, len(app.Services))
		}
		w.Flush()

		return nil
	},
}

var appInspectCommand = cli.Command{
	Name:  "inspect",
	Usage: "view application details",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "format",
			Usage: "format to display output (text, json)",
			Value: "text",
		},
	},
	Action: func(c *cli.Context) error {
		client, err := getClient(c)
		if err != nil {
			return err
		}
		defer client.Close()

		name := c.Args().First()
		if name == "" {
			return fmt.Errorf("you must specify an application name")
		}
		app, err := client.Application().Get(name)
		if err != nil {
			return err
		}

		var f func(app *api.App) error

		format := c.String("format")
		switch format {
		case "json":
			f = appInspectOutputJSON
		case "text":
			f = appInspectOutputText
		default:
			return fmt.Errorf("invalid output format: %s", format)
		}

		return f(app)
	},
}
