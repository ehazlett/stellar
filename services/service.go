package services

import (
	"google.golang.org/grpc"
)

// Service is the interface that all stellar services must implement
type Service interface {
	// ID returns the id of the service
	ID() string
	// Register registers the service with the GRPC server
	Register(*grpc.Server) error
	// Start provides a mechanism to start service specific actions
	Start() error
	// stop provides a mechanism to stop the service
	Stop() error
}
