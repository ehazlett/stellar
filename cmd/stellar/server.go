package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/stellarproject/element"
	"github.com/ehazlett/stellar"
	"github.com/ehazlett/stellar/server"
	"github.com/ehazlett/stellar/services"
	applicationservice "github.com/ehazlett/stellar/services/application"
	clusterservice "github.com/ehazlett/stellar/services/cluster"
	datastoreservice "github.com/ehazlett/stellar/services/datastore"
	eventsservice "github.com/ehazlett/stellar/services/events"
	gatewayservice "github.com/ehazlett/stellar/services/gateway"
	healthservice "github.com/ehazlett/stellar/services/health"
	nameserverservice "github.com/ehazlett/stellar/services/nameserver"
	networkservice "github.com/ehazlett/stellar/services/network"
	nodeservice "github.com/ehazlett/stellar/services/node"
	proxyservice "github.com/ehazlett/stellar/services/proxy"
	schedulerservice "github.com/ehazlett/stellar/services/scheduler"
	versionservice "github.com/ehazlett/stellar/services/version"
	"github.com/sirupsen/logrus"
)

const (
	localhost = "127.0.0.1"
)

var serverCommand = cli.Command{
	Name:   "server",
	Usage:  "run the stellar daemon",
	Action: serverAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "path to config file",
			Value: "",
		},
	},
}

func serverAction(ctx *cli.Context) error {
	p := ctx.String("config")
	if p == "" {
		return fmt.Errorf("config file not specified")
	}

	cfg, err := loadConfigFromFile(p)
	if err != nil {
		return err
	}

	// services
	svcs := []func(cfg *stellar.Config, agent *element.Agent) (services.Service, error){
		versionservice.New,
		healthservice.New,
		clusterservice.New,
		datastoreservice.New,
		networkservice.New,
		gatewayservice.New,
		nodeservice.New,
		applicationservice.New,
		nameserverservice.New,
		proxyservice.New,
		eventsservice.New,
		schedulerservice.New,
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		return err
	}

	// register services
	if err := srv.Register(svcs); err != nil {
		return err
	}

	if err := srv.Run(); err != nil {
		return err
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	doneCh := make(chan bool, 1)

	go func() {
		for {
			select {
			case sig := <-signals:
				switch sig {
				case syscall.SIGUSR1:
					logrus.Debug("generating debug profile")
					profilePath, err := srv.GenerateProfile()
					if err != nil {
						logrus.Error(err)
						continue
					}
					logrus.WithFields(logrus.Fields{
						"profile": profilePath,
					}).Info("generated memory profile")
				case syscall.SIGTERM, syscall.SIGINT:
					logrus.Info("shutting down")
					if err := srv.Stop(); err != nil {
						logrus.Error(err)
					}
					doneCh <- true
				default:
					logrus.Warnf("unhandled signal %s", sig)
				}
			}
		}
	}()

	<-doneCh

	return nil
}
