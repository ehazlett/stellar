package cluster

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/cluster/v1"
)

func (s *service) Nodes(ctx context.Context, req *api.NodesRequest) (*api.NodesResponse, error) {
	peer, err := s.agent.LocalNode()
	if err != nil {
		return nil, err
	}
	nodes := []*api.Node{
		{
			Name: peer.Name,
			Addr: peer.Addr,
		},
	}

	peers, err := s.agent.Peers()
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		nodes = append(nodes, &api.Node{
			Name: peer.Name,
			Addr: peer.Addr,
		})
	}

	return &api.NodesResponse{
		Nodes: nodes,
	}, nil
}
