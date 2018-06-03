package cluster

import (
	"github.com/containerd/containerd"
	"github.com/ehazlett/element/agent"
	api "github.com/ehazlett/element/api/services/cluster/v1"
	"google.golang.org/grpc"
)

const (
	serviceID = "element.services.cluster.v1"
)

type service struct {
	containerdAddr string
	namespace      string
	agent          *agent.Agent
}

func New(a *agent.Agent, containerdAddr, namespace string) (*service, error) {
	return &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
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

func (s *service) containerd() (*containerd.Client, error) {
	return containerd.New(s.containerdAddr, containerd.WithDefaultNamespace(s.namespace))
}
