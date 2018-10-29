package events

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/events/v1"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/nats-io/gnatsd/server"
	nats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.events.v1"
)

var (
	empty = &ptypes.Empty{}
)

type service struct {
	agent *element.Agent
	// set on start
	srv    *server.Server
	config *stellar.Config
}

func New(cfg *stellar.Config, a *element.Agent) (*service, error) {
	return &service{
		agent:  a,
		config: cfg,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterEventsServer(server, s)
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
	// TODO: get peers to build routes
	routes := []*url.URL{}
	peers, err := s.agent.Peers()
	if err != nil {
		return err
	}
	for _, peer := range peers {
		c, err := client.NewClient(peer.Address)
		if err != nil {
			return err
		}
		ep, err := c.Events().Endpoint(&api.EndpointRequest{})
		if err != nil {
			return err
		}
		routes = append(routes, &url.URL{
			Scheme: "nats",
			Host:   ep.Address,
		})
		c.Close()
	}

	natsHost, snatsPort, err := net.SplitHostPort(s.config.EventsAddress)
	if err != nil {
		return err
	}
	natsPort, err := strconv.Atoi(snatsPort)
	if err != nil {
		return err
	}

	clusterHost, sclusterPort, err := net.SplitHostPort(s.config.EventsClusterAddress)
	if err != nil {
		return err
	}
	clusterPort, err := strconv.Atoi(sclusterPort)
	if err != nil {
		return err
	}

	opts := &server.Options{
		Debug: true,
		Host:  natsHost,
		Port:  natsPort,
		Cluster: server.ClusterOpts{
			Host: clusterHost,
			Port: clusterPort,
		},
		Routes: routes,
		NoSigs: true,
	}

	// enable http if configured
	if s.config.EventsHTTPAddress != "" {
		httpHost, shttpPort, err := net.SplitHostPort(s.config.EventsHTTPAddress)
		if err != nil {
			return err
		}
		httpPort, err := strconv.Atoi(shttpPort)
		if err != nil {
			return err
		}
		opts.HTTPHost = httpHost
		opts.HTTPPort = httpPort
	}

	logrus.WithFields(logrus.Fields{
		"opts": fmt.Sprintf("%+v", opts),
	}).Debug("events configuration")

	srv := server.New(opts)
	s.srv = srv
	go srv.Start()
	return nil
}

func (s *service) Stop() error {
	s.srv.Shutdown()
	return nil
}

func (s *service) client() (*client.Client, error) {
	return client.NewClient(s.agent.Self().Address)
}

func (s *service) natsClient() (*nats.Conn, error) {
	return nats.Connect("nats://" + s.config.EventsAddress)
}
