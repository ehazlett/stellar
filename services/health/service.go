package health

import (
	"context"
	"runtime"
	"time"

	"github.com/cloudfoundry/gosigar"
	api "github.com/ehazlett/stellar/api/services/health/v1"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.health.v1"
)

type service struct {
	started time.Time
}

func New() (*service, error) {
	return &service{
		started: time.Now(),
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterHealthServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Started() time.Time {
	return s.started
}

func (s *service) Health(ctx context.Context, _ *ptypes.Empty) (*api.HealthResponse, error) {
	osInfo, err := OSInfo()
	if err != nil {
		return nil, err
	}

	memory := sigar.Mem{}
	if err := memory.Get(); err != nil {
		return nil, err
	}

	ts, err := ptypes.TimestampProto(s.started)
	if err != nil {
		return nil, err
	}
	return &api.HealthResponse{
		OSName:      osInfo.OSName,
		OSVersion:   osInfo.OSVersion,
		StartedAt:   ts,
		Cpus:        int64(runtime.NumCPU()),
		MemoryTotal: int64(memory.Total),
		MemoryFree:  int64(memory.Free),
		MemoryUsed:  int64(memory.Used),
	}, nil
}
