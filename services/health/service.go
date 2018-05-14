package health

import (
	"context"
	"time"

	"github.com/containerd/containerd"
	api "github.com/ehazlett/element/api/services/health/v1"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

type service struct {
	containerdAddr string
	namespace      string
	started        time.Time
}

func Register(server *grpc.Server, containerdAddr, namespace string) error {
	s := &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
		started:        time.Now(),
	}
	api.RegisterHealthServer(server, s)
	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return containerd.New(s.containerdAddr, containerd.WithDefaultNamespace(s.namespace))
}

func (s *service) Health(ctx context.Context, _ *types.Empty) (*api.HealthResponse, error) {
	c, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	containers, err := c.Containers(context.Background())
	if err != nil {
		return nil, err
	}

	images, err := c.ListImages(context.Background())
	if err != nil {
		return nil, err
	}

	osInfo, err := OSInfo()
	if err != nil {
		return nil, err
	}

	return &api.HealthResponse{
		OsName:     osInfo.OSName,
		OsVersion:  osInfo.OSVersion,
		Uptime:     types.DurationProto(time.Now().Sub(s.started)),
		Containers: int64(len(containers)),
		Images:     int64(len(images)),
	}, nil
}
