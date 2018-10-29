package events

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/events/v1"
	"github.com/gogo/protobuf/proto"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Subscribe(req *api.SubscribeRequest, srv api.Events_SubscribeServer) error {
	c, err := s.natsClient()
	if err != nil {
		return err
	}
	defer c.Close()

	sub, err := c.SubscribeSync(req.Subject)
	if err != nil {
		return err
	}
	for {
		m, err := sub.NextMsgWithContext(context.Background())
		if err != nil {
			return err
		}

		any := &ptypes.Any{}
		if err := proto.Unmarshal(m.Data, any); err != nil {
			return err
		}

		if err := srv.Send(&api.Message{
			Subject: m.Subject,
			Data:    any,
		}); err != nil {
			logrus.WithError(err).Error("error sending message to subscriber")
		}
	}

	return nil
}
