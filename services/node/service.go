package node

import (
	"github.com/containerd/containerd"
	"github.com/ehazlett/stellar"
	networkapi "github.com/ehazlett/stellar/api/services/network/v1"
	api "github.com/ehazlett/stellar/api/services/node/v1"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.node.v1"
)

type service struct {
	containerdAddr string
	namespace      string
	bridge         string
	networkService networkapi.NetworkServer
}

func New(containerdAddr, namespace, bridge string, svc networkapi.NetworkServer) (*service, error) {
	return &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
		bridge:         bridge,
		networkService: svc,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNodeServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}
