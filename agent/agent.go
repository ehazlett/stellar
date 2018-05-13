package agent

import (
	"errors"
	"time"

	"github.com/hashicorp/memberlist"
)

const (
	nodeHeartbeatInterval = time.Second * 10
	nodeReconcileTimeout  = nodeHeartbeatInterval * 2
	nodeUpdateTimeout     = nodeHeartbeatInterval / 2
)

var (
	ErrUnknownConnectionType = errors.New("unknown connection type")
)

type PeerAgent struct {
	Name    string
	Addr    string
	Updated time.Time
}

type Agent struct {
	config         *Config
	members        *memberlist.Memberlist
	peerUpdateChan chan bool
}

func NewAgent(cfg *Config) (*Agent, error) {
	updateCh := make(chan bool)
	mc, err := setupMemberlistConfig(cfg, updateCh)
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
	}, nil
}
