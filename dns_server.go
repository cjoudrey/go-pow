package main

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
)

type DnsServer struct {
	Configuration *Configuration
}

func NewDnsServer(configuration *Configuration) *DnsServer {
	server := DnsServer{
		Configuration: configuration,
	}
	return &server
}

func (s *DnsServer) Listen() error {
	s.setupHandlers()
	server := &dns.Server{Addr: fmt.Sprintf("127.0.0.1:%v", s.Configuration.DnsPort), Net: "udp"}
	err := server.ListenAndServe()
	return err
}

var handler = func(w dns.ResponseWriter, r *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = append(m.Answer, &dns.A{
		Hdr: dns.RR_Header{
			Name:   r.Question[0].Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    600,
		},
		A: net.ParseIP("127.0.0.1").To4(),
	})
	w.WriteMsg(m)
}

func (s *DnsServer) setupHandlers() {
	for _, domain := range s.Configuration.Domains {
		dns.HandleFunc(fmt.Sprintf("%s.", domain), handler)
	}

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := &dns.Msg{}
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
	})
}
