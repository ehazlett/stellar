package server

import (
	"net"
	"reflect"
	"testing"
)

func TestClusterPeersGetIPPort(t *testing.T) {
	expectedIP := "127.0.0.1"
	expectedPort := "9000"
	address := "localhost:" + expectedPort
	ip, port, err := getIPPort(address)
	if err != nil {
		t.Fatal(err)
	}

	if ip != expectedIP {
		t.Errorf("expected ip %s; received %s", expectedIP, ip)
	}

	if port != expectedPort {
		t.Errorf("expected port %s; received %s", expectedPort, port)
	}
}

func TestClusterPeersCachedPeers(t *testing.T) {
	cases := []struct {
		name        string
		seedPeers   []string
		cachedPeers []string
		expected    []string
	}{
		{
			name:        "initial node",
			seedPeers:   []string{},
			cachedPeers: nil,
			expected:    []string{},
		},
		{
			name:        "initial node with cached peer",
			seedPeers:   []string{},
			cachedPeers: []string{"127.0.0.1:65000"},
			expected: []string{
				"127.0.0.1:65000",
			},
		},
		{
			name:        "initial node with cached named peer",
			seedPeers:   []string{},
			cachedPeers: []string{"localhost:65000"},
			expected: []string{
				"localhost:65000",
			},
		},
		{
			name:        "initial node with cached peers (1 simulated down)",
			seedPeers:   []string{},
			cachedPeers: []string{"127.0.0.1:65000", "127.0.1.1:65001", "127.0.1.1:0"},
			expected: []string{
				"127.0.0.1:65000",
				"127.0.1.1:65001",
			},
		},
		{
			name:        "node with seed peer and multiple cached peers",
			seedPeers:   []string{"127.0.0.1:65000"},
			cachedPeers: []string{"127.0.0.1:65000", "127.0.1.1:65000", "127.0.1.2:65001"},
			expected: []string{
				"127.0.0.1:65000",
				"127.0.1.1:65000",
				"127.0.1.2:65001",
			},
		},
	}

	for _, c := range cases {
		// start listeners for tcp check to pass
		listeners := []net.Listener{}
		for _, p := range c.cachedPeers {
			l, err := net.Listen("tcp", p)
			if err != nil {
				continue
			}
			listeners = append(listeners, l)
		}
		peers, err := getClusterPeers(c.seedPeers, c.cachedPeers)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(peers, c.expected) {
			t.Errorf("case %s fail: expected %+v; received %+v", c.name, c.expected, peers)
		}
		// close listeners for cached peers
		for _, l := range listeners {
			l.Close()
		}
	}
}

func TestClusterPeersSeedPeers(t *testing.T) {
	// find free port
	cases := []struct {
		name        string
		listen      bool
		seedPeers   []string
		cachedPeers []string
		expected    []string
		err         error
	}{
		{
			name:        "node with seed peers and no cached",
			listen:      true,
			seedPeers:   []string{"127.0.0.1:65000", "127.0.1.1:65000"},
			cachedPeers: []string{},
			expected: []string{
				"127.0.0.1:65000",
				"127.0.1.1:65000",
			},
			err: nil,
		},
		{
			name:        "node with seed peers and no cached (1 seed down)",
			listen:      true,
			seedPeers:   []string{"127.0.0.1:65000", "127.0.1.1:65001", "127.0.1.1:0"},
			cachedPeers: []string{},
			expected: []string{
				"127.0.0.1:65000",
				"127.0.1.1:65001",
			},
			err: nil,
		},
		{
			name:        "node with seed peers and cached",
			listen:      true,
			seedPeers:   []string{"127.0.0.1:65000"},
			cachedPeers: []string{"127.0.0.1:65000"},
			expected: []string{
				"127.0.0.1:65000",
			},
			err: nil,
		},
		{
			name:        "node with seed peers and mismatching cached",
			listen:      true,
			seedPeers:   []string{"127.0.0.1:65000"},
			cachedPeers: []string{"127.0.1.1:65000"},
			expected: []string{
				"127.0.0.1:65000",
			},
			err: nil,
		},
		{
			name:        "node with seed peers and no cached (all down)",
			listen:      false,
			seedPeers:   []string{"127.0.0.1:65000"},
			cachedPeers: []string{"127.0.0.1:65000"},
			expected:    nil,
			err:         ErrNoAvailablePeers,
		},
	}

	for _, c := range cases {
		// start listeners for tcp check to pass
		listeners := []net.Listener{}
		if c.listen {
			for _, p := range c.seedPeers {
				l, err := net.Listen("tcp", p)
				if err != nil {
					continue
				}
				listeners = append(listeners, l)
			}
		}
		peers, err := getClusterPeers(c.seedPeers, c.cachedPeers)
		if err != c.err {
			t.Error(err)
		}

		if !reflect.DeepEqual(peers, c.expected) {
			t.Errorf("case %s fail: expected %+v; received %+v", c.name, c.expected, peers)
		}
		if c.listen {
			// close listeners for cached peers
			for _, l := range listeners {
				l.Close()
			}
		}
	}
}
