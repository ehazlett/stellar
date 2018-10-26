package application

import (
	"github.com/containerd/containerd"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.application.v1"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr string
	namespace      string
	dataDir        string
	agent          *element.Agent
	config         *stellar.Config
}

func New(cfg *stellar.Config, agent *element.Agent) (*service, error) {
	return &service{
		// TODO: simplify config access
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		dataDir:        cfg.DataDir,
		agent:          agent,
		config:         cfg,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterApplicationServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Start() error {
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
