package network

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/containerd/containerd/errdefs"
	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

var (
	ErrRouteExists = errors.New("route exists in configuration")
	// format: routes.<cidr>
	dsRoutesKey = "routes.%s"
)

func (s *service) AddRoute(ctx context.Context, req *api.AddRouteRequest) (*ptypes.Empty, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	routeData := []byte(req.CIDR + ":" + req.Target)
	routeKey := fmt.Sprintf(dsRoutesKey, req.CIDR)
	if err := c.Datastore().Set(dsNetworkBucketName, routeKey, routeData, true); err != nil {
		return nil, err
	}
	return &ptypes.Empty{}, nil
}

func (s *service) DeleteRoute(ctx context.Context, req *api.DeleteRouteRequest) (*ptypes.Empty, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	routeKey := fmt.Sprintf(dsRoutesKey, req.CIDR)
	if err := c.Datastore().Delete(dsNetworkBucketName, routeKey, true); err != nil {
		return nil, err
	}
	return &ptypes.Empty{}, nil
}

func (s *service) Routes(ctx context.Context, _ *ptypes.Empty) (*api.RoutesResponse, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var routes []*api.Route
	searchKey := fmt.Sprintf(dsRoutesKey, "")
	results, err := c.Datastore().Search(dsNetworkBucketName, searchKey)
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
	}
	for _, kv := range results {
		rt := strings.Split(string(kv.Value), ":")
		if len(rt) != 2 {
			logrus.Errorf("invalid route format: %s", string(kv.Value))
			continue
		}
		routes = append(routes, &api.Route{
			CIDR:   rt[0],
			Target: rt[1],
		})
	}
	return &api.RoutesResponse{
		Routes: routes,
	}, nil
}
