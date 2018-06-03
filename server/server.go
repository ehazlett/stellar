package server

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehazlett/element/agent"
	"github.com/ehazlett/element/services"
	clusterservice "github.com/ehazlett/element/services/cluster"
	healthservice "github.com/ehazlett/element/services/health"
	nodeservice "github.com/ehazlett/element/services/node"
	versionservice "github.com/ehazlett/element/services/version"
	"github.com/sirupsen/logrus"
)

var (
	heartbeatInterval = time.Second * 10
)

type Server struct {
	agent *agent.Agent
}

type Config struct {
	AgentConfig    *agent.Config
	ContainerdAddr string
	Namespace      string
}

func NewServer(cfg *Config) (*Server, error) {
	a, err := agent.NewAgent(cfg.AgentConfig)
	if err != nil {
		return nil, err
	}

	// services
	vs, err := versionservice.New(cfg.ContainerdAddr, cfg.Namespace)
	if err != nil {
		return nil, err
	}

	ns, err := nodeservice.New(cfg.ContainerdAddr, cfg.Namespace)
	if err != nil {
		return nil, err
	}

	hs, err := healthservice.New()
	if err != nil {
		return nil, err
	}

	cs, err := clusterservice.New(a, cfg.ContainerdAddr, cfg.Namespace)
	if err != nil {
		return nil, err
	}

	// register with agent
	for _, svc := range []services.Service{vs, ns, hs, cs} {
		if err := a.Register(svc); err != nil {
			return nil, err
		}
		logrus.WithFields(logrus.Fields{
			"id": svc.ID(),
		}).Info("registered service")
	}

	return &Server{
		agent: a,
	}, nil
}

func (s *Server) Run() error {
	signals := make(chan os.Signal, 32)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(heartbeatInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.heartbeat()
			case sig := <-signals:
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT:
					logrus.Debug("shutting down")
					if err := s.agent.Shutdown(); err != nil {
						logrus.Error(err)
					}
				}
			}
		}
	}()

	return s.agent.Start(signals)
}
