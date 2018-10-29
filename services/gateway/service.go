package cluster

import (
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
}

func New(cfg *stellar.Config, a *element.Agent) (*service, error) {
	return &service{
		grpcHost:    cfg.AgentConfig.ClusterAddress,
		gatewayAddr: cfg.GatewayAddress,
	}, nil
}

func (s *service) Register(server *grpc.Server) error {
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Info(ctx context.Context, req *cluster.InfoRequest) (*cluster.InfoResponse, error) {
	return &cluster.InfoResponse{
		ID: serviceID,
	}, nil
}

func (s *service) Start() error {
	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithWaitForHandshake(),
	}

	err := cluster.RegisterClusterHandlerFromEndpoint(ctx, mux, s.grpcHost, opts)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"grpcHost":    s.grpcHost,
		"gatewayAddr": s.gatewayAddr,
	}).Info("starting http gateway")

	go func() {
		if err := http.ListenAndServe(s.gatewayAddr, mux); err != nil {
			logrus.Error(err)
		}
	}()

	return nil
}

func (s *service) Stop() error {
	return nil
}
