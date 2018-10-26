package datastore

import (
	"context"
	"encoding/json"
	"io"

	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/client"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Sync returns a stream of all datastore buckets and keys
func (s *service) Sync(_ *api.SyncRequest, srv api.Datastore_SyncServer) error {
	logrus.Debug("syncing datastore")
	// set operations
	if err := s.db.View(func(tx *bolt.Tx) error {
		if err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			bucket := string(name)
			// skip tombstone buckets
			if bucket == s.dsTombstoneBucketName {
				return nil
			}
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if err := srv.Send(&api.SyncOperation{
					Bucket: bucket,
					Key:    string(k),
					Value:  v,
					Action: api.SyncAction_SET,
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	// delete operations
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.dsTombstoneBucketName))
		if b == nil {
			// if bucket doesn't exist there is no need to sync
			return nil
		}

		if err := b.ForEach(func(k, v []byte) error {
			var t *tombstone
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			if err := srv.Send(&api.SyncOperation{
				Bucket: t.Bucket,
				Key:    t.Key,
				Value:  t.Value,
				Action: api.SyncAction_DELETE,
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// PeerSync issues the local node to sync with the requested peer
func (s *service) PeerSync(ctx context.Context, req *api.PeerSyncRequest) (*ptypes.Empty, error) {
	logrus.Debugf("performing datastore sync with peer %s", req.ID)
	c, err := client.NewClient(req.Address)
	if err != nil {
		return empty, err
	}
	defer c.Close()

	stream, err := c.DatastoreService().Sync(ctx, &api.SyncRequest{})
	if err != nil {
		return empty, err
	}

	for {
		op, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return empty, errors.Wrap(err, "error syncing datastore")
		}

		switch op.Action {
		case api.SyncAction_SET:
			if _, err := s.Set(ctx, &api.SetRequest{
				Bucket: op.Bucket,
				Key:    op.Key,
				Value:  op.Value,
			}); err != nil {
				return empty, errors.Wrapf(err, "sync: error setting key %s", op.Key)
			}
		case api.SyncAction_DELETE:
			if _, err := s.Delete(ctx, &api.DeleteRequest{
				Bucket:      op.Bucket,
				Key:         op.Key,
				NoTombstone: true,
			}); err != nil {
				return empty, errors.Wrapf(err, "sync: error removing key %s", op.Key)
			}
		}
	}
	return empty, nil
}

func (s *service) replicateToPeers(ctx context.Context) error {
	localNode := s.agent.Self()
	peers, err := s.agent.Peers()
	if err != nil {
		return err
	}

	if len(peers) == 0 {
		return nil
	}

	for _, peer := range peers {
		logrus.Debugf("performing sync with peer %s", peer.ID)
		c, err := client.NewClient(peer.Address)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"peer": peer.ID,
			}).Errorf("error performing sync: %s", err)
			continue
		}
		defer c.Close()

		if _, err := c.DatastoreService().PeerSync(ctx, &api.PeerSyncRequest{
			ID:      localNode.ID,
			Address: localNode.Address,
		}); err != nil {
			logrus.WithFields(logrus.Fields{
				"peer": peer.ID,
			}).Errorf("peer sync error: %s", err)
			continue
		}
	}

	return nil
}
