package proxy

import (
	"fmt"
	"time"

	"github.com/containerd/containerd"
	"github.com/ehazlett/blackbird"
	blackbirdserver "github.com/ehazlett/blackbird/server"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	serviceID         = "stellar.services.proxy.v1"
	blackbirdGRPCAddr = "unix:///run/blackbird.sock"
	dsProxyBucketName = "stellar." + stellar.APIVersion + ".services.proxy"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr string
	namespace      string
	agent          *element.Agent
	cfg            *stellar.Config
	errCh          chan error

	// set on start
	server  *blackbirdserver.Server
	bclient *blackbird.Client
}

func New(cfg *stellar.Config, agent *element.Agent) (*service, error) {
	errCh := make(chan error)
	go func() {
		for {
			err := <-errCh
			logrus.Errorf("proxy: %s", err)
		}
	}()

	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		agent:          agent,
		cfg:            cfg,
		errCh:          errCh,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterProxyServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Start() error {
	client, err := s.client()
	if err != nil {
		return err
	}
	config := &blackbird.Config{
		GRPCAddr:  blackbirdGRPCAddr,
		HTTPPort:  s.cfg.ProxyHTTPPort,
		HTTPSPort: s.cfg.ProxyHTTPSPort,
		Debug:     false,
	}
	ds, err := newDatastore(client)
	if err != nil {
		return err
	}
	srv, err := blackbirdserver.NewServer(config, ds)
	if err != nil {
		return err
	}
	if err := srv.Run(); err != nil {
		return err
	}

	bc, err := blackbird.NewClient(blackbirdGRPCAddr)
	if err != nil {
		return err
	}
	s.bclient = bc

	t := time.NewTicker(5 * time.Second)
	go func() {
		for range t.C {
			logrus.Debug("proxy reload")
			// we do a periodic reload.  this might be better at scale
			// if we check for application updates before trying to reload
			if err := s.reload(); err != nil {
				logrus.Error(err)
				continue
			}
		}
	}()

	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) client() (*client.Client, error) {
	peer := s.agent.Self()
	return client.NewClient(peer.Address)
}

func (s *service) peerAddr() (string, error) {
	peer := s.agent.Self()
	return peer.Address, nil
}

func (s *service) nodeName() string {
	return s.agent.Self().ID
}

func (s *service) nodeClient(id string) (*client.Client, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	nodes, err := c.Cluster().Nodes()
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if node.ID == id {
			return client.NewClient(node.Address)
		}
	}

	return nil, fmt.Errorf("node %s not found in cluster", id)
}

func (s *service) getApplications() ([]*applicationapi.App, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.Application().List()
}
