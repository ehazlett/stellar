package client

import (
	"context"

	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
)

type cluster struct {
	client clusterapi.ClusterClient
}

func (c *cluster) Nodes() ([]*clusterapi.Node, error) {
	ctx := context.Background()
	resp, err := c.client.Nodes(ctx, &clusterapi.NodesRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Nodes, nil
}

func (c *cluster) Containers() ([]*nodeapi.Container, error) {
	ctx := context.Background()
	resp, err := c.client.Containers(ctx, &clusterapi.ContainersRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Containers, nil
}
