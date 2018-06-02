package agent

import (
	"errors"
	"time"

	"github.com/ehazlett/element/services"
	"github.com/hashicorp/memberlist"
	"google.golang.org/grpc"
)

const (
	nodeHeartbeatInterval = time.Second * 10
	nodeReconcileTimeout  = nodeHeartbeatInterval * 3
	nodeUpdateTimeout     = nodeHeartbeatInterval / 2
)

var (
	ErrUnknownConnectionType = errors.New("unknown connection type")
)

type Agent struct {
	config         *Config
	members        *memberlist.Memberlist
	peerUpdateChan chan bool
	nodeEventChan  chan *NodeEvent
	grpcServer     *grpc.Server
}

func NewAgent(cfg *Config, svcs ...services.Service) (*Agent, error) {
	updateCh := make(chan bool)
	nodeEventCh := make(chan *NodeEvent)
	mc, err := setupMemberlistConfig(cfg, updateCh, nodeEventCh)
	if err != nil {
		return nil, err
	}

	ml, err := memberlist.Create(mc)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()
	for _, svc := range svcs {
		svc.Register(grpcServer)
	}

	return &Agent{
		config:         cfg,
		members:        ml,
		peerUpdateChan: updateCh,
		nodeEventChan:  nodeEventCh,
		grpcServer:     grpcServer,
	}, nil
}
