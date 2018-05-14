package health

import (
	"context"
	"runtime"
	"time"

	"github.com/cloudfoundry/gosigar"
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	containers, err := c.Containers(ctx)
	if err != nil {
		return nil, err
	}

	images, err := c.ListImages(ctx)
	if err != nil {
		return nil, err
	}

	osInfo, err := OSInfo()
	if err != nil {
		return nil, err
	}

	memory := sigar.Mem{}
	if err := memory.Get(); err != nil {
		return nil, err
	}

	return &api.HealthResponse{
		OsName:      osInfo.OSName,
		OsVersion:   osInfo.OSVersion,
		Uptime:      types.DurationProto(time.Now().Sub(s.started)),
		Cpus:        int64(runtime.NumCPU()),
		MemoryTotal: int64(memory.Total),
		MemoryFree:  int64(memory.Free),
		MemoryUsed:  int64(memory.Used),
		Containers:  int64(len(containers)),
		Images:      int64(len(images)),
	}, nil
}
