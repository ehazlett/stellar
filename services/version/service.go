package version

import (
	"context"

	api "github.com/ehazlett/element/api/services/version/v1"
	"github.com/ehazlett/element/version"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

type service struct{}

func Register(server *grpc.Server) error {
	s := &service{}
	api.RegisterVersionServer(server, s)
	return nil
}

func (s *service) Version(ctx context.Context, _ *types.Empty) (*api.VersionResponse, error) {
	return &api.VersionResponse{
		Version:            version.Version,
		Revision:           version.GitCommit,
		ContainerdVersion:  "",
		ContainerdRevision: "",
	}, nil
}
