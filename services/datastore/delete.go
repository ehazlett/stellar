package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/containerd/containerd/errdefs"
	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *service) Delete(ctx context.Context, req *api.DeleteRequest) (*ptypes.Empty, error) {
	s.lock.Lock()
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(req.Bucket))
		if b == nil {
			return errdefs.ToGRPC(errors.Wrapf(errdefs.ErrNotFound, "bucket %s", req.Bucket))
		}
		val := b.Get([]byte(req.Key))
		if err := b.Delete([]byte(req.Key)); err != nil {
			return err
		}
		if !req.NoTombstone {
			// add tombstone
			tb := tx.Bucket([]byte(s.dsTombstoneBucketName))
			if tb == nil {
				return errdefs.ToGRPC(errors.Wrapf(errdefs.ErrNotFound, "bucket %s", s.dsTombstoneBucketName))
			}

			t := &tombstone{
				Timestamp: time.Now(),
				Bucket:    req.Bucket,
				Key:       req.Key,
				Value:     val,
			}
			data, err := json.Marshal(t)
			if err != nil {
				return err
			}
			key := fmt.Sprintf("%s.%s", s.agent.Config().NodeName, time.Now().Format(time.RFC3339Nano))
			if err := tb.Put([]byte(key), data); err != nil {
				return err
			}
			logrus.Debugf("datastore: added tombstone record for %s:%s", req.Bucket, req.Key)
		}
		return nil
	})
	s.lock.Unlock()

	logrus.WithFields(logrus.Fields{
		"bucket": req.Bucket,
		"key":    req.Key,
		"sync":   req.Sync,
	}).Debug("removed key from datastore")
	if req.Sync {
		if err := s.replicateToPeers(ctx); err != nil {
			return empty, err
		}
	}

	return empty, err
}
