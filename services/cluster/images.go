package cluster

import (
	"context"

	api "github.com/ehazlett/element/api/services/cluster/v1"
	nodeapi "github.com/ehazlett/element/api/services/node/v1"
)

func (s *service) Images(ctx context.Context, req *api.ImagesRequest) (*api.ImagesResponse, error) {
	var images []*nodeapi.Image

	return &api.ImagesResponse{
		Images: images,
	}, nil
}
