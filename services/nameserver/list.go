package nameserver

import (
	"context"
	"encoding/json"

	"github.com/containerd/containerd/errdefs"
	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
)

func (s *service) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	kvs, err := c.Datastore().Search(dsNameserverBucketName, "*")
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
	}

	var records []*api.Record
	for _, kv := range kvs {
		var r []*api.Record
		if err := json.Unmarshal(kv.Value, &r); err != nil {
			return nil, err
		}
		records = append(records, r...)
	}

	return &api.ListResponse{
		Records: records,
	}, nil
}
