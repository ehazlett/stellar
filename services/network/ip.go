package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/containerd/containerd/errdefs"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

var (
	ErrNoAvailableIP = errors.New("IP allocation exhausted")
	// format: ips.<node>.<id>
	dsIPsKey = "ips.%s.%s"
)

func (s *service) AllocateIP(ctx context.Context, req *api.AllocateIPRequest) (*api.AllocateIPResponse, error) {
	reservedIPs, err := s.getIPs(ctx, req.Node)
	if err != nil {
		return nil, err
	}

	if ip, exists := reservedIPs[req.ID]; exists {
		return &api.AllocateIPResponse{
			IP:   ip.String(),
			Node: req.Node,
		}, nil
	}

	lookup := map[string]string{}
	for id, ip := range reservedIPs {
		lookup[ip.String()] = id
	}

	logrus.Debugf("allocating ip for %s", req.ID)
	ip, ipnet, err := net.ParseCIDR(req.SubnetCIDR)
	if err != nil {
		return nil, err
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); nextIP(ip) {
		// filter out network, gateway and broadcast
		if !validIP(ip) {
			continue
		}
		if _, exists := lookup[ip.String()]; exists {
			// ip already reserved
			continue
		}
		ipKey := fmt.Sprintf(dsIPsKey, req.Node, req.ID)
		logrus.Debugf("ip key: %s", ipKey)
		if _, err := s.ds.Set(ctx, &datastoreapi.SetRequest{
			Bucket: dsNetworkBucketName,
			Key:    ipKey,
			Value:  []byte(ip.String()),
			Sync:   true,
		}); err != nil {
			return nil, err
		}

		logrus.Debugf("ip for %s: %s", req.ID, ip.String())
		return &api.AllocateIPResponse{
			IP:   ip.String(),
			Node: req.Node,
		}, nil
	}

	return nil, ErrNoAvailableIP
}

func (s *service) GetIP(ctx context.Context, req *api.GetIPRequest) (*api.GetIPResponse, error) {
	ipKey := fmt.Sprintf(dsIPsKey, req.Node, req.ID)
	resp, err := s.ds.Get(ctx, &datastoreapi.GetRequest{
		Bucket: dsNetworkBucketName,
		Key:    ipKey,
	})
	if err != nil {
		return nil, err
	}

	return &api.GetIPResponse{
		IP: string(resp.Data.Value),
	}, nil
}

func (s *service) ReleaseIP(ctx context.Context, req *api.ReleaseIPRequest) (*ptypes.Empty, error) {
	ipKey := fmt.Sprintf(dsIPsKey, req.Node, req.ID)
	if _, err := s.ds.Delete(ctx, &datastoreapi.DeleteRequest{
		Bucket: dsNetworkBucketName,
		Key:    ipKey,
		Sync:   true,
	}); err != nil {
		return empty, err
	}

	return empty, nil
}

func (s *service) getIPs(ctx context.Context, node string) (map[string]net.IP, error) {
	searchKey := fmt.Sprintf(dsIPsKey, node, "")
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
	ips := make(map[string]net.IP, len(searchResp.Data))
	for _, kv := range searchResp.Data {
		p := strings.Split(kv.Key, ".")
		if len(p) < 3 {
			logrus.Errorf("unexpected IP key format: %s", kv.Key)
			continue
		}
		id := strings.Join(p[2:], ".")
		ip := net.ParseIP(string(kv.Value))
		ips[id] = ip
	}

	return ips, nil
}

func nextIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func validIP(ip net.IP) bool {
	v := ip[len(ip)-1]
	switch v {
	case 0, 1, 255:
		return false
	}
	return true
}
