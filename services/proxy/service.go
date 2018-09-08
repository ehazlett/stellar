package proxy

import (
	"fmt"
	"net/url"
	"time"

	"github.com/containerd/containerd"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/version"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/mholt/caddy"
	_ "github.com/mholt/caddy/caddyhttp"
	"github.com/mholt/caddy/caddytls"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.proxy.v1"
)

var (
	empty = &ptypes.Empty{}
)

type endpoint struct {
	url *url.URL
	tls bool
}

type service struct {
	containerdAddr string
	namespace      string
	agent          *element.Agent
	cfg            *stellar.Config
	instance       *caddy.Instance
	currentServers []*Server
	configID       string
	errCh          chan error
}

func New(cfg *stellar.Config, agent *element.Agent) (*service, error) {
	errCh := make(chan error)
	go func() {
		for {
			err := <-errCh
			logrus.Errorf("proxy: %s", err)
		}
	}()

	caddy.AppName = version.Name
	caddy.AppVersion = version.FullVersion()
	caddy.Quiet = true
	caddytls.Agreed = true
	caddytls.DefaultCAUrl = "https://acme-v02.api.letsencrypt.org/directory"
	caddytls.DefaultEmail = cfg.ProxyTLSEmail

	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		cfg:            cfg,
		currentServers: []*Server{},
		errCh:          errCh,
		agent:          agent,
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
	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(s.defaultLoader))
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
			logrus.Debug("proxy update check")
			// we do a periodic reload.  this might be better at scale
			// if we check for application updates before trying to reload
			if err := s.reload(); err != nil {
				logrus.Error(err)
				continue
			}
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

func (s *service) getApplications() ([]*applicationapi.App, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.Application().List()
}

func (s *service) defaultLoader(serverType string) (caddy.Input, error) {
	return caddy.CaddyfileInput{
		Contents:       []byte(fmt.Sprintf(":%d", s.cfg.ProxyHTTPPort)),
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: serverType,
	}, nil

}

func (s *service) getCaddyConfig() (caddy.Input, error) {
	data, err := s.generateConfig()
	if err != nil {
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       data,
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: "http",
	}, nil

}
