package cluster

import (
	"github.com/containerd/containerd"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/cluster/v1"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.cluster.v1"
)

type service struct {
	containerdAddr string
	namespace      string
	agent          *element.Agent
}

func New(cfg *stellar.Config, a *element.Agent) (*service, error) {
	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		agent:          a,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterClusterServer(server, s)
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

func (s *service) nodes() ([]*api.Node, error) {
	peer := s.agent.Self()
	nodes := []*api.Node{
		{
			ID:      peer.ID,
			Address: peer.Address,
		},
	}

	peers, err := s.agent.Peers()
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		nodes = append(nodes, &api.Node{
			ID:      peer.ID,
			Address: peer.Address,
		})
	}

	return nodes, nil
}
