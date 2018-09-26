package stellar

import (
	"encoding/json"
	"net"
	"time"

	"github.com/ehazlett/element"
)

// Config is the configuration used for the stellar server
// Note: in order to make user configuration from file a better user experience
// there is a custom marshal/unmarshal below.  Those must be updated if fields are
// added or removed from `Config`.
type Config struct {
	// AgentConfig is the element config for the server
	AgentConfig *element.Config `json:"-"`
	// ContainerdAddr is the address to the containerd socket
	ContainerdAddr string
	// Namespace is the containerd namespace to manage
	Namespace string
	// Subnet is the subnet to use for stellar networking
	Subnet *net.IPNet
	// DataDir is the directory used to store stellar data
	DataDir string
	// State is the directory to store run state
	StateDir string
	// Bridge is the name of the bridge for networking
	Bridge string
	// UpstreamDNSAddr is the address to use for external queries
	UpstreamDNSAddr string
	// ProxyHTTPPort is the http port to use for the proxy service
	ProxyHTTPPort int
	// ProxyHTTPSPort is the https port to use for the proxy service
	ProxyHTTPSPort int
	// ProxyTLSEmail is the email address used when requesting letsencrypt certs
	ProxyTLSEmail string
	// ProxyHealthcheckInterval is the interval used by the proxy service to check upstreams
	ProxyHealthcheckInterval time.Duration
	// GatewayAddr is the http addr to use for the http/json API
	GatewayAddr string
	// GatewayPort is the http port to use for the http/json API
	GatewayPort int
}

// MarshalJSON is a custom json marshaller for better ux
func (c *Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	type Agent element.Config
	return json.Marshal(&struct {
		*Agent
		*Alias
		Peers                    []string
		Subnet                   string
		ProxyHealthcheckInterval string
	}{
		Alias:  (*Alias)(c),
		Agent:  (*Agent)(c.AgentConfig),
		Peers:  c.AgentConfig.Peers,
		Subnet: c.Subnet.String(),
		ProxyHealthcheckInterval: c.ProxyHealthcheckInterval.String(),
	})
}

// UnmarshalJSON is a custom json unmarshaller for better ux
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	type Agent element.Config
	tmp := &struct {
		*Alias
		*Agent
		Subnet                   string
		ProxyHealthcheckInterval string
	}{
		Alias: (*Alias)(c),
		Agent: (*Agent)(c.AgentConfig),
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	c.AgentConfig = &element.Config{
		NodeName:       tmp.NodeName,
		AgentAddr:      tmp.AgentAddr,
		AgentPort:      tmp.AgentPort,
		ConnectionType: tmp.ConnectionType,
		BindAddr:       tmp.BindAddr,
		BindPort:       tmp.BindPort,
		AdvertiseAddr:  tmp.AdvertiseAddr,
		AdvertisePort:  tmp.AdvertisePort,
		Peers:          tmp.Peers,
	}

	_, subnet, err := net.ParseCIDR(tmp.Subnet)
	if err != nil {
		return err
	}
	c.Subnet = subnet

	d, err := time.ParseDuration(tmp.ProxyHealthcheckInterval)
	if err != nil {
		return err
	}
	c.ProxyHealthcheckInterval = d

	return nil
}
