package proxy

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/proxy/v1"
)

func (s *service) Backends(ctx context.Context, req *api.BackendRequest) (*api.BackendResponse, error) {
	var backends []*api.Backend

	return &api.BackendResponse{
		Backends: backends,
	}, nil
}
