package network

import (
	"encoding/binary"
	"fmt"
	"net"
)

// adapted from https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloudup/subnets.go
func divideSubnet(sub *net.IPNet, maxSubnets int) ([]*net.IPNet, error) {
	length, _ := sub.Mask.Size()
	length += 10

	var subnets []*net.IPNet
	for i := 0; i < maxSubnets; i++ {
		ip4 := sub.IP.To4()
		if ip4 != nil {
			n := binary.BigEndian.Uint32(ip4)
			n += uint32(i) << uint(32-length)
			subnetIP := make(net.IP, len(ip4))
			binary.BigEndian.PutUint32(subnetIP, n)

			subnets = append(subnets, &net.IPNet{
				IP:   subnetIP,
				Mask: net.CIDRMask(length, 32),
			})

		} else {
			return nil, fmt.Errorf("unexpected IP address type: %s", sub)
		}
	}

	return subnets, nil
}
