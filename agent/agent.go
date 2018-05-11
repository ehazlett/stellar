package agent

import (
	"errors"

	"github.com/hashicorp/memberlist"
)

var (
	ErrUnknownConnectionType = errors.New("unknown connection type")
)

type Agent struct {
	config  *Config
	members *memberlist.Memberlist
}

func NewAgent(cfg *Config) (*Agent, error) {
	mc, err := setupMemberlistConfig(cfg)
	if err != nil {
		return nil, err
	}

	ml, err := memberlist.Create(mc)
	if err != nil {
		return nil, err
	}

	return &Agent{
		config:  cfg,
		members: ml,
	}, nil
}
