package element

import (
	"errors"
	"time"

	"github.com/hashicorp/memberlist"
	"google.golang.org/grpc"
)

const (
	defaultInterval      = time.Second * 10
	nodeReconcileTimeout = defaultInterval * 3
	nodeUpdateTimeout    = defaultInterval / 2
)

var (
	ErrUnknownConnectionType = errors.New("unknown connection type")
)

// Agent represents the node agent
type Agent struct {
	config             *Config
	members            *memberlist.Memberlist
	peerUpdateChan     chan bool
	nodeEventChan      chan *NodeEvent
	grpcServer         *grpc.Server
	registeredServices map[string]struct{}
	memberConfig       *memberlist.Config
}

// NewAgent returns a new node agent
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
	grpcServer := grpc.NewServer()
	return &Agent{
		config:         cfg,
		members:        ml,
		peerUpdateChan: updateCh,
		nodeEventChan:  nodeEventCh,
		grpcServer:     grpcServer,
		memberConfig:   mc,
	}, nil
}

// SyncInterval returns the cluster sync interval
func (a *Agent) SyncInterval() time.Duration {
	return a.memberConfig.PushPullInterval
}

// Subscribe subscribes to the node event channel
func (a *Agent) Subscribe(ch chan *NodeEvent) {
	go func() {
		for {
			select {
			case evt := <-a.nodeEventChan:
				ch <- evt
			}
		}
	}()
}
