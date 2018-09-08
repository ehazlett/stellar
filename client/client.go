package client

import (
	"context"
	"time"

	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	healthapi "github.com/ehazlett/stellar/api/services/health/v1"
	nameserverapi "github.com/ehazlett/stellar/api/services/nameserver/v1"
	networkapi "github.com/ehazlett/stellar/api/services/network/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
	proxyapi "github.com/ehazlett/stellar/api/services/proxy/v1"
	versionapi "github.com/ehazlett/stellar/api/services/version/v1"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

var (
	empty = &ptypes.Empty{}
)

type Client struct {
	conn               *grpc.ClientConn
	versionService     versionapi.VersionClient
	healthService      healthapi.HealthClient
	nodeService        nodeapi.NodeClient
	clusterService     clusterapi.ClusterClient
	datastoreService   datastoreapi.DatastoreClient
	networkService     networkapi.NetworkClient
	applicationService applicationapi.ApplicationClient
	nameserverService  nameserverapi.NameserverClient
	proxyService       proxyapi.ProxyClient
}

func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	c, err := grpc.DialContext(ctx,
		addr,
		grpc.WithInsecure(),
		grpc.WithWaitForHandshake(),
	)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:               c,
		versionService:     versionapi.NewVersionClient(c),
		healthService:      healthapi.NewHealthClient(c),
		nodeService:        nodeapi.NewNodeClient(c),
		clusterService:     clusterapi.NewClusterClient(c),
		datastoreService:   datastoreapi.NewDatastoreClient(c),
		networkService:     networkapi.NewNetworkClient(c),
		applicationService: applicationapi.NewApplicationClient(c),
		nameserverService:  nameserverapi.NewNameserverClient(c),
		proxyService:       proxyapi.NewProxyClient(c),
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Application() *application {
	return &application{
		client: c.applicationService,
	}
}

func (c *Client) Node() *node {
	return &node{
		client: c.nodeService,
	}
}

func (c *Client) Cluster() *cluster {
	return &cluster{
		client: c.clusterService,
	}
}

func (c *Client) Datastore() *datastore {
	return &datastore{
		client: c.datastoreService,
	}
}

func (c *Client) Network() *network {
	return &network{
		client: c.networkService,
	}
}

func (c *Client) Nameserver() *nameserver {
	return &nameserver{
		client: c.nameserverService,
	}
}

func (c *Client) Proxy() *proxy {
	return &proxy{
		client: c.proxyService,
	}
}

func (c *Client) Version() *version {
	return &version{
		client: c.versionService,
	}
}

func (c *Client) Health() *health {
	return &health{
		client: c.healthService,
	}
}

func (c *Client) VersionService() versionapi.VersionClient {
	return c.versionService
}

func (c *Client) HealthService() healthapi.HealthClient {
	return c.healthService
}

func (c *Client) NodeService() nodeapi.NodeClient {
	return c.nodeService
}

func (c *Client) ClusterService() clusterapi.ClusterClient {
	return c.clusterService
}

func (c *Client) DatastoreService() datastoreapi.DatastoreClient {
	return c.datastoreService
}

func (c *Client) NetworkService() networkapi.NetworkClient {
	return c.networkService
}

func (c *Client) NameserverService() nameserverapi.NameserverClient {
	return c.nameserverService
}
