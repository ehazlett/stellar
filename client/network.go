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
