package network

import (
	"context"
	"net"

	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) AllocateSubnet(ctx context.Context, req *api.AllocateSubnetRequest) (*ptypes.Empty, error) {
	// TODO: add alias address for subnet gateway IP to local bind nic

	// TODO: notify peers of new route

	return &ptypes.Empty{}, nil
}

func (s *service) DeallocateSubnet(ctx context.Context, req *api.DeallocateSubnetRequest) (*ptypes.Empty, error) {
	// TODO: remove alias gateway address from local nic

	// TODO: notify peers of removed route

	return &ptypes.Empty{}, nil
}

func (s *service) Subnets(ctx context.Context, _ *ptypes.Empty) (*api.SubnetsResponse, error) {
	subs, err := divideSubnet(s.network, maxSubnets)
	if err != nil {
		return nil, err
	}

	var subnets []*api.Subnet
	for _, subnet := range subs {
		ip := subnet.IP
		gw := net.IPv4(ip[0], ip[1], ip[2], 1)
		subnets = append(subnets, &api.Subnet{
			CIDR:    subnet.String(),
			Gateway: gw.String(),
		})
	}

	return &api.SubnetsResponse{
		Subnets: subnets,
	}, nil
}
