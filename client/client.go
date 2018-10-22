package client

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/ehazlett/stellar"
	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	eventsapi "github.com/ehazlett/stellar/api/services/events/v1"
	healthapi "github.com/ehazlett/stellar/api/services/health/v1"
	nameserverapi "github.com/ehazlett/stellar/api/services/nameserver/v1"
	networkapi "github.com/ehazlett/stellar/api/services/network/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
	proxyapi "github.com/ehazlett/stellar/api/services/proxy/v1"
	versionapi "github.com/ehazlett/stellar/api/services/version/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	eventsService      eventsapi.EventsClient
}

func NewClient(addr string, opts ...grpc.DialOption) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if len(opts) == 0 {
		opts = []grpc.DialOption{
			grpc.WithInsecure(),
		}
	}

	opts = append(opts, grpc.WithWaitForHandshake())
	c, err := grpc.DialContext(ctx,
		addr,
		opts...,
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
		eventsService:      eventsapi.NewEventsClient(c),
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

func (c *Client) Events() *events {
	return &events{
		client: c.eventsService,
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

func (c *Client) EventsService() eventsapi.EventsClient {
	return c.eventsService
}

func DialOptionsFromConfig(cfg *stellar.Config) ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{}
	if cfg.TLSClientCertificate != "" {
		logrus.WithField("cert", cfg.TLSClientCertificate)
		var creds credentials.TransportCredentials
		if cfg.TLSClientKey != "" {
			logrus.WithField("key", cfg.TLSClientKey)
			cert, err := tls.LoadX509KeyPair(cfg.TLSClientCertificate, cfg.TLSClientKey)
			if err != nil {
				return nil, err
			}
			creds = credentials.NewTLS(&tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
			})
		} else {
			c, err := credentials.NewClientTLSFromFile(cfg.TLSClientCertificate, "")
			if err != nil {
				return nil, err
			}
			creds = c
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	return opts, nil
}
