package stellar

import (
	"context"

	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
)

func (c *Client) Nodes() ([]*clusterapi.Node, error) {
	ctx := context.Background()
	resp, err := c.ClusterService.Nodes(ctx, &clusterapi.NodesRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Nodes, nil
}
