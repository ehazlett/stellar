package events

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/events/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Publish(ctx context.Context, req *api.Message) (*ptypes.Empty, error) {
	c, err := s.natsClient()
	if err != nil {
		return empty, err
	}
	defer c.Close()

	data, err := req.Data.Marshal()
	if err != nil {
		return empty, err
	}

	if err := c.Publish(req.Subject, data); err != nil {
		return empty, err
	}

	return empty, nil
}
