package cluster

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/cluster/v1"
	"github.com/ehazlett/stellar/client"
)

func (s *service) Containers(ctx context.Context, req *api.ContainersRequest) (*api.ContainersResponse, error) {
	var containers []*api.Container

	resp, err := s.Nodes(ctx, &api.NodesRequest{})
	if err != nil {
		return nil, err
	}

	for _, node := range resp.Nodes {
		c, err := client.NewClient(node.Addr)
		if err != nil {
			return nil, err
		}
		cont, err := c.Node().Containers()
		if err != nil {
			return nil, err
		}
		for _, container := range cont {
			containers = append(containers, &api.Container{
				Container: container,
				Node: &api.Node{
					Name: node.Name,
					Addr: node.Addr,
				},
			})
		}
		c.Close()
	}

	return &api.ContainersResponse{
		Containers: containers,
	}, nil
}
