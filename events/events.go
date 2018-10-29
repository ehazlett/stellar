package events

import (
	"github.com/containerd/typeurl"
	eventsapi "github.com/ehazlett/stellar/api/services/events/v1"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
)

// MarshalEvent marshals and event for transport
func MarshalEvent(v interface{}) (*ptypes.Any, error) {
	any, err := typeurl.MarshalAny(v)
	if err != nil {
		return nil, err
	}
	return any, err
}

// UnmarshalEvent unmarshals an event from transport
func UnmarshalEvent(m *eventsapi.Message) (interface{}, error) {
	v, err := typeurl.UnmarshalAny(m.Data)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// IsEvent returns true if the type of the message data is the same as v
func IsEvent(m *eventsapi.Message, v interface{}) bool {
	return typeurl.Is(m.Data, v)
}

// PublishEvent publishes the event to the subject using the specified client
func PublishEvent(c *client.Client, subject string, v interface{}) error {
	any, err := MarshalEvent(v)
	if err != nil {
		return err
	}

	if err := c.Events().Publish(subject, any); err != nil {
		return err
	}

	return nil
}
