package health

import (
	"context"
	"runtime"
	"time"

	"github.com/cloudfoundry/gosigar"
	"github.com/ehazlett/element"
	api "github.com/ehazlett/stellar/api/services/health/v1"
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

func New(a *element.Agent) (*service, error) {
	return &service{
		agent:   a,
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

func (s *service) Start() error {
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
			Name: peer.Name,
			Addr: peer.Addr,
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
			MemoryUsed:  int64(memory.Used),
			Peers:       peers,
		},
	}, nil
}
