package cluster

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/cluster/v1"
	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
)

func (s *service) Images(ctx context.Context, req *api.ImagesRequest) (*api.ImagesResponse, error) {
	var images []*runtimeapi.Image

	return &api.ImagesResponse{
		Images: images,
	}, nil
}
