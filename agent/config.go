package agent

import (
	"io/ioutil"
	"log"

	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
)

type ConnectionType string

const (
	LAN   ConnectionType = "lan"
	WAN   ConnectionType = "wan"
	Local ConnectionType = "local"
)

type Config struct {
	NodeName       string
	AgentAddr      string
	ConnectionType string
	BindAddr       string
	BindPort       int
	AdvertiseAddr  string
	AdvertisePort  int
}

type agentDelegate struct {
	addr string
}

func (d *agentDelegate) NodeMeta(limit int) []byte {
	return []byte(d.addr)
}

func (d *agentDelegate) NotifyMsg(buf []byte) {
	// this can be used to receive messages sent (i.e. SendReliable)
}

// GetBroadcasts is called when user messages can be broadcast
func (d *agentDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

func (d *agentDelegate) LocalState(join bool) []byte {
	// TODO: serialize a known struct for node info to connect with GRPC
	return []byte(d.addr)
}

func (d *agentDelegate) MergeRemoteState(buf []byte, join bool) {
	// TODO: handle receiving remote state and connect to GRPC
	// TODO: check if peer GRPC is connected (from health service) and connect if not
	logrus.Debugf("merge: %s", string(buf))
}

func setupMemberlistConfig(cfg *Config) (*memberlist.Config, error) {
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
	mc.Delegate = &agentDelegate{addr: cfg.AgentAddr}

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
