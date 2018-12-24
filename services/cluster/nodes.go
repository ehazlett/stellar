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

func (s *service) nodes() ([]*api.Node, error) {
	peer := s.agent.Self()
	nodes := []*api.Node{
		{
			ID:      peer.ID,
			Address: peer.Address,
			Labels:  peer.Labels,
		},
	}

	peers, err := s.agent.Peers()
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		nodes = append(nodes, &api.Node{
			ID:      peer.ID,
			Address: peer.Address,
			Labels:  peer.Labels,
		})
	}

	return nodes, nil
}
