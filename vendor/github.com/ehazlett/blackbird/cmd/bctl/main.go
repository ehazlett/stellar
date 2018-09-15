package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/blackbird"
	"github.com/ehazlett/blackbird/version"
	"github.com/sirupsen/logrus"
)

func main() {
	app := cli.NewApp()
	app.Name = "bctl"
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
			Name:  "addr, a",
			Usage: "blackbird grpc address",
			Value: "unix:///run/blackbird.sock",
		},
	}
	app.Commands = []cli.Command{
		serversCommand,
		reloadCommand,
		configCommand,
	}
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func getClient(ctx *cli.Context) (*blackbird.Client, error) {
	return blackbird.NewClient(ctx.GlobalString("addr"))
}

var reloadCommand = cli.Command{
	Name:   "reload",
	Usage:  "reload proxy service",
	Action: reload,
}

func reload(ctx *cli.Context) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Reload()
}

var configCommand = cli.Command{
	Name:   "config",
	Usage:  "get current proxy config",
	Action: config,
}

func config(ctx *cli.Context) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	data, err := client.Config()
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
