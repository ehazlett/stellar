package application

import (
	"github.com/containerd/typeurl"
	"github.com/ehazlett/stellar/events"
)

func init() {
	typeurl.Register(&UpdateEvent{}, serviceID+"/UpdateEvent")
}

// UpdateEvent is the event published when an application is updated
type UpdateEvent struct {
	Application string
	Action      string
}

func (s *service) publish(v interface{}) error {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := events.PublishEvent(c, serviceID, v); err != nil {
		return err
	}

	return nil
}
