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
	proxyapi "github.com/ehazlett/stellar/api/services/proxy/v1"
	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
	schedulerapi "github.com/ehazlett/stellar/api/services/scheduler/v1"
	versionapi "github.com/ehazlett/stellar/api/services/version/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	empty = &ptypes.Empty{}
)

// Client provides access to Stellar cluster services
type Client struct {
	conn               *grpc.ClientConn
	versionService     versionapi.VersionClient
	healthService      healthapi.HealthClient
	nodeService        runtimeapi.NodeClient
	clusterService     clusterapi.ClusterClient
	datastoreService   datastoreapi.DatastoreClient
	networkService     networkapi.NetworkClient
	applicationService applicationapi.ApplicationClient
	nameserverService  nameserverapi.NameserverClient
	proxyService       proxyapi.ProxyClient
	eventsService      eventsapi.EventsClient
	schedulerService   schedulerapi.SchedulerClient
}

// NewClient returns a new client configured with the specified Stellar GRPC address and dial options
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
		nodeService:        runtimeapi.NewNodeClient(c),
		clusterService:     clusterapi.NewClusterClient(c),
		datastoreService:   datastoreapi.NewDatastoreClient(c),
		networkService:     networkapi.NewNetworkClient(c),
		applicationService: applicationapi.NewApplicationClient(c),
		nameserverService:  nameserverapi.NewNameserverClient(c),
		proxyService:       proxyapi.NewProxyClient(c),
		eventsService:      eventsapi.NewEventsClient(c),
		schedulerService:   schedulerapi.NewSchedulerClient(c),
	}

	return client, nil
}

// Conn returns the current configured client connection
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Close closes the underlying GRPC client
func (c *Client) Close() error {
	return c.conn.Close()
}

// Application is a helper to return the application service client
func (c *Client) Application() *application {
	return &application{
		client: c.applicationService,
	}
}

// Node is a helper to return the node service client
func (c *Client) Node() *node {
	return &node{
		client: c.nodeService,
	}
}

// Cluster is a helper to return the cluster service client
func (c *Client) Cluster() *cluster {
	return &cluster{
		client: c.clusterService,
	}
}

// Datastore is a helper to return the datastore service client
func (c *Client) Datastore() *datastore {
	return &datastore{
		client: c.datastoreService,
	}
}

// Network is a helper to return the network service client
func (c *Client) Network() *network {
	return &network{
		client: c.networkService,
	}
}

// Events is a helper to return the events service client
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

// Proxy is a helper to return the proxy service client
func (c *Client) Proxy() *proxy {
	return &proxy{
		client: c.proxyService,
	}
}

// Scheduler is a helper to return the scheduler service client
func (c *Client) Scheduler() *scheduler {
	return &scheduler{
		client: c.schedulerService,
	}
}

// Version is a helper to return the version service client
func (c *Client) Version() *version {
	return &version{
		client: c.versionService,
	}
}

// Health is a helper to return the health service client
func (c *Client) Health() *health {
	return &health{
		client: c.healthService,
	}
}

// VersionService returns the direct version service api client for advanced usage
func (c *Client) VersionService() versionapi.VersionClient {
	return c.versionService
}

// HealthService returns the direct health service api client for advanced usage
func (c *Client) HealthService() healthapi.HealthClient {
	return c.healthService
}

// NodeService returns the direct node service api client for advanced usage
func (c *Client) NodeService() runtimeapi.NodeClient {
	return c.nodeService
}

// ClusterService returns the direct cluster service api client for advanced usage
func (c *Client) ClusterService() clusterapi.ClusterClient {
	return c.clusterService
}

// DatastoreService returns the direct datastore service api client for advanced usage
func (c *Client) DatastoreService() datastoreapi.DatastoreClient {
	return c.datastoreService
}

// NetworkService returns the direct network service api client for advanced usage
func (c *Client) NetworkService() networkapi.NetworkClient {
	return c.networkService
}

// NameserverService returns the direct nameserver service api client for advanced usage
func (c *Client) NameserverService() nameserverapi.NameserverClient {
	return c.nameserverService
}

// EventsService returns the direct events service api client for advanced usage
func (c *Client) EventsService() eventsapi.EventsClient {
	return c.eventsService
}

// DialOptionsFromConfig returns dial options configured from a Stellar config
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
