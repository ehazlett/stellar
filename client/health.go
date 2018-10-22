package client

import (
	"context"

	healthapi "github.com/ehazlett/stellar/api/services/health/v1"
	ptypes "github.com/gogo/protobuf/types"
)

type health struct {
	client healthapi.HealthClient
}

func (h *health) ID() (string, error) {
	ctx := context.Background()
	resp, err := h.client.Info(ctx, &healthapi.InfoRequest{})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (h *health) Peers() ([]*healthapi.Peer, error) {
	ctx := context.Background()
	resp, err := h.client.Health(ctx, &ptypes.Empty{})
	if err != nil {
		return nil, err
	}

	return resp.Health.Peers, nil
}
