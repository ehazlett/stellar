package client

import (
	"context"

	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
)

type node struct {
	client nodeapi.NodeClient
}

func (n *node) Containers() ([]*nodeapi.Container, error) {
	ctx := context.Background()
	resp, err := n.client.Containers(ctx, &nodeapi.ContainersRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Containers, nil
}

func (n *node) Images() ([]*nodeapi.Image, error) {
	ctx := context.Background()
	resp, err := n.client.Images(ctx, &nodeapi.ImagesRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Images, nil
}
