package network

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) AllocateIP(ctx context.Context, req *api.AllocateIPRequest) (*api.AllocateIPResponse, error) {
	// TODO: allocate ip and save to datastore

	return &api.AllocateIPResponse{}, nil
}

func (s *service) ReleaseIP(ctx context.Context, req *api.ReleaseIPRequest) (*ptypes.Empty, error) {
	// TODO: remove from datastore

	return &ptypes.Empty{}, nil
}
