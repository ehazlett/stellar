package client

import (
	"context"

	proxyapi "github.com/ehazlett/stellar/api/services/proxy/v1"
)

type proxy struct {
	client proxyapi.ProxyClient
}

func (p *proxy) Reload() error {
	ctx := context.Background()
	if _, err := p.client.Reload(ctx, &proxyapi.ReloadRequest{}); err != nil {
		return err
	}
	return nil
}

func (p *proxy) Backends() ([]*proxyapi.Backend, error) {
	ctx := context.Background()
	resp, err := p.client.Backends(ctx, &proxyapi.BackendRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Backends, nil
}
