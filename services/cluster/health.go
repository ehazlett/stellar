package cluster

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/cluster/v1"
	healthapi "github.com/ehazlett/stellar/api/services/health/v1"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Health(ctx context.Context, req *api.HealthRequest) (*api.HealthResponse, error) {
	nodes, err := s.nodes()
	if err != nil {
		return nil, err
	}

	status := map[*api.Node]*healthapi.HealthResponse{}
	for _, node := range nodes {
		nc, err := client.NewClient(node.Addr)
		if err != nil {
			return nil, err
		}

		resp, err := nc.HealthService().Health(context.Background(), &ptypes.Empty{})
		if err != nil {
			return nil, err
		}
		status[node] = resp
		nc.Close()
	}

	nodeHealth := []*api.NodeHealth{}
	for node, resp := range status {
		nodeHealth = append(nodeHealth, &api.NodeHealth{
			Node:   node,
			Health: resp.Health,
		})
	}

	return &api.HealthResponse{
		Nodes: nodeHealth,
	}, nil
}
