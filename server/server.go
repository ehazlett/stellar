package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehazlett/element"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	clusterservice "github.com/ehazlett/stellar/services/cluster"
	datastoreservice "github.com/ehazlett/stellar/services/datastore"
	healthservice "github.com/ehazlett/stellar/services/health"
	networkservice "github.com/ehazlett/stellar/services/network"
	nodeservice "github.com/ehazlett/stellar/services/node"
	versionservice "github.com/ehazlett/stellar/services/version"
	"github.com/sirupsen/logrus"
)

var (
	dsServerBucketName = "stellar.server"
	reconcileInterval  = time.Second * 10
)

type Server struct {
	agent  *element.Agent
	config *Config
	synced bool
}

type Config struct {
	AgentConfig    *element.Config
	ContainerdAddr string
	Namespace      string
	Subnet         *net.IPNet
	DataDir        string
	Bridge         string
}

func NewServer(cfg *Config) (*Server, error) {
	a, err := element.NewAgent(cfg.AgentConfig)
	if err != nil {
		return nil, err
	}

	// services
	// TODO: implement dependencies for services to alleviate the loading order
	vs, err := versionservice.New(cfg.ContainerdAddr, cfg.Namespace)
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

	ds, err := datastoreservice.New(a, cfg.DataDir)
	if err != nil {
		return nil, err
	}

	netSvc, err := networkservice.New(ds, a, cfg.Subnet)
	if err != nil {
		return nil, err
	}

	ns, err := nodeservice.New(cfg.ContainerdAddr, cfg.Namespace, netSvc)
	if err != nil {
		return nil, err
	}

	// register with agent
	for _, svc := range []services.Service{vs, ns, hs, cs, ds, netSvc} {
		if err := a.Register(svc); err != nil {
			return nil, err
		}
		logrus.WithFields(logrus.Fields{
			"id": svc.ID(),
		}).Info("registered service")
	}

	return &Server{
		agent:  a,
		config: cfg,
	}, nil
}

func (s *Server) NodeName() string {
	return s.config.AgentConfig.NodeName
}

func (s *Server) waitForPeers(timeout time.Duration) error {
	logrus.Debugf("waiting on peers")
	doneChan := make(chan bool)
	go func() {
		for {
			peers, _ := s.agent.Peers()
			if len(peers) > 0 {
				doneChan <- true
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()

	select {
	case <-doneChan:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("no peers detected")
	}
}

func (s *Server) syncDatastore() error {
	// check if joining; if so, clear current datastore and sync from peer
	logrus.Debug("joining cluster; clearing current datastore")
	if err := s.waitForPeers(time.Second * 10); err != nil {
		return err
	}
	// sync entire datastore with peer
	peers, err := s.agent.Peers()
	if err != nil {
		return err
	}
	peer := peers[0]
	c, err := client.NewClient(peer.Addr)
	if err != nil {
		return err
	}
	ctx := context.Background()
	logrus.Debugf("getting backup from peer %s", peer)
	bResp, err := c.DatastoreService().Backup(ctx, &datastoreapi.BackupRequest{})
	if err != nil {
		return err
	}

	lc, err := client.NewClient(fmt.Sprintf("%s:%d", s.config.AgentConfig.AgentAddr, s.config.AgentConfig.AgentPort))
	if err != nil {
		return err
	}
	if _, err := lc.DatastoreService().Restore(ctx, &datastoreapi.RestoreRequest{Data: bResp.Data}); err != nil {
		return err
	}
	logrus.Debugf("restored %d bytes", len(bResp.Data))

	return nil
}

func (s *Server) init() error {
	started := time.Now()
	// initialize networking
	if err := s.initNetworking(); err != nil {
		return err
	}

	logrus.Debugf("initializion duration: %s", time.Since(started))

	return nil
}

func (s *Server) Run() error {
	signals := make(chan os.Signal, 32)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	if err := s.agent.Start(); err != nil {
		return err
	}

	if len(s.config.AgentConfig.Peers) > 0 {
		if err := s.syncDatastore(); err != nil {
			return err
		}
	}

	if err := s.init(); err != nil {
		return err
	}

	ticker := time.NewTicker(reconcileInterval)

	for {
		select {
		case <-ticker.C:
			s.reconcile()
		case sig := <-signals:
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				logrus.Debug("shutting down")
				if err := s.agent.Shutdown(); err != nil {
					return err
				}
				return nil
			}
		}
	}

	return nil
}

func (s *Server) client() (*client.Client, error) {
	localNode, err := s.agent.LocalNode()
	if err != nil {
		return nil, err
	}

	return client.NewClient(localNode.Addr)
}
