package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	// ErrServiceRegistered is returned if an existing service is already registered for the specified type
	ErrServiceRegistered = errors.New("service is already registered for the specified type")
	// ErrServiceNotRegistered is returned if a service is not registered that is required by another service
	ErrServiceNotRegistered = errors.New("service is not registered for the specified type")

	dsServerBucketName = "stellar.server"
	// TODO: make configurable
	reconcileInterval     = time.Second * 10
	datastoreSyncInterval = time.Second * 300
	// timeouts for services to start and stop
	serviceStartTimeout = time.Second * 5
	serviceStopTimeout  = time.Second * 5

	// local server state db
	localDBFilename       = "local.db"
	dsLocalPeerBucketName = "peers"
)

type Server struct {
	agent               *element.Agent
	grpcServer          *grpc.Server
	config              *stellar.Config
	synced              bool
	nodeEventCh         chan *element.NodeEvent
	mu                  *sync.Mutex
	db                  *localDB
	tickerReconcile     *time.Ticker
	tickerDatastoreSync *time.Ticker
	services            map[services.Type]services.Service
	errCh               chan error
}

func NewServer(cfg *stellar.Config) (*Server, error) {
	db, err := NewLocalDB(filepath.Join(cfg.DataDir, localDBFilename))
	if err != nil {
		return nil, err
	}

	logrus.WithField("seedPeers", cfg.AgentConfig.Peers).Debug("seed peers")
	// check if there are cached peers
	peers, err := getPeersFromCache(db, cfg.AgentConfig.Peers)
	if err != nil {
		return nil, err
	}

	// override agent config peers with available
	cfg.AgentConfig.Peers = peers
	logrus.WithFields(logrus.Fields{
		"peers":      peers,
		"seed_peers": cfg.AgentConfig.Peers,
	}).Debug("cluster peers")

	a, err := element.NewAgent(&element.Peer{
		ID:      cfg.NodeID,
		Address: cfg.GRPCAddress,
	}, cfg.AgentConfig)
	if err != nil {
		return nil, err
	}

	grpcOpts := []grpc.ServerOption{}
	if cfg.TLSServerCertificate != "" && cfg.TLSServerKey != "" {
		logrus.WithFields(logrus.Fields{
			"cert": cfg.TLSServerCertificate,
			"key":  cfg.TLSServerKey,
		}).Debug("configuring TLS for GRPC")
		cert, err := tls.LoadX509KeyPair(cfg.TLSServerCertificate, cfg.TLSServerKey)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewTLS(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			ClientAuth:         tls.RequestClientCert,
			InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
		})
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(grpcOpts...)

	nodeEventCh := a.Subscribe()

	srv := &Server{
		agent:       a,
		grpcServer:  grpcServer,
		config:      cfg,
		db:          db,
		services:    make(map[services.Type]services.Service),
		mu:          &sync.Mutex{},
		nodeEventCh: nodeEventCh,
		errCh:       make(chan error),
	}

	go srv.eventHandler(nodeEventCh)

	return srv, nil
}

func (s *Server) Register(svcs []func(*stellar.Config, *element.Agent) (services.Service, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// register services from caller
	for _, svc := range svcs {
		i, err := svc(s.config, s.agent)
		if err != nil {
			return err
		}
		if err := i.Register(s.grpcServer); err != nil {
			return err
		}
		// check for existing service
		if _, exists := s.services[i.Type()]; exists {
			return errors.Wrap(ErrServiceRegistered, string(i.Type()))
		}
		logrus.WithFields(logrus.Fields{
			"id":   i.ID(),
			"type": i.Type(),
		}).Info("registered service")
		s.services[i.Type()] = i
	}

	return nil
}

func (s *Server) NodeID() string {
	return s.config.NodeID
}

func (s *Server) eventHandler(ch chan *element.NodeEvent) {
	for {
		evt := <-ch
		logrus.Debugf("event: %+v", evt)
		switch evt.Type {
		case element.NodeUpdate:
			s.eventHandlerNodeUpdate(evt)
		}
	}
}

func (s *Server) eventHandlerNodeUpdate(evt *element.NodeEvent) {
	node := evt.Node
	peer := &Peer{
		ID:      node.Name,
		Address: node.Address(),
	}

	if err := s.cachePeer(peer); err != nil {
		logrus.WithError(err).Error("error caching cluster peer")
	}
}

func (s *Server) waitForPeers() error {
	logrus.Infof("waiting on initial cluster sync (could take up to %s)", s.agent.SyncInterval()*2)

	doneChan := make(chan bool)
	errChan := make(chan error)

	go func() {
		for {
			peers, err := s.agent.Peers()
			if err != nil {
				errChan <- err
			}

			if len(peers) > 0 {
				peer := peers[0]
				ac, err := s.client(peer.Address)
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

				// check if list of peers plus self equal the reported number
				// of cluster nodes from the peer
				if len(peers)+1 == len(clusterNodes) {
					logrus.Debugf("discovered %d cluster nodes (%s); cluster membership in sync", len(peers), peers)
					doneChan <- true
					return
				}
				logrus.Debugf("waiting on peers; detected %d of %d nodes", len(peers)+1, len(clusterNodes))
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
	// sync entire datastore with peer
	peers, err := s.agent.Peers()
	if err != nil {
		return err
	}
	peer := peers[0]
	c, err := s.client(peer.Address)
	if err != nil {
		return err
	}

	ctx := context.Background()
	logrus.Debugf("getting backup from peer %s", peer)
	bResp, err := c.DatastoreService().Backup(ctx, &datastoreapi.BackupRequest{})
	if err != nil {
		return err
	}

	lc, err := s.client(s.agent.Self().Address)
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
	logrus.Debug("initializing server")
	started := time.Now()

	// initialize local networking
	if err := s.initNetworking(); err != nil {
		return err
	}

	logrus.Debugf("initializion duration: %s", time.Since(started))

	return nil
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", s.config.GRPCAddress)
	if err != nil {
		return err
	}
	logrus.Debug("starting server agent")
	if err := s.agent.Start(); err != nil {
		return err
	}

	isPeer := len(s.config.AgentConfig.Peers) > 0

	if isPeer {
		if err := s.waitForPeers(); err != nil {
			return err
		}
	}

	logrus.WithField("addr", s.config.GRPCAddress).Debug("starting grpc server")
	go s.grpcServer.Serve(l)

	if isPeer {
		// check if joining; if so, clear current datastore and sync from peer
		logrus.Debug("joining cluster; clearing current datastore")
		if err := s.syncDatastore(); err != nil {
			return err
		}
	}

	if err := s.init(); err != nil {
		return err
	}

	go func() {
		for {
			err := <-s.errCh
			logrus.Error(err)
		}
	}()

	s.tickerReconcile = time.NewTicker(reconcileInterval)
	s.tickerDatastoreSync = time.NewTicker(datastoreSyncInterval)

	go func() {
		for range s.tickerReconcile.C {
			if err := s.reconcile(); err != nil {
				s.errCh <- err
			}
		}
	}()

	go func() {
		for range s.tickerDatastoreSync.C {
			if err := s.syncPeerDatastores(); err != nil {
				s.errCh <- err
			}
		}
	}()

	// start services
	wg := &sync.WaitGroup{}
	start := time.Now()
	for _, svc := range s.services {
		if err := s.validateServiceRequires(svc.Requires()); err != nil {
			return err
		}
		wg.Add(1)
		go func(svc services.Service) {
			defer wg.Done()
			logrus.WithFields(logrus.Fields{
				"service": svc.ID(),
			}).Debug("starting service")
			doneCh := make(chan bool, 1)
			go func() {
				if err := svc.Start(); err != nil {
					s.errCh <- err
				}
				doneCh <- true
			}()
			select {
			case <-doneCh:
				return
			case <-time.After(serviceStartTimeout):
				s.errCh <- fmt.Errorf("timeout starting service %s", svc.ID())
			}
		}(svc)
	}
	wg.Wait()

	logrus.WithFields(logrus.Fields{
		"duration": time.Since(start),
	}).Debug("services started")

	return nil
}

func (s *Server) validateServiceRequires(reqs []services.Type) error {
	for _, t := range reqs {
		if _, ok := s.services[t]; !ok {
			return errors.Wrap(ErrServiceNotRegistered, string(t))
		}
	}
	return nil
}

func (s *Server) syncPeerDatastores() error {
	lc, err := s.client(s.agent.Self().Address)
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
			"peer": peer.ID,
			"addr": peer.Address,
		}).Debug("syncing peer datastore")
		c, err := s.client(peer.Address)
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

		logrus.Debugf("synchronized %d operations from peer %s", count, peer.ID)

	}

	return nil
}

func (s *Server) GenerateProfile() (string, error) {
	tmpfile, err := ioutil.TempFile("", "stellar-profile-")
	if err != nil {
		return "", err
	}
	runtime.GC()
	if err := pprof.WriteHeapProfile(tmpfile); err != nil {
		return "", err
	}
	tmpfile.Close()
	return tmpfile.Name(), nil
}

func (s *Server) Stop() error {
	logrus.Debug("stopping server")
	s.tickerReconcile.Stop()
	s.tickerDatastoreSync.Stop()

	// shutdown server
	if err := s.shutdown(); err != nil {
		return err
	}

	logrus.Debug("shutting down agent")

	// shutdown element agent
	if err := s.agent.Shutdown(); err != nil {
		return err
	}

	return nil
}

func (s *Server) client(address string) (*client.Client, error) {
	opts, err := client.DialOptionsFromConfig(s.config)
	if err != nil {
		return nil, err
	}
	return client.NewClient(address, opts...)
}

func (s *Server) shutdown() error {
	// shutdown services
	wg := &sync.WaitGroup{}
	start := time.Now()
	for _, svc := range s.services {
		wg.Add(1)
		go func(svc services.Service) {
			defer wg.Done()
			logrus.WithFields(logrus.Fields{
				"service": svc.ID(),
			}).Debug("shutting down service")
			doneCh := make(chan bool, 1)
			go func() {
				if err := svc.Stop(); err != nil {
					s.errCh <- err
				}
				doneCh <- true
			}()
			select {
			case <-doneCh:
				return
			case <-time.After(serviceStopTimeout):
				s.errCh <- fmt.Errorf("timeout stopping service %s", svc.ID())
			}
		}(svc)
	}
	wg.Wait()
	logrus.WithFields(logrus.Fields{
		"duration": time.Since(start),
	}).Debug("services stopped")

	return nil
}
