package couchdb

import (
	"context"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// ServeDNS implements the plugin.Handler interface.
// The values the integer can take are the DNS RCODEs, dns.RcodeServerFailure, dns.RcodeNotImplemented, dns.RcodeSuccess, etc..
// A successful return value indicates the plugin has written to the client.
func (couchDB *CouchDB) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}
	qname := state.Name()
	qtype := state.QType()

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	qrecord, z := couchDB.Find(qname)
	if z.Rev == "" {
		return dns.RcodeRefused, nil // Refused   - Query Refused
	}

	switch qtype {
	case dns.TypeA:
		answers, extras = couchDB.A(qrecord, &z)
		if answers == nil {
			answers, extras = couchDB.CNAME(qrecord, &z)
		}
	case dns.TypeNS:
		answers, extras = couchDB.NS(qrecord, &z)
	case dns.TypeCNAME:
		answers, extras = couchDB.CNAME(qrecord, &z)
	case dns.TypeAAAA:
		answers, extras = couchDB.AAAA(qrecord, &z)
	case dns.TypeTXT:
		answers, extras = couchDB.TXT(qrecord, &z)
	case dns.TypeMX:
		answers, extras = couchDB.MX(qrecord, &z)
	case dns.TypeSOA:
		answers, extras = couchDB.SOA(qrecord, &z)
	case dns.TypeSRV:
		answers, extras = couchDB.SRV(qrecord, &z)
	case dns.TypeCAA:
		answers, extras = couchDB.CAA(qrecord, &z)
	default:
		return dns.RcodeNotImplemented, nil // NotImp    - Not Implemented
	}

	if answers == nil {
		return dns.RcodeServerFailure, nil // ServFail  - Server Failure
	}

	// response
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	// write response
	state.SizeAndDo(m)
	m = state.Scrub(m)
	w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

// Name of the plugin, returns "couchdb".
func (couchDB *CouchDB) Name() string { return "couchdb" }
