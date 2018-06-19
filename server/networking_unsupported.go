// +build !linux

package server

import "fmt"

func (s *Server) initNetworking() error {
	return fmt.Errorf("networking not supported")
}

func (s *Server) setupRoutes() error {
	return fmt.Errorf("networking not supported")
}
