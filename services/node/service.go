package node

import (
	"github.com/containerd/containerd"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/node/v1"
	"github.com/ehazlett/stellar/client"
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
	agent          *element.Agent
}

func New(cfg *stellar.Config, agent *element.Agent) (*service, error) {
	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		bridge:         cfg.Bridge,
		dataDir:        cfg.DataDir,
		stateDir:       cfg.StateDir,
		agent:          agent,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNodeServer(server, s)
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

func (s *service) client() (*client.Client, error) {
	peer, err := s.agent.LocalNode()
	if err != nil {
		return nil, err
	}
	return client.NewClient(peer.Addr)
}

func (s *service) peerAddr() (string, error) {
	peer, err := s.agent.LocalNode()
	if err != nil {
		return "", err
	}

	return peer.Addr, nil
}

func (s *service) nodeName() string {
	return s.agent.Config().NodeName
}
