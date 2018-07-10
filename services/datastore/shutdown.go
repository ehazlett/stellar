package datastore

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Shutdown(ctx context.Context, req *api.ShutdownRequest) (*ptypes.Empty, error) {
	peers, err := s.agent.Peers()
	if err != nil {
		return empty, err
	}

	if len(peers) == 0 {
		return empty, nil
	}

	for _, peer := range peers {
		logrus.Debugf("performing shutdown sync with peer %s", peer.Name)
		c, err := client.NewClient(peer.Addr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"peer": peer.Name,
			}).Errorf("error performing shutdown sync: %s", err)
			continue
		}
		defer c.Close()

		if _, err := c.DatastoreService().PeerSync(ctx, &api.PeerSyncRequest{
			Name: peer.Name,
			Addr: peer.Addr,
		}); err != nil {
			logrus.WithFields(logrus.Fields{
				"peer": peer.Name,
			}).Errorf("peer sync error: %s", err)
			continue
		}
	}
	return empty, nil
}
