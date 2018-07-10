package application

import (
	"github.com/containerd/containerd"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	networkapi "github.com/ehazlett/stellar/api/services/network/v1"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.application.v1"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr string
	namespace      string
	networkService networkapi.NetworkServer
}

func New(containerdAddr, namespace string, svc networkapi.NetworkServer) (*service, error) {
	return &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
		networkService: svc,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterApplicationServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) containerd() (*containerd.Client, error) {
	return containerd.New(s.containerdAddr, containerd.WithDefaultNamespace(s.namespace))
}
