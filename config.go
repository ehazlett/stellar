package stellar

import (
	"net"
	"time"

	"github.com/ehazlett/element"
)

type Config struct {
	AgentConfig              *element.Config
	ContainerdAddr           string
	Namespace                string
	Subnet                   *net.IPNet
	DataDir                  string
	StateDir                 string
	Bridge                   string
	UpstreamDNSAddr          string
	ProxyHTTPPort            int
	ProxyHealthcheckInterval time.Duration
}
