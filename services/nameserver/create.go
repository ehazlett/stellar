package nameserver

import (
	"context"
	"encoding/json"

	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Create(ctx context.Context, req *api.CreateRequest) (*ptypes.Empty, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	data, err := json.Marshal(req.Records)
	if err != nil {
		return nil, err
	}

	if err := c.Datastore().Set(dsNameserverBucketName, req.Name, data, true); err != nil {
		return empty, err
	}

	return empty, nil
}
