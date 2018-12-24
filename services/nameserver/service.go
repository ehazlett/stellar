package nameserver

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/stellarproject/element"
	"github.com/ehazlett/stellar"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"

	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
)

const (
	serviceID              = "stellar.services.nameserver.v1"
	dsNameserverBucketName = "stellar." + stellar.APIVersion + ".services.nameserver"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr  string
	namespace       string
	bridge          string
	upstreamDNSAddr string
	agent           *element.Agent
	config          *stellar.Config
}

func New(cfg *stellar.Config, agent *element.Agent) (services.Service, error) {
	srv := &service{
		containerdAddr:  cfg.ContainerdAddr,
		namespace:       cfg.Namespace,
		bridge:          cfg.Bridge,
		upstreamDNSAddr: cfg.UpstreamDNSAddr,
		agent:           agent,
		config:          cfg,
	}

	return srv, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNameserverServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Type() services.Type {
	return services.NameserverService
}

func (s *service) Requires() []services.Type {
	return []services.Type{
		services.DatastoreService,
	}
}

func (s *service) Info(ctx context.Context, req *api.InfoRequest) (*api.InfoResponse, error) {
	return &api.InfoResponse{
		ID: serviceID,
	}, nil
}

func (s *service) Start() error {
	return s.startDNSServer()
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
