package cluster

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/cluster/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
	"github.com/ehazlett/stellar/client"
)

func (s *service) Containers(ctx context.Context, req *api.ContainersRequest) (*api.ContainersResponse, error) {
	var containers []*nodeapi.Container

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
		containers = append(containers, cont...)
		c.Close()
	}

	return &api.ContainersResponse{
		Containers: containers,
	}, nil
}
