package node

import (
	"context"
	"time"

	"github.com/containerd/containerd"
	api "github.com/ehazlett/stellar/api/services/node/v1"
	"github.com/golang/protobuf/ptypes/any"
)

func (s *service) Containers(ctx context.Context, req *api.ContainersRequest) (*api.ContainersResponse, error) {
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

	conv, err := s.containersToProto(containers)
	if err != nil {
		return nil, err
	}

	return &api.ContainersResponse{
		Containers: conv,
	}, nil
}

func (s *service) containersToProto(containers []containerd.Container) ([]*api.Container, error) {
	var c []*api.Container
	for _, container := range containers {
		info, err := container.Info(context.Background())
		if err != nil {
			return nil, err
		}
		c = append(c, &api.Container{
			ID:     container.ID(),
			Image:  info.Image,
			Labels: info.Labels,
			Spec: &any.Any{
				TypeUrl: info.Spec.TypeUrl,
				Value:   info.Spec.Value,
			},
			Snapshotter: info.Snapshotter,
		})
	}

	return c, nil
}
