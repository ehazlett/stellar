package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/blackbird"
	"github.com/ehazlett/blackbird/ds/memory"
	"github.com/ehazlett/blackbird/server"
	"github.com/ehazlett/blackbird/version"
	"github.com/sirupsen/logrus"
)

func main() {
	app := cli.NewApp()
	app.Name = version.Name
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
			Name:  "grpc-addr, g",
			Usage: "grpc listen address",
			Value: "unix:///run/blackbird.sock",
		},
		cli.IntFlag{
			Name:  "http-port",
			Usage: "http port",
			Value: 80,
		},
		cli.IntFlag{
			Name:  "https-port",
			Usage: "https port",
			Value: 443,
		},
	}
	app.Action = start
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func start(ctx *cli.Context) error {
	cfg := &blackbird.Config{
		GRPCAddr:  ctx.String("grpc-addr"),
		HTTPPort:  ctx.Int("http-port"),
		HTTPSPort: ctx.Int("https-port"),
		Debug:     ctx.Bool("debug"),
	}

	memDs := memory.NewMemory()

	srv, err := server.NewServer(cfg, memDs)
	if err != nil {
		return err
	}

	if err := srv.Run(); err != nil {
		return err
	}

	// wait
	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		logrus.Info("shutting down")
		if err := srv.Stop(); err != nil {
			logrus.Error(err)
		}
		done <- true
	}()

	<-done
	return nil
}
