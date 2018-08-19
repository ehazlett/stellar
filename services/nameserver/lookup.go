package nameserver

import (
	"context"
	"encoding/json"

	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
	"github.com/sirupsen/logrus"
)

func (s *service) Lookup(ctx context.Context, req *api.LookupRequest) (*api.LookupResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	logrus.Debugf("lookup query=%s", req.Query)
	kvs, err := c.Datastore().Search(dsNameserverBucketName, req.Query)
	if err != nil {
		return nil, err
	}

	var records []*api.Record
	for _, kv := range kvs {
		var r []*api.Record
		if err := json.Unmarshal(kv.Value, &r); err != nil {
			return nil, err
		}

		records = append(records, r...)
	}

	return &api.LookupResponse{
		Name:    req.Query,
		Records: records,
	}, nil
}
