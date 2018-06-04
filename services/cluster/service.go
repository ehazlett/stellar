package cluster

import (
	"github.com/containerd/containerd"
	"github.com/ehazlett/element"
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

func New(a *element.Agent, containerdAddr, namespace string) (*service, error) {
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
