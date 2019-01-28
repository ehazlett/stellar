package nameserver

import (
	"github.com/containerd/typeurl"
	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
	"github.com/ehazlett/stellar/events"
)

func init() {
	typeurl.Register(&CreateEvent{}, serviceID+"/CreateEvent")
	typeurl.Register(&DeleteEvent{}, serviceID+"/DeleteEvent")
}

// CreateEvent is the event published when a record is created
type CreateEvent struct {
	Name    string
	Records []*api.Record
}

// DeleteEvent is the event published when a record is created
type DeleteEvent struct {
	Type api.RecordType
	Name string
}

func (s *service) publish(v interface{}) error {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := events.PublishEvent(c, s.ID(), v); err != nil {
		return err
	}

	return nil
}
