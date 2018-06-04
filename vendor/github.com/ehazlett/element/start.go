package element

import (
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

// Start activates the GRPC listener as well as joins the cluster if specified and blocks until a SIGTERM or SIGINT is received
func (a *Agent) Start(signals chan os.Signal) error {
	logrus.Infof("starting agent: grpc=%s:%d bind=%s:%d advertise=%s:%d",
		a.config.AgentAddr,
		a.config.AgentPort,
		a.config.BindAddr,
		a.config.BindPort,
		a.config.AdvertiseAddr,
		a.config.AdvertisePort,
	)
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.AgentAddr, a.config.AgentPort))
	if err != nil {
		return err
	}
	go a.grpcServer.Serve(l)

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

	for {
		select {
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
