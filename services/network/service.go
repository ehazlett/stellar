package network

import (
	"net"

	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	api "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.network.v1"
	// TODO: make configurable
	// default max subnets (max nodes)
	maxSubnets          = 1024
	dsNetworkBucketName = "stellar." + stellar.APIVersion + ".services.network"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	network *net.IPNet
	agent   *element.Agent
	ds      datastoreapi.DatastoreServer
}

func New(ds datastoreapi.DatastoreServer, agent *element.Agent, network *net.IPNet) (*service, error) {
	return &service{
		network: network,
		agent:   agent,
		ds:      ds,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNetworkServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Start() error {
	return nil
}
