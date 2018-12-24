package client

import (
	"context"
	"time"

	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
)

type node struct {
	client runtimeapi.NodeClient
}

func (n *node) ID() (string, error) {
	ctx := context.Background()
	resp, err := n.client.Info(ctx, &runtimeapi.InfoRequest{})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (n *node) Containers(filters ...string) ([]*runtimeapi.Container, error) {
	ctx := context.Background()
	resp, err := n.client.Containers(ctx, &runtimeapi.ContainersRequest{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	return resp.Containers, nil
}

func (n *node) Container(id string) (*runtimeapi.Container, error) {
	ctx := context.Background()
	resp, err := n.client.Container(ctx, &runtimeapi.ContainerRequest{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return resp.Container, nil
}

func (n *node) CreateContainer(appName string, service *runtimeapi.Service, containerID string) error {
	ctx := context.Background()
	if _, err := n.client.CreateContainer(ctx, &runtimeapi.CreateContainerRequest{
		Application: appName,
		Service:     service,
		ContainerID: containerID,
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) RestartContainer(id string) error {
	ctx := context.Background()
	if _, err := n.client.RestartContainer(ctx, &runtimeapi.RestartContainerRequest{
		ID: id,
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) DeleteContainer(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err := n.client.DeleteContainer(ctx, &runtimeapi.DeleteContainerRequest{
		ID: id,
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) SetupContainerNetwork(id, ip, network, gateway string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err := n.client.SetupContainerNetwork(ctx, &runtimeapi.ContainerNetworkRequest{
		ID:      id,
		IP:      ip,
		Network: network,
		Gateway: gateway,
	}); err != nil {
		return err
	}

	return nil
}

func (n *node) Images() ([]*runtimeapi.Image, error) {
	ctx := context.Background()
	resp, err := n.client.Images(ctx, &runtimeapi.ImagesRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Images, nil
}
