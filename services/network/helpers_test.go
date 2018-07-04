package network

import (
	"net"
	"testing"
)

func TestNetworkSubdivide172(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		t.Fatal(err)
	}

	maxSubnets := 1024

	subnets, err := divideSubnet(ipnet, maxSubnets)
	if err != nil {
		t.Fatal(err)
	}

	if v := len(subnets); v != maxSubnets {
		t.Fatalf("expected %d subnets; received %d", maxSubnets, v)
	}
	testSub1 := subnets[0]
	testIP1 := net.IPv4(172, 16, 0, 42)
	testSub2 := subnets[len(subnets)-1]
	testIP2 := net.IPv4(172, 31, 252, 42)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", string(testIP1), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", string(testIP2), testSub2)
	}
}

func TestNetworkSubdivide10(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}

	maxSubnets := 1024

	subnets, err := divideSubnet(ipnet, maxSubnets)
	if err != nil {
		t.Fatal(err)
	}

	if v := len(subnets); v != maxSubnets {
		t.Fatalf("expected %d subnets; received %d", maxSubnets, v)
	}
	testSub1 := subnets[0]
	testIP1 := net.IPv4(10, 0, 0, 42)
	testSub2 := subnets[len(subnets)-1]
	testIP2 := net.IPv4(10, 255, 192, 42)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", string(testIP1), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", string(testIP2), testSub2)
	}
}
