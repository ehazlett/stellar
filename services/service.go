package services

import "google.golang.org/grpc"

type Service interface {
	// ID returns the id of the service
	ID() string
	// Register registers the service with the GRPC server
	Register(srv *grpc.Server) error
	// Start provides a mechanism to start service specific actions
	Start() error
}
