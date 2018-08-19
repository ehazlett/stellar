package nameserver

import (
	"github.com/containerd/containerd"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"

	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
)

const (
	serviceID              = "stellar.services.nameserver.v1"
	dsNameserverBucketName = "stellar." + stellar.APIVersion + ".services.nameserver"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr string
	namespace      string
	bridge         string
	agent          *element.Agent
}

func New(containerdAddr, namespace, bridge string, agent *element.Agent) (*service, error) {
	srv := &service{
		containerdAddr: containerdAddr,
		namespace:      namespace,
		bridge:         bridge,
		agent:          agent,
	}

	return srv, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterNameserverServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Start() error {
	return s.startDNSServer()
}

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) client() (*client.Client, error) {
	peer, err := s.agent.LocalNode()
	if err != nil {
		return nil, err
	}
	return client.NewClient(peer.Addr)
}
