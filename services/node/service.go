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

func Register(server *grpc.Server, containerdAddr, namespace string) error {
	s := &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
	}
	api.RegisterNodeServer(server, s)
	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return containerd.New(s.containerdAddr, containerd.WithDefaultNamespace(s.namespace))
}
