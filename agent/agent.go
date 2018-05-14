package agent

import (
	"errors"
	"time"

	"github.com/hashicorp/memberlist"
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
}

func NewAgent(cfg *Config) (*Agent, error) {
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

	return &Agent{
		config:         cfg,
		members:        ml,
		peerUpdateChan: updateCh,
		nodeEventChan:  nodeEventCh,
	}, nil
}
