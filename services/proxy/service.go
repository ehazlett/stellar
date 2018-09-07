package proxy

import (
	"fmt"
	"time"

	"github.com/containerd/containerd"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/version"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/mholt/caddy"
	_ "github.com/mholt/caddy/caddyhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.proxy.v1"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr           string
	namespace                string
	agent                    *element.Agent
	proxyHTTPPort            int
	proxyHTTPSPort           int
	proxyHealthcheckInterval time.Duration
	server                   *caddy.Instance
	errCh                    chan error
}

func init() {
	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(caddyLoader))
}

func New(cfg *stellar.Config, agent *element.Agent) (*service, error) {
	errCh := make(chan error)
	go func() {
		for {
			err := <-errCh
			logrus.Errorf("proxy: %s", err)
		}
	}()

	caddy.AppName = "stellar.proxy"
	caddy.AppVersion = version.Version

	return &service{
		containerdAddr:           cfg.ContainerdAddr,
		namespace:                cfg.Namespace,
		proxyHTTPPort:            cfg.ProxyHTTPPort,
		proxyHealthcheckInterval: cfg.ProxyHealthcheckInterval,
		errCh: errCh,
		agent: agent,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterProxyServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Start() error {
	caddyfile, err := caddy.LoadCaddyfile("http")
	if err != nil {
		return err
	}

	s.instance, err = caddy.Start(caddyfile)
	if err != nil {
		return err
	}
	t := time.NewTicker(5 * time.Second)
	go func() {
		for range t.C {
			//c, err := s.client()
			//if err != nil {
			//	logrus.Errorf("proxy: %s", err)
			//	continue
			//}
		}
	}()

	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) client() (*client.Client, error) {
	peer, err := s.agent.LocalNode()
	if err != nil {
		return nil, err
	}
	return client.NewClient(peer.Addr)
}

func (s *service) peerAddr() (string, error) {
	peer, err := s.agent.LocalNode()
	if err != nil {
		return "", err
	}

	return peer.Addr, nil
}

func (s *service) nodeName() string {
	return s.agent.Config().NodeName
}

func (s *service) nodeClient(id string) (*client.Client, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	nodes, err := c.Cluster().Nodes()
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if node.Name == id {
			return client.NewClient(node.Addr)
		}
	}

	return nil, fmt.Errorf("node %s not found in cluster", id)
}

func (s *service) caddyLoader(serverType string) (caddy.Input, error) {
	data, err := generateConfig()
	if err != nil {
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: serverType,
	}, nil

}
