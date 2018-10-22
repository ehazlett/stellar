package client

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/events/v1"
	ptypes "github.com/gogo/protobuf/types"
)

type events struct {
	client api.EventsClient
}

func (e *events) ID() (string, error) {
	ctx := context.Background()
	resp, err := e.client.Info(ctx, &api.InfoRequest{})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (e *events) Endpoint(req *api.EndpointRequest) (*api.Endpoint, error) {
	ctx := context.Background()
	resp, err := e.client.Endpoint(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Endpoint, nil
}

func (e *events) Publish(subject string, v *ptypes.Any) error {
	ctx := context.Background()
	if _, err := e.client.Publish(ctx, &api.Message{
		Subject: subject,
		Data:    v,
	}); err != nil {
		return err
	}

	return nil
}
