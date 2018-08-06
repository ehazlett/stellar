package nameserver

import (
	"context"
	"fmt"
	"net"

	"github.com/containerd/typeurl"
	api "github.com/ehazlett/stellar/api/services/nameserver/v1"
	"github.com/ehazlett/stellar/api/types"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// TODO: support multiple RR per name
// TODO: add forwarder for nameserver

func (s *service) startDNSServer() error {
	// TODO: make configurable
	gw, err := s.gateway()
	if err != nil {
		return err
	}

	dns.HandleFunc(".", s.handler)

	for _, proto := range []string{"tcp4", "udp4"} {
		srv := &dns.Server{
			Addr: fmt.Sprintf("%s:53", gw.String()),
			Net:  proto,
		}
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				logrus.Errorf("error starting dns server on 53/%s", srv.Net)
			}
		}()
	}

	return nil
}

func (s *service) gateway() (net.IP, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	subnetCIDR, err := c.Network().AllocateSubnet(s.agent.Config().NodeName)
	if err != nil {
		return nil, err
	}
	if subnetCIDR == "" {
		return nil, fmt.Errorf("unable to detect subnet for node")
	}
	logrus.Debugf("gateway cidr: %s", subnetCIDR)
	ip, ipnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return nil, err
	}

	gw := ip.Mask(ipnet.Mask)
	gw[3]++

	return gw, nil
}

func (s *service) handler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	defer w.WriteMsg(m)
	query := m.Question[0].Name
	logrus.Debugf("nameserver: query=%q", query)
	name := getName(query)
	logrus.Debugf("nameserver: looking up %s", name)
	resp, err := s.Lookup(context.Background(), &api.LookupRequest{
		Query: name,
	})
	if err != nil {
		logrus.Error(errors.Wrapf(err, "nameserver: error performing lookup for %s", name))
		//w.WriteMsg(m)
		return
	}
	m.Answer = make([]dns.RR, len(resp.Records))
	for i, record := range resp.Records {
		var rr dns.RR
		switch record.Type {
		case api.RecordType_A:
			ip := net.ParseIP(string(record.Value))
			rr = &dns.A{
				Hdr: dns.RR_Header{
					Name:   query,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: ip,
			}
		case api.RecordType_CNAME:
			rr = &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   query,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Target: string(record.Value),
			}
		case api.RecordType_TXT:
			rr = &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   query,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{string(record.Value)},
			}
		case api.RecordType_SRV:
			v, err := typeurl.UnmarshalAny(record.Options)
			if err != nil {
				logrus.Errorf("nameserver: unmarshalling record options: %s", err)
				return
			}
			o, ok := v.(*types.SRVOptions)
			if !ok {
				logrus.Error("nameserver: invalid type for record options; expected SRVOptions")
				return
			}
			rr = &dns.SRV{
				Hdr: dns.RR_Header{
					Name:   query,
					Rrtype: dns.TypeSRV,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Target:   string(record.Value),
				Priority: o.Priority,
				Weight:   o.Weight,
				Port:     o.Port,
			}
		case api.RecordType_MX:
			rr = &dns.MX{
				Hdr: dns.RR_Header{
					Name:   query,
					Rrtype: dns.TypeMX,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Mx: string(record.Value),
			}
		default:
			logrus.Errorf("nameserver: unsupported record type %s for %s", record.Type, name)
			//w.WriteMsg(m)
			return
		}

		m.Answer[i] = rr
	}

	//w.WriteMsg(m)
}

func getName(query string) string {
	return query[:len(query)-1]
}
