package node

import (
	"context"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/node/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func (s *service) Containers(ctx context.Context, req *api.ContainersRequest) (*api.ContainersResponse, error) {
	c, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	containers, err := c.Containers(ctx, req.Filters...)
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

func (s *service) DeleteContainer(ctx context.Context, req *api.DeleteContainerRequest) (*ptypes.Empty, error) {
	c, err := s.containerd()
	if err != nil {
		return empty, err
	}
	defer c.Close()

	client, err := s.client()
	if err != nil {
		return empty, err
	}
	defer client.Close()

	wg := &sync.WaitGroup{}
	logrus.Debugf("delete: LoadContainer")
	container, err := c.LoadContainer(ctx, req.ID)
	if err != nil {
		return empty, err
	}

	logrus.Debugf("delete: container.Task")
	task, err := container.Task(ctx, nil)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return empty, err
		}
	}
	logrus.Debugf("delete: task wait")
	wait, err := task.Wait(ctx)
	if err != nil {
		return empty, err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Debugf("delete: task kill")
		if err := task.Kill(ctx, unix.SIGTERM, containerd.WithKillAll); err != nil {
			logrus.Errorf("error killing application task: %s", err)
			return
		}
		select {
		case <-wait:
			logrus.Debugf("delete: task delete")
			task.Delete(ctx)
			return
		case <-time.After(5 * time.Second):
			logrus.Debugf("delete: task force kill")
			if err := task.Kill(ctx, unix.SIGKILL, containerd.WithKillAll); err != nil {
				logrus.Errorf("error force killing application task: %s", err)
				return
			}
			return
		}
	}()

	wg.Wait()

	networkEnabled, err := s.networkEnabled(ctx, container)
	if err != nil {
		return empty, err
	}

	logrus.Debugf("delete: container delete")
	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return empty, err
	}

	if networkEnabled {
		// release IP
		logrus.Debugf("delete: GetIP")
		ip, err := client.Network().GetIP(req.ID, s.agent.Config().NodeName)
		if err != nil {
			logrus.Warnf("error getting ip for containerd %s: %s", req.ID, err)
		}
		logrus.Debugf("delete: ReleaseIP")
		if _, err := client.Network().ReleaseIP(req.ID, ip.String(), s.agent.Config().NodeName); err != nil {
			logrus.Warnf("error getting ip for containerd %s: %s", req.ID, err)
		}
	}

	return empty, nil
}

func (s *service) networkEnabled(ctx context.Context, container containerd.Container) (bool, error) {
	labels, err := container.Labels(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := labels[stellar.StellarNetworkLabel]; ok {
		return true, nil
	}

	return false, nil
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
