package cluster

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/cluster/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	"github.com/stellarproject/element"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.cluster.v1"
)

type service struct {
	containerdAddr string
	namespace      string
	agent          *element.Agent
	config         *stellar.Config
}

func New(cfg *stellar.Config, agent *element.Agent) (services.Service, error) {
	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		agent:          agent,
		config:         cfg,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterClusterServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Type() services.Type {
	return services.ClusterService
}

func (s *service) Requires() []services.Type {
	return []services.Type{
		services.RuntimeService,
		services.HealthService,
	}
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

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) client(address string) (*client.Client, error) {
	opts, err := client.DialOptionsFromConfig(s.config)
	if err != nil {
		return nil, err
	}
	return client.NewClient(address, opts...)
}
