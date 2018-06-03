package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/element"
	"github.com/ehazlett/element/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := cli.NewApp()
	app.Name = version.Name + " cli"
	app.Version = version.BuildVersion()
	app.Author = "@ehazlett"
	app.Email = ""
	app.Usage = version.Description
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug logging",
		},
		cli.StringFlag{
			Name:  "addr",
			Usage: "element daemon address",
			Value: "127.0.0.1:9000",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}
	app.Commands = []cli.Command{
		nodeCommand,
		clusterCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getClient(c *cli.Context) (*element.Client, error) {
	return element.NewClient(c.GlobalString("addr"))
}
