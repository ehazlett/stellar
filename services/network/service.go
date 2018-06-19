package network

import (
	"net"

	"github.com/ehazlett/element"
	api "github.com/ehazlett/stellar/api/services/network/v1"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.network.v1"
	// TODO: make configurable
	// default max subnets (max nodes)
	maxSubnets = 1024
)

type service struct {
	network *net.IPNet
	agent   *element.Agent
}

func New(agent *element.Agent, network *net.IPNet) (*service, error) {
	return &service{
		network: network,
		agent:   agent,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNetworkServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}
