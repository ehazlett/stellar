package network

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Configure(ctx context.Context, req *api.ConfigureRequest) (*ptypes.Empty, error) {
	// TODO: configure local route
	return &ptypes.Empty{}, nil
}
