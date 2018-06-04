package stellar

import (
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	healthapi "github.com/ehazlett/stellar/api/services/health/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
	versionapi "github.com/ehazlett/stellar/api/services/version/v1"
	"google.golang.org/grpc"
)

type Client struct {
	conn           *grpc.ClientConn
	VersionService versionapi.VersionClient
	HealthService  healthapi.HealthClient
	NodeService    nodeapi.NodeClient
	ClusterService clusterapi.ClusterClient
}

func NewClient(addr string) (*Client, error) {
	c, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn: c,
	}
	client.VersionService = versionapi.NewVersionClient(c)
	client.HealthService = healthapi.NewHealthClient(c)
	client.NodeService = nodeapi.NewNodeClient(c)
	client.ClusterService = clusterapi.NewClusterClient(c)

	return client, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
