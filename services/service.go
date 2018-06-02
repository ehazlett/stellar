package services

import "google.golang.org/grpc"

type Service interface {
	ID() string
	Register(srv *grpc.Server) error
}
