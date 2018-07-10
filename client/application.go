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
