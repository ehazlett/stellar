package events

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/events/v1"
)

func (s *service) Endpoint(ctx context.Context, _ *api.EndpointRequest) (*api.EndpointResponse, error) {
	return &api.EndpointResponse{
		Endpoint: &api.Endpoint{
			Address: s.config.EventsClusterAddress,
		},
	}, nil
}
