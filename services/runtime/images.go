package runtime

import (
	"context"
	"time"

	"github.com/containerd/containerd"
	api "github.com/ehazlett/stellar/api/services/runtime/v1"
)

func (s *service) Images(ctx context.Context, req *api.ImagesRequest) (*api.ImagesResponse, error) {
	c, err := s.containerd()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// TODO: make timeout configurable
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	images, err := c.ListImages(ctx)
	if err != nil {
		return nil, err
	}

	return &api.ImagesResponse{
		Images: s.imagesToProto(images),
	}, nil
}

func (s *service) imagesToProto(images []containerd.Image) []*api.Image {
	var i []*api.Image
	for _, image := range images {
		i = append(i, &api.Image{
			ID: image.Name(),
		})
	}

	return i
}
