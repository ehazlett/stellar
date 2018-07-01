package network

import (
	"context"
	"errors"
	"fmt"
	"strings"

	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	api "github.com/ehazlett/stellar/api/services/network/v1"
	"github.com/ehazlett/stellar/errdefs"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

var (
	ErrRouteExists = errors.New("route exists in configuration")
	// format: routes.<cidr>
	dsRoutesKey = "routes.%s"
)

func (s *service) AddRoute(ctx context.Context, req *api.AddRouteRequest) (*ptypes.Empty, error) {
	routeData := []byte(req.CIDR + ":" + req.Target)
	routeKey := fmt.Sprintf(dsRoutesKey, req.CIDR)
	if _, err := s.ds.Set(ctx, &datastoreapi.SetRequest{
		Bucket: dsNetworkBucketName,
		Key:    routeKey,
		Value:  routeData,
		Sync:   true,
	}); err != nil {
		return nil, err
	}
	return &ptypes.Empty{}, nil
}

func (s *service) DeleteRoute(ctx context.Context, req *api.DeleteRouteRequest) (*ptypes.Empty, error) {
	routeKey := fmt.Sprintf(dsRoutesKey, req.CIDR)
	if _, err := s.ds.Delete(ctx, &datastoreapi.DeleteRequest{
		Bucket: dsNetworkBucketName,
		Key:    routeKey,
		Sync:   true,
	}); err != nil {
		return nil, err
	}
	return &ptypes.Empty{}, nil
}

func (s *service) Routes(ctx context.Context, _ *ptypes.Empty) (*api.RoutesResponse, error) {
	var routes []*api.Route
	searchKey := fmt.Sprintf(dsRoutesKey, "")
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
	for _, kv := range searchResp.Data {
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
