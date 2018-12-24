package datastore

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/stellarproject/element"
	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/ehazlett/stellar/services"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	serviceID = "stellar.services.datastore.v1"
)

var (
	dbFilename = serviceID + ".db"
	empty      = &ptypes.Empty{}
)

type service struct {
	agent                 *element.Agent
	config                *stellar.Config
	dir                   string
	lock                  *sync.Mutex
	lockChan              chan bool
	db                    *bolt.DB
	dsTombstoneBucketName string
}

type tombstone struct {
	Timestamp time.Time
	Bucket    string
	Key       string
	Value     []byte
}

func New(cfg *stellar.Config, agent *element.Agent) (services.Service, error) {
	svc := &service{
		agent:                 agent,
		config:                cfg,
		dir:                   cfg.DataDir,
		lock:                  &sync.Mutex{},
		lockChan:              make(chan bool),
		dsTombstoneBucketName: "stellar." + stellar.APIVersion + "." + agent.Self().ID + ".services.datastore.tombstone",
	}

	db, err := svc.openDB()
	if err != nil {
		return nil, err
	}

	svc.db = db

	return svc, nil
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterDatastoreServer(server, s)
	return nil
}

func (s *service) ID() string {
	return serviceID
}

func (s *service) Type() services.Type {
	return services.DatastoreService
}

func (s *service) Requires() []services.Type {
	return nil
}

func (s *service) Info(ctx context.Context, req *api.InfoRequest) (*api.InfoResponse, error) {
	return &api.InfoResponse{
		ID: serviceID,
	}, nil
}

func (s *service) Start() error {
	s.lock.Lock()
	if err := s.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(s.dsTombstoneBucketName)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	s.lock.Unlock()

	if err := s.db.View(func(tx *bolt.Tx) error {
		if err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			bucket := string(name)
			logrus.Debugf("datastore: bucket %s", bucket)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	t := time.NewTicker(time.Second * 60)
	go func() {
		for range t.C {
			logrus.Debug("pruning datastore")
			if err := s.prune(); err != nil {
				logrus.Errorf("error pruning datastore: %s", err)
			}
		}
	}()

	return nil
}

func (s *service) Stop() error {
	logrus.Debug("performing shutdown sync with peers")
	ctx := context.Background()
	if err := s.replicateToPeers(ctx); err != nil {
		return err
	}

	return nil
}

func (s *service) Close() {
	s.db.Close()
}

func (s *service) Reset() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.db.Close()

	if err := os.Remove(dbFilename); err != nil {
		return err
	}

	db, err := s.openDB()
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *service) openDB() (*bolt.DB, error) {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(s.dir, dbFilename), 0600, &bolt.Options{Timeout: time.Second * 1})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *service) client(address string) (*client.Client, error) {
	opts, err := client.DialOptionsFromConfig(s.config)
	if err != nil {
		return nil, err
	}
	return client.NewClient(address, opts...)
}
