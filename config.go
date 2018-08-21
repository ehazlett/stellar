package stellar

import (
	"net"

	"github.com/ehazlett/element"
)

type Config struct {
	AgentConfig    *element.Config
	ContainerdAddr string
	Namespace      string
	Subnet         *net.IPNet
	DataDir        string
	StateDir       string
	Bridge         string
	ProxyHTTPPort  int
}
