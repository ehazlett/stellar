package proxy

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/ehazlett/blackbird"
	blackbirdserver "github.com/ehazlett/blackbird/server"
	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	eventsapi "github.com/ehazlett/stellar/api/services/events/v1"
	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/events"
	appsvc "github.com/ehazlett/stellar/services/application"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	serviceID         = "stellar.services.proxy.v1"
	blackbirdGRPCAddr = "unix:///run/blackbird.sock"
	dsProxyBucketName = "stellar." + stellar.APIVersion + ".services.proxy"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	containerdAddr string
	namespace      string
	agent          *element.Agent
	config         *stellar.Config
	errCh          chan error

	// set on start
	server  *blackbirdserver.Server
	bclient *blackbird.Client
}

func New(cfg *stellar.Config, agent *element.Agent) (*service, error) {
	errCh := make(chan error)
	go func() {
		for {
			err := <-errCh
			logrus.Errorf("proxy: %s", err)
		}
	}()

	return &service{
		containerdAddr: cfg.ContainerdAddr,
		namespace:      cfg.Namespace,
		agent:          agent,
		config:         cfg,
		errCh:          errCh,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterProxyServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Info(ctx context.Context, req *api.InfoRequest) (*api.InfoResponse, error) {
	return &api.InfoResponse{
		ID: serviceID,
	}, nil
}

func (s *service) Start() error {
	client, err := s.client(s.agent.Self().Address)
	if err != nil {
		return err
	}
	config := &blackbird.Config{
		GRPCAddr:  blackbirdGRPCAddr,
		HTTPPort:  s.config.ProxyHTTPPort,
		HTTPSPort: s.config.ProxyHTTPSPort,
		Debug:     false,
	}
	ds, err := newDatastore(client)
	if err != nil {
		return err
	}
	srv, err := blackbirdserver.NewServer(config, ds)
	if err != nil {
		return err
	}
	if err := srv.Run(); err != nil {
		return err
	}

	bc, err := blackbird.NewClient(blackbirdGRPCAddr)
	if err != nil {
		return err
	}
	s.bclient = bc

	// initial reload
	if err := s.reload(); err != nil {
		return err
	}

	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return err
	}

	// start listener for application events
	go func() {
		defer c.Close()

		subject, err := c.Application().ID()
		if err != nil {
			logrus.WithError(err).Error("error getting application subject for events")
			return
		}
		stream, err := c.EventsService().Subscribe(context.Background(), &eventsapi.SubscribeRequest{
			Subject: subject,
		})
		if err != nil {
			logrus.WithError(err).Error("error subscribing to application events")
			return
		}

		for {
			evt, err := stream.Recv()
			if err != nil {
				logrus.WithError(err).Error("error subscribing to application events")
				return
			}

			if events.IsEvent(evt, &appsvc.UpdateEvent{}) {
				logrus.Debug("reloading proxy")
				if err := s.reload(); err != nil {
					logrus.Error(err)
					continue
				}
			}
		}
	}()

	return nil
}

func (s *service) Stop() error {
	return nil
}

func (s *service) containerd() (*containerd.Client, error) {
	return stellar.DefaultContainerd(s.containerdAddr, s.namespace)
}

func (s *service) nodeName() string {
	return s.agent.Self().ID
}

func (s *service) client(address string) (*client.Client, error) {
	opts, err := client.DialOptionsFromConfig(s.config)
	if err != nil {
		return nil, err
	}
	return client.NewClient(address, opts...)
}

func (s *service) getApplications() ([]*applicationapi.App, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.Application().List()
}
