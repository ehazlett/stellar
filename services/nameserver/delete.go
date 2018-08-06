package nameserver

import (
	"context"
	"encoding/json"

	"github.com/containerd/containerd/errdefs"
	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Delete(ctx context.Context, req *api.DeleteRequest) (*ptypes.Empty, error) {
	c, err := s.client()
	if err != nil {
		return empty, err
	}
	defer c.Close()

	kvs, err := c.Datastore().Search(dsNameserverBucketName, req.Name)
	if err != nil {
		err = errdefs.FromGRPC(err)
		if !errdefs.IsNotFound(err) {
			return nil, err
		}
	}

	for _, kv := range kvs {
		var records []*api.Record
		if err := json.Unmarshal(kv.Value, &records); err != nil {
			return nil, err
		}

		for _, record := range records {
			if record.Type == req.Type {
				logrus.Debugf("nameserver: deleting record %s %s", record.Type, record.Name)
				if err := c.Datastore().Delete(dsNameserverBucketName, req.Name, true); err != nil {
					return empty, err
				}
			}
		}
	}

	return empty, nil
}
