package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	applicationservice "github.com/ehazlett/stellar/services/application"
	clusterservice "github.com/ehazlett/stellar/services/cluster"
	datastoreservice "github.com/ehazlett/stellar/services/datastore"
	healthservice "github.com/ehazlett/stellar/services/health"
	nameserverservice "github.com/ehazlett/stellar/services/nameserver"
	networkservice "github.com/ehazlett/stellar/services/network"
	nodeservice "github.com/ehazlett/stellar/services/node"
	proxyservice "github.com/ehazlett/stellar/services/proxy"
	versionservice "github.com/ehazlett/stellar/services/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	dsServerBucketName = "stellar.server"
	// TODO: make configurable
	reconcileInterval = time.Second * 10
	// TODO: make configurable
	datastoreSyncInterval = time.Second * 300
)

type Server struct {
	agent       *element.Agent
	config      *stellar.Config
	synced      bool
	nodeEventCh chan *element.NodeEvent
	services    []services.Service
}

func NewServer(cfg *stellar.Config) (*Server, error) {
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

	hs, err := healthservice.New(a)
	if err != nil {
		return nil, err
	}

	cs, err := clusterservice.New(cfg, a)
	if err != nil {
		return nil, err
	}

	ds, err := datastoreservice.New(cfg, a)
	if err != nil {
		return nil, err
	}

	netSvc, err := networkservice.New(cfg, a, ds)
	if err != nil {
		return nil, err
	}

	nodeSvc, err := nodeservice.New(cfg, a)
	if err != nil {
		return nil, err
	}

	appSvc, err := applicationservice.New(cfg, a)
	if err != nil {
		return nil, err
	}

	nsSvc, err := nameserverservice.New(cfg, a)
	if err != nil {
		return nil, err
	}
	proxySvc, err := proxyservice.New(cfg, a)
	if err != nil {
		return nil, err
	}
	// register with agent
	svcs := []services.Service{vs, nodeSvc, hs, cs, ds, netSvc, appSvc, nsSvc, proxySvc}
	for _, svc := range svcs {
		if err := a.Register(svc); err != nil {
			return nil, err
		}
		logrus.WithFields(logrus.Fields{
			"id": svc.ID(),
		}).Info("registered service")
	}

	nodeEventCh := make(chan *element.NodeEvent)
	a.Subscribe(nodeEventCh)

	srv := &Server{
		agent:       a,
		config:      cfg,
		nodeEventCh: nodeEventCh,
		services:    svcs,
	}

	go srv.eventHandler(nodeEventCh)

	return srv, nil
}

func (s *Server) NodeName() string {
	return s.config.AgentConfig.NodeName
}

func (s *Server) eventHandler(ch chan *element.NodeEvent) {
	for {
		evt := <-ch
		logrus.Debugf("event: %+v", evt)
	}
}

func (s *Server) waitForPeers() error {
	logrus.Infof("waiting on initial cluster sync (could take up to %s)", s.agent.SyncInterval()*2)

	doneChan := make(chan bool)
	errChan := make(chan error)

	localNode, err := s.agent.LocalNode()
	if err != nil {
		return err
	}

	go func() {
		for {
			peers, err := s.agent.Peers()
			if err != nil {
				errChan <- err
			}

			if len(peers) > 0 {
				peer := peers[0]
				ac, err := client.NewClient(peer.Addr)
				if err != nil {
					errChan <- err
					break
				}
				clusterNodes, err := ac.Cluster().Nodes()
				if err != nil {
					errChan <- err
					break
				}
				ac.Close()

				lc, err := client.NewClient(localNode.Addr)
				if err != nil {
					errChan <- err
					break
				}

				localClusterNodes, err := lc.Cluster().Nodes()
				if err != nil {
					errChan <- err
					break
				}

				if len(localClusterNodes) == len(clusterNodes) {
					logrus.Debugf("discovered %d cluster nodes (%s); cluster membership in sync", len(localClusterNodes), localClusterNodes)
					doneChan <- true
					return
				}
			}

			time.Sleep(time.Millisecond * 500)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-doneChan:
		return nil
	}

	return nil
}

func (s *Server) syncDatastore() error {
	// check if joining; if so, clear current datastore and sync from peer
	logrus.Debug("joining cluster; clearing current datastore")
	if err := s.waitForPeers(); err != nil {
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

	// TODO: handle network init with CNI
	//// initialize networking
	//if err := s.initNetworking(); err != nil {
	//	return err
	//}

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

	errCh := make(chan error)
	tickerReconcile := time.NewTicker(reconcileInterval)
	tickerDatastoreSync := time.NewTicker(datastoreSyncInterval)

	go func() {
		for range tickerReconcile.C {
			if err := s.reconcile(); err != nil {
				errCh <- err
			}
		}
	}()

	go func() {
		for range tickerDatastoreSync.C {
			if err := s.syncPeerDatastores(); err != nil {
				errCh <- err
			}
		}
	}()

	// start services
	for _, svc := range s.services {
		if err := svc.Start(); err != nil {
			errCh <- err
		}
	}

	for {
		select {
		case err := <-errCh:
			logrus.Error(err)
		case sig := <-signals:
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				logrus.Info("shutting down")

				tickerReconcile.Stop()
				tickerDatastoreSync.Stop()

				// shutdown server
				if err := s.shutdown(); err != nil {
					logrus.Error(err)
				}

				// shutdown element agent
				if err := s.agent.Shutdown(); err != nil {
					return err
				}

				return nil
			}
		}
	}

	return nil
}

func (s *Server) syncPeerDatastores() error {
	localNode, err := s.agent.LocalNode()
	if err != nil {
		return err
	}

	lc, err := client.NewClient(localNode.Addr)
	if err != nil {
		return err
	}
	defer lc.Close()

	peers, err := s.agent.Peers()
	if err != nil {
		return err
	}

	for _, peer := range peers {
		logrus.WithFields(logrus.Fields{
			"peer": peer.Name,
			"addr": peer.Addr,
		}).Debug("syncing peer datastore")
		c, err := client.NewClient(peer.Addr)
		if err != nil {
			return err
		}
		defer c.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		stream, err := c.DatastoreService().Sync(ctx, &datastoreapi.SyncRequest{})
		if err != nil {
			return err
		}

		count := 0
		for {
			op, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				return errors.Wrap(err, "error syncing datastore")
			}

			// TODO: handle operation
			if _, err := lc.DatastoreService().Set(ctx, &datastoreapi.SetRequest{
				Bucket: op.Bucket,
				Key:    op.Key,
				Value:  op.Value,
			}); err != nil {
				return errors.Wrapf(err, "error syncing key %s", op.Key)
			}
			count++
		}

		logrus.Debugf("synchronized %d operations from peer %s", count, peer.Name)

	}

	return nil
}

func (s *Server) shutdown() error {
	// signal datastore to shutdown
	localNode, err := s.agent.LocalNode()
	if err != nil {
		return err
	}
	lc, err := client.NewClient(localNode.Addr)
	if err != nil {
		return err
	}
	defer lc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if _, err := lc.DatastoreService().Shutdown(ctx, &datastoreapi.ShutdownRequest{}); err != nil {
		return err
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
