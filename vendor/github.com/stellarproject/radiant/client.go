package radiant

import (
	"context"
	"net/url"
	"time"

	api "github.com/stellarproject/radiant/api/v1"
	"google.golang.org/grpc"
)

var (
	defaultTimeout = time.Second * 30
)

type Client struct {
	conn         *grpc.ClientConn
	proxyService api.ProxyClient
}

func NewClient(addr string) (*Client, error) {
	grpcAddr, err := getGRPCAddr(addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	c, err := grpc.DialContext(ctx,
		grpcAddr,
		grpc.WithInsecure(),
		grpc.WithWaitForHandshake(),
	)
	if err != nil {
		return nil, err

	}
	proxyService := api.NewProxyClient(c)

	return &Client{
		conn:         c,
		proxyService: proxyService,
	}, nil
}

func (c *Client) AddServer(host string, opts ...AddOpts) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	srv := &api.Server{
		Host: host,
		Path: "/",
	}

	for _, o := range opts {
		if err := o(ctx, srv); err != nil {
			return err
		}
	}

	if _, err := c.proxyService.AddServer(ctx, &api.AddServerRequest{
		Server: srv,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveServer(host string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := c.proxyService.RemoveServer(ctx, &api.RemoveServerRequest{
		Host: host,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Client) Servers() ([]*api.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := c.proxyService.Servers(ctx, &api.ServersRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Servers, nil
}

func (c *Client) Reload() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := c.proxyService.Reload(ctx, &api.ReloadRequest{}); err != nil {
		return err
	}
	return nil
}

func (c *Client) Config() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := c.proxyService.Config(ctx, &api.ConfigRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func getGRPCAddr(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	switch u.Scheme {
	case "unix":
		return "passthrough:///" + uri, nil
	default:
		return u.Host, nil
	}
}
