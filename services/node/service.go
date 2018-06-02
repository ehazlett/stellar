package node

import (
	"github.com/containerd/containerd"
	api "github.com/ehazlett/element/api/services/node/v1"
	"google.golang.org/grpc"
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
	return "node"
}

func (s *service) containerd() (*containerd.Client, error) {
	return containerd.New(s.containerdAddr, containerd.WithDefaultNamespace(s.namespace))
}
