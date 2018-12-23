package health

import (
	"context"
	"runtime"
	"time"

	"github.com/cloudfoundry/gosigar"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/health/v1"
	"github.com/ehazlett/stellar/services"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.health.v1"
)

type service struct {
	agent   *element.Agent
	started time.Time
}

func New(_ *stellar.Config, agent *element.Agent) (services.Service, error) {
	return &service{
		agent:   agent,
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

func (s *service) Type() services.Type {
	return services.HealthService
}

func (s *service) Requires() []services.Type {
	return nil
}

func (s *service) Info(ctx context.Context, req *api.InfoRequest) (*api.InfoResponse, error) {
	return &api.InfoResponse{
		ID: serviceID,
	}, nil
}

func (s *service) Start() error {
	return nil
}

func (s *service) Stop() error {
	return nil
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
	p, err := s.agent.Peers()
	if err != nil {
		return nil, err
	}
	peers := []*api.Peer{}
	for _, peer := range p {
		peers = append(peers, &api.Peer{
			ID:      peer.ID,
			Address: peer.Address,
		})
	}

	return &api.HealthResponse{
		Health: &api.NodeHealth{
			OSName:      osInfo.OSName,
			OSVersion:   osInfo.OSVersion,
			StartedAt:   ts,
			Cpus:        int64(runtime.NumCPU()),
			MemoryTotal: int64(memory.Total),
			MemoryFree:  int64(memory.Free),
			MemoryUsed:  int64(memory.ActualUsed),
			Peers:       peers,
		},
	}, nil
}
