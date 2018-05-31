package element

import (
	"context"

	nodeapi "github.com/ehazlett/element/api/services/node/v1"
)

func (c *Client) Containers() ([]*nodeapi.Container, error) {
	ctx := context.Background()
	resp, err := c.NodeService.Containers(ctx, &nodeapi.ContainersRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Containers, nil
}
