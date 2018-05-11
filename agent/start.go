package agent

import (
	"net"
	"os"
	"syscall"
	"time"

	versionservice "github.com/ehazlett/element/services/version"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func (a *Agent) Start(signals chan os.Signal) error {
	logrus.Infof("starting agent: bind=%s:%d advertise=%s:%d",
		a.config.BindAddr,
		a.config.BindPort,
		a.config.AdvertiseAddr,
		a.config.AdvertisePort,
	)
	grpcServer := grpc.NewServer()
	// TODO: make services into plugins that register
	versionservice.Register(grpcServer)
	l, err := net.Listen("tcp", a.config.AgentAddr)
	if err != nil {
		return err
	}
	go grpcServer.Serve(l)

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			a.heartbeat()
		case s := <-signals:
			switch s {
			case syscall.SIGTERM, syscall.SIGINT:
				if err := a.Shutdown(); err != nil {
					return err
				}

				return nil
			}
		}
	}
}
