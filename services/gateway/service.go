package cluster

import (
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/ehazlett/element"
	"github.com/ehazlett/stellar"
	cluster "github.com/ehazlett/stellar/api/services/cluster/v1"
)

const (
	serviceID = "stellar.services.gateway.v1"
)

type service struct {
	grpcHost    string
	gatewayAddr string
	gatewayPort int
}

func New(cfg *stellar.Config, a *element.Agent) (*service, error) {
	return &service{
		grpcHost:    fmt.Sprintf("%s:%d", cfg.AgentConfig.AgentAddr, cfg.AgentConfig.AgentPort),
		gatewayAddr: cfg.GatewayAddr,
		gatewayPort: cfg.GatewayPort,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Start() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := cluster.RegisterClusterHandlerFromEndpoint(ctx, mux, s.grpcHost, opts)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"grpcHost":    s.grpcHost,
		"gatewayAddr": s.gatewayAddr,
		"gatewayPort": s.gatewayPort,
	}).Info("starting http gateway")

	go func() {
		defer cancel()
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.gatewayAddr, s.gatewayPort), mux); err != nil {
			logrus.Error(err)
		}
	}()

	return nil
}
