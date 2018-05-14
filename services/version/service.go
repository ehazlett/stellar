package version

import (
	"context"

	"github.com/containerd/containerd"
	api "github.com/ehazlett/element/api/services/version/v1"
	"github.com/ehazlett/element/version"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

type service struct {
	containerdAddr string
	namespace      string
}

func Register(server *grpc.Server, containerdAddr, namespace string) error {
	s := &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
	}
	api.RegisterVersionServer(server, s)
	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return containerd.New(s.containerdAddr, containerd.WithDefaultNamespace(s.namespace))
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
