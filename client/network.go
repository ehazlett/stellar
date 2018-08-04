package client

import (
	"context"
	"net"
	"time"

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

func (n *network) GetSubnet(node string) (string, error) {
	ctx := context.Background()
	resp, err := n.client.GetSubnet(ctx, &networkapi.GetSubnetRequest{
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

func (n *network) AllocateIP(id, node, subnetCIDR string) (net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := n.client.AllocateIP(ctx, &networkapi.AllocateIPRequest{
		ID:         id,
		Node:       node,
		SubnetCIDR: subnetCIDR,
	})
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(resp.IP)
	return ip, nil
}
