package server

import (
	"github.com/ehazlett/stellar/client"
)

func (s *Server) reconcile() error {
	localNode, err := s.agent.LocalNode()
	if err != nil {
		return err
	}
	c, err := client.NewClient(localNode.Addr)
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
