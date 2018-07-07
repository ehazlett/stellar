package datastore

import (
	"os"
	"sync"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/ehazlett/element"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
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
	agent    *element.Agent
	dir      string
	lock     *sync.Mutex
	lockChan chan bool
	db       *bolt.DB
}

func New(a *element.Agent, dir string) (*service, error) {
	svc := &service{
		agent:    a,
		dir:      dir,
		lock:     &sync.Mutex{},
		lockChan: make(chan bool),
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
	db, err := bolt.Open(dbFilename, 0600, &bolt.Options{Timeout: time.Second * 1})
	if err != nil {
		return nil, err
	}

	return db, nil
}
