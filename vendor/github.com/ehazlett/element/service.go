package element

import "google.golang.org/grpc"

type Service interface {
	// ID is the name of the service
	ID() string
	// Register is used to register the GRPC service
	Register(srv *grpc.Server) error
}
