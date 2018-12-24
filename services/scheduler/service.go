package scheduler

import (
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/scheduler/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	"github.com/stellarproject/element"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.scheduler.v1"
)

type service struct {
	config *stellar.Config
}

func New(cfg *stellar.Config, _ *element.Agent) (services.Service, error) {
	return &service{
		config: cfg,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterSchedulerServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Type() services.Type {
	return services.SchedulerService
}

func (s *service) Requires() []services.Type {
	return nil
}

func (s *service) Start() error {
	return nil
}

func (s *service) Stop() error {
	return nil
}

func (s *service) client(address string) (*client.Client, error) {
	opts, err := client.DialOptionsFromConfig(s.config)
	if err != nil {
		return nil, err
	}
	return client.NewClient(address, opts...)
}
