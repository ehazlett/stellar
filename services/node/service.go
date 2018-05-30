package node

import (
	"context"
	"time"

	"github.com/containerd/containerd"
	api "github.com/ehazlett/element/api/services/node/v1"
	"github.com/gogo/protobuf/types"
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

func (s *service) Containers(ctx context.Context, _ *types.Empty) (*api.ContainersResponse, error) {
	c, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	containers, err := c.Containers(ctx)
	if err != nil {
		return nil, err
	}

	return &api.ContainersResponse{
		Containers: s.containersToProto(containers),
	}, nil
}

func (s *service) containersToProto(containers []containerd.Container) []*api.Container {
	var c []*api.Container
	for _, container := range containers {
		c = append(c, &api.Container{
			ID: container.ID(),
		})
	}

	return c
}
