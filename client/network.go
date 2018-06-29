package client

import (
	"context"

	networkapi "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
)

type network struct {
	client networkapi.NetworkClient
}

func (n *network) Subnets() ([]*networkapi.Subnet, error) {
	ctx := context.Background()
	resp, err := n.client.Subnets(ctx, &ptypes.Empty{})
	if err != nil {
		return nil, err
	}

	return resp.Subnets, nil
}

func (n *network) AllocateSubnet(node string) (string, error) {
	ctx := context.Background()
	resp, err := n.client.AllocateSubnet(ctx, &networkapi.AllocateSubnetRequest{
		Node: node,
	})
	if err != nil {
		return "", err
	}

	return resp.SubnetCIDR, nil
}

func (n *network) AddRoute(cidr, target string) error {
	ctx := context.Background()
	if _, err := n.client.AddRoute(ctx, &networkapi.AddRouteRequest{
		CIDR:   cidr,
		Target: target,
	}); err != nil {
		return err
	}

	return nil
}

func (n *network) DeleteRoute(cidr, target string) error {
	ctx := context.Background()
	if _, err := n.client.DeleteRoute(ctx, &networkapi.DeleteRouteRequest{
		CIDR:   cidr,
		Target: target,
	}); err != nil {
		return err
	}

	return nil
}

func (n *network) Routes() ([]*networkapi.Route, error) {
	ctx := context.Background()
	resp, err := n.client.Routes(ctx, &ptypes.Empty{})
	if err != nil {
		return nil, err
	}

	return resp.Routes, nil
}
