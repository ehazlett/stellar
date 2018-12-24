package version

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/stellarproject/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/version/v1"
	"github.com/ehazlett/stellar/services"
	"github.com/ehazlett/stellar/version"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.version.v1"
)

type service struct {
	containerdAddr string
	namespace      string
}

func New(cfg *stellar.Config, _ *element.Agent) (services.Service, error) {
	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterVersionServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Type() services.Type {
	return services.VersionService
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

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) containerdVersion(ctx context.Context) (containerd.Version, error) {
	c, err := s.containerd()
	if err != nil {
		return containerd.Version{}, err
	}
	defer c.Close()

	return c.Version(ctx)
}

func (s *service) Version(ctx context.Context, _ *types.Empty) (*api.VersionResponse, error) {
	v, err := s.containerdVersion(ctx)
	if err != nil {
		return nil, err
	}

	return &api.VersionResponse{
		Version:            version.Version,
		Revision:           version.GitCommit,
		ContainerdVersion:  v.Version,
		ContainerdRevision: v.Revision,
	}, nil
}
