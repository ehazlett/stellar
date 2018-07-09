package datastore

import (
	"context"
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
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			bucket := string(name)
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				logrus.Debugf("sending key=%s", string(k))
				if err := srv.Send(&api.SyncOperation{
					Bucket: bucket,
					Key:    string(k),
					Value:  v,
				}); err != nil {
					return err
				}
			}
			return nil
		})
	})
	return err
}

// PeerSync issues the local node to sync with the requested peer
func (s *service) PeerSync(ctx context.Context, req *api.PeerSyncRequest) (*ptypes.Empty, error) {
	logrus.Debugf("performing immediate sync with peer %s", req.Name)
	c, err := client.NewClient(req.Addr)
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

		if _, err := s.Set(ctx, &api.SetRequest{
			Bucket: op.Bucket,
			Key:    op.Key,
			Value:  op.Value,
		}); err != nil {
			return empty, errors.Wrapf(err, "error syncing key %s", op.Key)
		}
	}
	return empty, nil
}
