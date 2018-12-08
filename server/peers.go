package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// ErrNoAvailablePeers returns when no peers are available upon node start
	ErrNoAvailablePeers = errors.New("no available peers")
)

type Peer struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

// cachePeer saves the specified cluster peer to the local cache
func (s *Server) cachePeer(peer *Peer) error {
	// don't add self
	if peer.ID == s.config.NodeID {
		return nil
	}
	logrus.WithField("peer", peer).Debug("caching cluster peer")
	// cache peers
	peerData, err := json.Marshal(peer)
	if err != nil {
		return err
	}
	if err := s.db.Set(dsLocalPeerBucketName, peer.ID, peerData); err != nil {
		return err
	}

	return nil
}

// getPeersFromCache returns the list of reachable (via tcp) peers either from the config or local cache
func getPeersFromCache(db *localDB, seedPeers []string) ([]string, error) {
	logrus.Debugf("getPeersFromCache: %+v", seedPeers)
	kvs, err := db.GetAll(dsLocalPeerBucketName)
	if err != nil {
		return nil, err
	}

	if kvs == nil {
		return seedPeers, nil
	}

	var cachedPeers []*Peer
	for _, kv := range kvs {
		var peer *Peer
		if err := json.Unmarshal(kv.Value, &peer); err != nil {
			return nil, err
		}
		cachedPeers = append(cachedPeers, peer)
	}

	cachedAddresses := []string{}
	for _, c := range cachedPeers {
		cachedAddresses = append(cachedAddresses, c.Address)
	}
	logrus.Debugf("cachedAddresses: %+v", cachedAddresses)

	peers, err := getClusterPeers(seedPeers, cachedAddresses)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

// getClusterPeers returns a list of available peers from the seed list
// this is used to get the initial list of peers when starting. there
// are a few cases:
//
// - start with different seed peers than present in the cache
//   cache peers are ignored and seed peer list is returned
//
// - start with no seed peers
//   if this is the case, the local cache will be checked for a peer list.
//   this is used in case the initial server started with no peers and is
//   now reconnecting to an existing cluster.  each peer is checked and
//   whichever peers are alive (via tcp) they are returned.  if none are
//   available it is assumed this should be the initial node.
//
// - start with seed peers
//   if seed peers are present each one is checked for availability (tcp)
//   whichever are reachable they are returned as the peer list.  if none
//   are reachable an error is returned stating no peers are available.
func getClusterPeers(seedPeers []string, cachedPeers []string) ([]string, error) {
	// initial node with no cached peers
	if len(seedPeers) == 0 && len(cachedPeers) == 0 {
		return []string{}, nil
	}
	// if no seed peers are passed use cached
	peers := cachedPeers

	// parse seed peers and validate cached
	if len(seedPeers) > 0 {
		seedsCached, err := cachedSeedPeers(seedPeers, cachedPeers)
		if err != nil {
			return nil, err
		}

		if !seedsCached {
			return resolvePeers(seedPeers), nil
		}
	}

	resolvedPeers := resolvePeers(peers)
	// check if a peer and no cached an error is returned
	if len(seedPeers) > 0 && len(resolvedPeers) == 0 {
		return nil, ErrNoAvailablePeers
	}

	return resolvePeers(peers), nil
}

// cachedSeedPeers checks that the specified seedPeers are in the local cached peers
func cachedSeedPeers(seedPeers, cachedPeers []string) (bool, error) {
	// build reference map for lookup
	cached := map[string]string{}
	for _, p := range cachedPeers {
		// perform a dns lookup to ensure that we check the addresses instead of names
		ip, port, err := getIPPort(p)
		if err != nil {
			return false, err
		}
		cached[fmt.Sprintf("%s:%s", ip, port)] = p
	}
	// validate that each seed peer is in the cache
	for _, p := range seedPeers {
		ip, port, err := getIPPort(p)
		if err != nil {
			return false, err
		}

		// check that seed address is one of the cached peers
		seedAddress := fmt.Sprintf("%s:%s", ip, port)
		if _, ok := cached[seedAddress]; !ok {
			logrus.Debugf("%s not in %+v", seedAddress, cached)
			return false, nil
		}
	}

	return true, nil
}

// resolvePeers returns a list of peers that are reachable by tcp
func resolvePeers(peers []string) []string {
	resolved := []string{}
	// perform tcp check to ensure peer is alive
	for _, peer := range peers {
		conn, err := net.Dial("tcp", peer)
		if err != nil {
			logrus.WithError(err).Warnf("error connecting to peer %s", peer)
			continue
		}
		conn.SetReadDeadline(time.Now())
		if _, err := conn.Read([]byte{}); err == io.EOF {
			logrus.Warnf("unable to connect to peer %s; skipping", peer)
			continue
		}

		resolved = append(resolved, peer)
	}

	return resolved
}

// getIPPort returns the ip address and port as resolved by dns (uses the first address returned)
func getIPPort(address string) (string, string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", "", errors.Wrapf(err, "invalid address: %s", address)
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", "", errors.Wrapf(err, "error looking up record for %s", host)
	}

	return addrs[0], port, nil
}
