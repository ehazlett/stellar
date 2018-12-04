package server

import (
	"fmt"
	"net"
	"net/url"
	"os"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/mholt/caddy"
	_ "github.com/mholt/caddy/caddyhttp"
	"github.com/mholt/caddy/caddytls"
	"github.com/sirupsen/logrus"
	"github.com/stellarproject/radiant"
	api "github.com/stellarproject/radiant/api/v1"
	"github.com/stellarproject/radiant/ds"
	"github.com/stellarproject/radiant/version"
	"google.golang.org/grpc"
)

var (
	empty = &ptypes.Empty{}
)

type Server struct {
	config     *radiant.Config
	grpcServer *grpc.Server
	instance   *caddy.Instance
	datastore  ds.Datastore
}

func NewServer(cfg *radiant.Config, datastore ds.Datastore) (*Server, error) {
	grpcServer := grpc.NewServer()
	srv := &Server{
		config:     cfg,
		grpcServer: grpcServer,
		datastore:  datastore,
	}

	logrus.WithFields(logrus.Fields{
		"type": datastore.Name(),
	}).Debug("registered datastore")

	api.RegisterProxyServer(grpcServer, srv)

	caddy.AppName = version.Name
	caddy.AppVersion = version.FullVersion()
	caddy.Quiet = true
	caddytls.Agreed = true
	caddytls.DefaultCAUrl = "https://acme-v02.api.letsencrypt.org/directory"
	caddytls.DefaultEmail = cfg.TLSEmail

	return srv, nil
}

func (s *Server) Run() error {
	// start grpc
	l, err := getGRPCListener(s.config.GRPCAddr)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"addr": s.config.GRPCAddr,
	}).Debug("starting grpc listener")
	go s.grpcServer.Serve(l)

	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(s.defaultLoader))

	caddyfile, err := caddy.LoadCaddyfile("http")
	if err != nil {
		return err
	}
	s.instance, err = caddy.Start(caddyfile)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"name":    version.Name,
		"version": version.BuildVersion(),
		"http":    s.config.HTTPPort,
		"https":   s.config.HTTPSPort,
	}).Info("server started")

	return nil
}

func (s *Server) Stop() error {
	s.grpcServer.Stop()
	return nil
}

func getGRPCListener(uri string) (net.Listener, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "tcp":
		return net.Listen("tcp", u.Host)
	case "unix":
		if _, err := os.Stat(u.Path); err == nil {
			os.Remove(u.Path)
		}
		return net.Listen("unix", u.Path)
	default:
		return nil, fmt.Errorf("unsupported listener scheme %s (supported are tcp:// and unix://)", u.Scheme)
	}
}

func (s *Server) defaultLoader(serverType string) (caddy.Input, error) {
	return caddy.CaddyfileInput{
		Contents:       []byte(fmt.Sprintf(":%d", s.config.HTTPPort)),
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: serverType,
	}, nil
}

func (s *Server) getCaddyConfig() (caddy.Input, error) {
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
