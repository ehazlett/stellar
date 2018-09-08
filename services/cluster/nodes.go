package cluster

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/cluster/v1"
)

func (s *service) Nodes(ctx context.Context, req *api.NodesRequest) (*api.NodesResponse, error) {
	nodes, err := s.nodes()
	if err != nil {
		return nil, err
	}
	return &api.NodesResponse{
		Nodes: nodes,
	}, nil
}
