package server

import (
	"github.com/ehazlett/stellar/client"
)

func (s *Server) reconcile() error {
	c, err := client.NewClient(s.agent.Self().Address)
	if err != nil {
		return err
	}
	defer c.Close()

	// setup cluster routes
	if err := s.setupRoutes(); err != nil {
		return err
	}

	return nil
}
