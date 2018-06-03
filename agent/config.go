package agent

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/memberlist"
)

// ConnectionType defines the type of connection for the agent to use (wan, lan, local)
type ConnectionType string

const (
	// LAN is a local area network connection intended for high speed, low latency networks
	LAN ConnectionType = "lan"
	// WAN is a wide area connection intended for long distance remote links
	WAN ConnectionType = "wan"
	// Local is a local connection intended for high speed local development links
	Local ConnectionType = "local"
)

// Config is the agent config
type Config struct {
	// NodeName is the name of the node.  Each node must have a unique name in the cluster.
	NodeName string
	// AgentAddr is the address on which the agent will serve the GRPC services
	AgentAddr string
	// AgentPort is the port on which the agent will serve the GRPC services
	AgentPort int
	// ConnectionType is the connection type the agent will use
	ConnectionType string
	// BindAddr is the cluster bind address
	BindAddr string
	// BindPort is the cluster bind port
	BindPort int
	// AdvertiseAddr is the cluster address that will be used for membership communication
	AdvertiseAddr string
	// AdvertisePort is the cluster port that will be used for membership communication
	AdvertisePort int
	// Peers is a local cache of peer members
	Peers []string
}

func setupMemberlistConfig(cfg *Config, peerUpdateChan chan bool, nodeEventChan chan *NodeEvent) (*memberlist.Config, error) {
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
	mc.Delegate = NewAgentDelegate(cfg.NodeName, fmt.Sprintf("%s:%d", cfg.AgentAddr, cfg.AgentPort), peerUpdateChan, nodeEventChan)
	mc.Events = NewEventHandler(nodeEventChan)

	// disable logging for memberlist
	// TODO: enable if debug
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
