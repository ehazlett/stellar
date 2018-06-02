package agent

import (
	"errors"
	"fmt"
	"time"

	"github.com/ehazlett/element/services"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
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
	registered := make(map[string]struct{}, len(svcs))
	for _, svc := range svcs {
		id := svc.ID()
		if _, exists := registered[id]; exists {
			return nil, fmt.Errorf("service %s already registered", id)
		}
		svc.Register(grpcServer)
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Info("registered service")
		registered[id] = struct{}{}
	}

	return &Agent{
		config:         cfg,
		members:        ml,
		peerUpdateChan: updateCh,
		nodeEventChan:  nodeEventCh,
		grpcServer:     grpcServer,
	}, nil
}
