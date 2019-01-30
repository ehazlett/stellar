package network

import (
	"net"
	"testing"
)

func TestNetworkSubdivide10Slash8(t *testing.T) {
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
	testIP2 := net.IPv4(10, 255, 255, 42)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash12(t *testing.T) {
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
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash16(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/16")
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
	testIP2 := net.IPv4(172, 16, 255, 254)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash17(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/17")
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
	testIP1 := net.IPv4(172, 16, 0, 10)
	testSub2 := subnets[len(subnets)-1]
	testIP2 := net.IPv4(172, 16, 127, 254)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash18(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/18")
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
	testIP1 := net.IPv4(172, 16, 0, 10)
	testSub2 := subnets[len(subnets)-1]
	testIP2 := net.IPv4(172, 16, 63, 254)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash19(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/19")
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
	testIP1 := net.IPv4(172, 16, 0, 1)
	testSub2 := subnets[len(subnets)-1]
	testIP2 := net.IPv4(172, 16, 31, 254)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash20(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/20")
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
	testIP1 := net.IPv4(172, 16, 0, 1)
	testSub2 := subnets[len(subnets)-1]
	testIP2 := net.IPv4(172, 16, 15, 254)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash21(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/21")
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
	testIP1 := net.IPv4(172, 16, 0, 1)
	testSub2 := subnets[len(subnets)-1]
	testIP2 := net.IPv4(172, 16, 7, 254)

	if !testSub1.Contains(testIP1) {
		t.Fatalf("expected ip %s in subnet %s", testIP1.String(), testSub1)
	}

	if !testSub2.Contains(testIP2) {
		t.Fatalf("expected ip %s in subnet %s", testIP2.String(), testSub2)
	}
}

func TestNetworkSubdivide172Slash22(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/22")
	if err != nil {
		t.Fatal(err)
	}

	maxSubnets := 1024

	_, err = divideSubnet(ipnet, maxSubnets)
	if err == nil {
		t.Fatal("expected error for /22 subnet")
	}
}

func TestNetworkSubdivide172Slash23(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("172.16.0.0/23")
	if err != nil {
		t.Fatal(err)
	}

	maxSubnets := 1024

	_, err = divideSubnet(ipnet, maxSubnets)
	if err == nil {
		t.Fatal("expected error for /23 subnet")
	}
}

func TestNetworkSubdivide192Slash24(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("192.168.0.0/24")
	if err != nil {
		t.Fatal(err)
	}

	maxSubnets := 1024

	_, err = divideSubnet(ipnet, maxSubnets)
	if err == nil {
		t.Fatal("expected error for /24 subnet")
	}
}
