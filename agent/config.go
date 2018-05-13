package agent

import (
	"io/ioutil"
	"log"

	"github.com/hashicorp/memberlist"
)

type ConnectionType string

const (
	LAN   ConnectionType = "lan"
	WAN   ConnectionType = "wan"
	Local ConnectionType = "local"
)

type Config struct {
	NodeName       string
	Namespace      string
	AgentAddr      string
	ContainerdAddr string
	ConnectionType string
	BindAddr       string
	BindPort       int
	AdvertiseAddr  string
	AdvertisePort  int
	Peers          []string
}

func setupMemberlistConfig(cfg *Config, peerUpdateChan chan bool) (*memberlist.Config, error) {
	var mc *memberlist.Config
	switch cfg.ConnectionType {
	case string(Local):
		mc = memberlist.DefaultLocalConfig()
	case string(WAN):
		mc = memberlist.DefaultWANConfig()
	case string(LAN):
		mc = memberlist.DefaultLANConfig()
	default:
		return nil, ErrUnknownConnectionType
	}

	mc.Name = cfg.NodeName
	mc.Delegate = NewAgentDelegate(cfg.NodeName, cfg.AgentAddr, peerUpdateChan)

	// disable logging for memberlist
	mc.Logger = log.New(ioutil.Discard, "", 0)

	// ml overrides for connection
	if v := cfg.BindAddr; v != "" {
		mc.BindAddr = v
	}
	if v := cfg.BindPort; v > 0 {
		mc.BindPort = v
	}
	if v := cfg.AdvertiseAddr; v != "" {
		mc.AdvertiseAddr = v
	}
	if v := cfg.AdvertisePort; v > 0 {
		mc.AdvertisePort = v
	}

	return mc, nil
}
