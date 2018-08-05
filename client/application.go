package client

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/application/v1"
)

type application struct {
	client api.ApplicationClient
}

func (a *application) Create(req *api.CreateRequest) error {
	ctx := context.Background()
	if _, err := a.client.Create(ctx, req); err != nil {
		return err
	}

	return nil
}

func (a *application) Delete(name string) error {
	ctx := context.Background()
	if _, err := a.client.Delete(ctx, &api.DeleteRequest{
		Name: name,
	}); err != nil {
		return err
	}

	return nil
}

func (a *application) List() ([]*api.App, error) {
	ctx := context.Background()
	resp, err := a.client.List(ctx, &api.ListRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Applications, nil
}

func (a *application) Get(name string) (*api.App, error) {
	ctx := context.Background()
	resp, err := a.client.Get(ctx, &api.GetRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return resp.Application, nil
}
