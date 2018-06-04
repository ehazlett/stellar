package node

import (
	"github.com/containerd/containerd"
	api "github.com/ehazlett/stellar/api/services/node/v1"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.node.v1"
)

type service struct {
	containerdAddr string
	namespace      string
}

func New(containerdAddr, namespace string) (*service, error) {
	return &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNodeServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) containerd() (*containerd.Client, error) {
	return containerd.New(s.containerdAddr, containerd.WithDefaultNamespace(s.namespace))
}
