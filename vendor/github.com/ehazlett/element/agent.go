package element

import (
	"errors"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
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
	*subscribers

	config             *Config
	members            *memberlist.Memberlist
	peerUpdateChan     chan bool
	nodeEventChan      chan *NodeEvent
	registeredServices map[string]struct{}
	memberConfig       *memberlist.Config
	state              *State
}

// NewAgent returns a new node agent
func NewAgent(info *Peer, cfg *Config) (*Agent, error) {
	var (
		updateCh    = make(chan bool, 64)
		nodeEventCh = make(chan *NodeEvent, 64)
	)
	a := &Agent{
		subscribers:    newSubscribers(),
		config:         cfg,
		peerUpdateChan: updateCh,
		nodeEventChan:  nodeEventCh,
		state: &State{
			Self:  info,
			Peers: make(map[string]*Peer),
		},
	}
	mc, err := cfg.memberListConfig(a)
	if err != nil {
		return nil, err
	}
	ml, err := memberlist.Create(mc)
	if err != nil {
		return nil, err
	}
	a.members = ml
	a.memberConfig = mc

	return a, nil
}

// SyncInterval returns the cluster sync interval
func (a *Agent) SyncInterval() time.Duration {
	return a.memberConfig.PushPullInterval
}

func newSubscribers() *subscribers {
	return &subscribers{
		subs: make(map[chan *NodeEvent]struct{}),
	}
}

type subscribers struct {
	mu sync.Mutex

	subs map[chan *NodeEvent]struct{}
}

// Subscribe subscribes to the node event channel
func (s *subscribers) Subscribe() chan *NodeEvent {
	ch := make(chan *NodeEvent, 64)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

// Unsubscribe removes the channel from node events
func (s *subscribers) Unsubscribe(ch chan *NodeEvent) {
	s.mu.Lock()
	delete(s.subs, ch)
	s.mu.Unlock()
}

func (s *subscribers) send(e *NodeEvent) {
	s.mu.Lock()
	for ch := range s.subs {
		// non-blocking send
		select {
		case ch <- e:
		default:
		}
	}
	s.mu.Unlock()
}
