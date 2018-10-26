package cluster

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/cluster/v1"
)

func (s *service) Containers(ctx context.Context, req *api.ContainersRequest) (*api.ContainersResponse, error) {
	var containers []*api.Container

	resp, err := s.Nodes(ctx, &api.NodesRequest{})
	if err != nil {
		return nil, err
	}

	for _, node := range resp.Nodes {
		c, err := s.client(node.Address)
		if err != nil {
			return nil, err
		}
		cont, err := c.Node().Containers(req.Filters...)
		if err != nil {
			return nil, err
		}
		for _, container := range cont {
			containers = append(containers, &api.Container{
				Container: container,
				Node: &api.Node{
					ID:      node.ID,
					Address: node.Address,
				},
			})
		}
		c.Close()
	}

	return &api.ContainersResponse{
		Containers: containers,
	}, nil
}
