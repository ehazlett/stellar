package client

import (
	"context"

	healthapi "github.com/ehazlett/stellar/api/services/health/v1"
	ptypes "github.com/gogo/protobuf/types"
)

type health struct {
	client healthapi.HealthClient
}

func (h *health) Peers() ([]*healthapi.Peer, error) {
	ctx := context.Background()
	resp, err := h.client.Health(ctx, &ptypes.Empty{})
	if err != nil {
		return nil, err
	}

	return resp.Health.Peers, nil
}
