package network

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/containerd/containerd/errdefs"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

var (
	// ErrNoAvailableSubnets returns an error if there are no available subnets to allocate
	ErrNoAvailableSubnets = errors.New("no available subnets in network configuration")
	// ErrSubnetNotFound returns an error if no subnet is found for the node
	ErrSubnetNotFound = errors.New("subnet not found")
	// format: subnets.<node>
	dsSubnetsKey = "subnets.%s"
)

func (s *service) AllocateSubnet(ctx context.Context, req *api.AllocateSubnetRequest) (*api.AllocateSubnetResponse, error) {
	// check for existing assigned subnet; if not, allocate
	localSubnetKey := fmt.Sprintf(dsSubnetsKey, req.Node)

	resp, err := s.Subnets(ctx, nil)
	if err != nil {
		return nil, err
	}

	subnets := resp.Subnets

	if len(subnets) == 0 {
		return nil, ErrNoAvailableSubnets
	}

	localSubnetResp, err := s.ds.Get(ctx, &datastoreapi.GetRequest{
		Bucket: dsNetworkBucketName,
		Key:    localSubnetKey,
	})
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
	}
	bSubnetCIDR := localSubnetResp.Data.Value

	if bytes.Equal(bSubnetCIDR, []byte("")) {
		logrus.Debug("local subnet key not found; assigning new subnet")

		searchKey := fmt.Sprintf(dsSubnetsKey, "")
		searchResp, err := s.ds.Search(ctx, &datastoreapi.SearchRequest{
			Bucket: dsNetworkBucketName,
			Prefix: searchKey,
		})
		if err != nil {
			err = errdefs.FromGRPC(err)
			if !errdefs.IsNotFound(err) {
				return nil, err
			}
		}
		existingSubnets := searchResp.Data

		assigned := len(existingSubnets)
		if len(subnets) < assigned {
			return nil, fmt.Errorf("no available subnet for current node; need %d subnets", assigned)
		}

		bSubnetCIDR = []byte(subnets[assigned].CIDR)
		if _, err := s.ds.Set(ctx, &datastoreapi.SetRequest{
			Bucket: dsNetworkBucketName,
			Key:    localSubnetKey,
			Value:  bSubnetCIDR,
			Sync:   true,
		}); err != nil {
			return nil, err
		}
	}

	return &api.AllocateSubnetResponse{
		SubnetCIDR: string(bSubnetCIDR),
		Node:       req.Node,
	}, nil
}

func (s *service) GetSubnet(ctx context.Context, req *api.GetSubnetRequest) (*api.GetSubnetResponse, error) {
	localSubnetKey := fmt.Sprintf(dsSubnetsKey, req.Node)

	localSubnetResp, err := s.ds.Get(ctx, &datastoreapi.GetRequest{
		Bucket: dsNetworkBucketName,
		Key:    localSubnetKey,
	})
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return nil, ErrSubnetNotFound
		}
	}
	return &api.GetSubnetResponse{
		SubnetCIDR: string(localSubnetResp.Data.Value),
	}, nil
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
