package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/stellar"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/version"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	app := cli.NewApp()
	app.Name = "sctl"
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
			Name:   "addr, a",
			Usage:  "stellar daemon address",
			Value:  "127.0.0.1:9000",
			EnvVar: "STELLAR_ADDR",
		},
		cli.StringFlag{
			Name:  "cert, c",
			Usage: "stellar client certificate",
			Value: "",
		},
		cli.StringFlag{
			Name:  "key, k",
			Usage: "stellar client key",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "skip-verify",
			Usage: "skip TLS verification",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}
	app.Commands = []cli.Command{
		appCommand,
		nodeCommand,
		clusterCommand,
		nameserverCommand,
		proxyCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getClient(c *cli.Context) (*client.Client, error) {
	opts := []grpc.DialOption{}
	cert := c.GlobalString("cert")
	key := c.GlobalString("key")
	skipVerification := c.GlobalBool("skip-verify")

	cfg := &stellar.Config{
		TLSClientCertificate:  cert,
		TLSClientKey:          key,
		TLSInsecureSkipVerify: skipVerification,
	}

	opts, err := client.DialOptionsFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return client.NewClient(c.GlobalString("addr"), opts...)
}
