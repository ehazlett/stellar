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

func (s *service) Container(ctx context.Context, req *api.ContainerRequest) (*api.ContainerResponse, error) {
	c, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	container, err := c.LoadContainer(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	cont, err := s.containerToProto(container)
	if err != nil {
		return nil, err
	}

	return &api.ContainerResponse{
		Container: cont,
	}, nil
}

func (s *service) containersToProto(containers []containerd.Container) ([]*api.Container, error) {
	var c []*api.Container
	for _, container := range containers {
		conv, err := s.containerToProto(container)
		if err != nil {
			return nil, err
		}

		c = append(c, conv)
	}

	return c, nil
}

func (s *service) containerToProto(container containerd.Container) (*api.Container, error) {
	info, err := container.Info(context.Background())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pid := uint32(0)

	// attempt to find task pid
	task, _ := container.Task(ctx, nil)
	if task != nil {
		pid = task.Pid()
	}

	return &api.Container{
		ID:     container.ID(),
		Image:  info.Image,
		Labels: info.Labels,
		Spec: &any.Any{
			TypeUrl: info.Spec.TypeUrl,
			Value:   info.Spec.Value,
		},
		Snapshotter: info.Snapshotter,
		Task: &api.Container_Task{
			Pid: pid,
		},
		Runtime: info.Runtime.Name,
	}, nil
}
