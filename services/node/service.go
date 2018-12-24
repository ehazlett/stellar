package node

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/stellarproject/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/node/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.node.v1"
)

type service struct {
	containerdAddr string
	namespace      string
	bridge         string
	dataDir        string
	stateDir       string
	cniBinPaths    []string
	agent          *element.Agent
	config         *stellar.Config
}

func New(cfg *stellar.Config, agent *element.Agent) (services.Service, error) {
	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		bridge:         cfg.Bridge,
		dataDir:        cfg.DataDir,
		stateDir:       cfg.StateDir,
		cniBinPaths:    cfg.CNIBinPaths,
		agent:          agent,
		config:         cfg,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNodeServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Type() services.Type {
	return services.NodeService
}

func (s *service) Requires() []services.Type {
	return nil
}

func (s *service) Info(ctx context.Context, req *api.InfoRequest) (*api.InfoResponse, error) {
	return &api.InfoResponse{
		ID: serviceID,
	}, nil
}

func (s *service) Start() error {
	return nil
}

func (s *service) Stop() error {
	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) client(address string) (*client.Client, error) {
	opts, err := client.DialOptionsFromConfig(s.config)
	if err != nil {
		return nil, err
	}
	return client.NewClient(address, opts...)
}

func (s *service) peerAddr() (string, error) {
	peer := s.agent.Self()
	return peer.Address, nil
}

func (s *service) nodeName() string {
	return s.agent.Self().ID
}
