package agent

import (
	"net"
	"os"
	"syscall"
	"time"

	healthservice "github.com/ehazlett/element/services/health"
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
	versionservice.Register(grpcServer, a.config.ContainerdAddr, a.config.Namespace)
	healthservice.Register(grpcServer, a.config.ContainerdAddr, a.config.Namespace)

	l, err := net.Listen("tcp", a.config.AgentAddr)
	if err != nil {
		return err
	}
	go grpcServer.Serve(l)

	// start node metadata updater
	go func() {
		for {
			<-a.peerUpdateChan
			if err := a.members.UpdateNode(nodeUpdateTimeout); err != nil {
				logrus.Errorf("error updating node metadata: %s", err)
			}
		}
	}()

	if len(a.config.Peers) > 0 {
		logrus.Debugf("joining peers: %v", a.config.Peers)
		n, err := a.members.Join(a.config.Peers)
		if err != nil {
			return err
		}

		logrus.Infof("joined %d peer(s)", n)
	}

	ticker := time.NewTicker(nodeHeartbeatInterval)
	for {
		select {
		case <-ticker.C:
			a.heartbeat()
		case s := <-signals:
			switch s {
			case syscall.SIGTERM, syscall.SIGINT:
				logrus.Debug("shutting down")
				if err := a.Shutdown(); err != nil {
					return err
				}

				return nil
			}
		}
	}
}
