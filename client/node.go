package client

import (
	"context"
	"time"

	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
)

type node struct {
	client nodeapi.NodeClient
}

func (n *node) Containers(filters ...string) ([]*nodeapi.Container, error) {
	ctx := context.Background()
	resp, err := n.client.Containers(ctx, &nodeapi.ContainersRequest{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	return resp.Containers, nil
}

func (n *node) Container(id string) (*nodeapi.Container, error) {
	ctx := context.Background()
	resp, err := n.client.Container(ctx, &nodeapi.ContainerRequest{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return resp.Container, nil
}

func (n *node) CreateContainer(appName string, service *nodeapi.Service) error {
	ctx := context.Background()
	if _, err := n.client.CreateContainer(ctx, &nodeapi.CreateContainerRequest{
		Application: appName,
		Service:     service,
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) DeleteContainer(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err := n.client.DeleteContainer(ctx, &nodeapi.DeleteContainerRequest{
		ID: id,
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) SetupContainerNetwork(id, ip, network, gateway string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err := n.client.SetupContainerNetwork(ctx, &nodeapi.ContainerNetworkRequest{
		ID:      id,
		IP:      ip,
		Network: network,
		Gateway: gateway,
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) Images() ([]*nodeapi.Image, error) {
	ctx := context.Background()
	resp, err := n.client.Images(ctx, &nodeapi.ImagesRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Images, nil
}
