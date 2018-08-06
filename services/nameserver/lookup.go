package nameserver

import (
	"context"
	"encoding/json"

	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
)

func (s *service) Lookup(ctx context.Context, req *api.LookupRequest) (*api.LookupResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	val, err := c.Datastore().Get(dsNameserverBucketName, req.Query)
	if err != nil {
		return nil, err
	}

	var records []*api.Record
	if err := json.Unmarshal(val, &records); err != nil {
		return nil, err
	}

	return &api.LookupResponse{
		Name:    req.Query,
		Records: records,
	}, nil
}
