package couchdb

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"

	couchdb "github.com/rhinoman/couchdb-go"
)

// CouchDB is the implementation of the "couchdb" CoreDNS plugin.
type CouchDB struct {
	Next plugin.Handler // Next handler in the list of plugins.

	Connection *couchdb.Connection
	DB         *couchdb.Database

	// Addr is the address of the CouchDB server.
	Address string
	Port    int
	DBname  string

	// BasicAuth for couchdb
	BasicAuth couchdb.Auth
}

// Record defines a DNS Record in JSON
// such as RFC 8427 (Representing DNS Messages in JSON)
// https://tools.ietf.org/html/draft-hoffman-dns-in-json-16
type Record struct {
	Name string `json:"name"`
	Data string `json:"data"`
	Type uint16 `json:"type"`
	TTL  uint32 `json:"TTL"`
}

// Zone defines a DNS Zone stored in couchDB.
type Zone struct {
	ID   string   `json:"_id"`
	Rev  string   `json:"_rev,omitempty"`
	Data []Record `json:"data,omitempty"`
}

const (
	defaultAddress = "127.0.0.1"
	defaultPort    = 5984
	defaultDBname  = "zones"
)

// New constructs a new instance of a couchdb plugin.
func New() *CouchDB {
	return &CouchDB{
		Address: defaultAddress,
		Port:    defaultPort,
		DBname:  defaultDBname,
	}
}

// connect establish connection to CouchDB server.
func (couchDB *CouchDB) connect() error {

	var timeout = time.Duration(500 * time.Millisecond)
	connection, err := couchdb.NewConnection(couchDB.Address, couchDB.Port, timeout)
	if err != nil {
		return err
	}

	couchDB.DB = connection.SelectDB(couchDB.DBname, couchDB.BasicAuth)

	if err := couchDB.DB.DbExists(); err != nil {
		return err
	}

	couchDB.Connection = connection

	if err = connection.Ping(); err != nil {
		return err
	}

	return nil
}

// loadZone loads a zone by zone.ID.
func (couchDB *CouchDB) loadZone(zone *Zone) error {

	db := couchDB.DB
	_, err := db.Read(zone.ID, &zone, nil)
	if err != nil {
		return err
	}
	// if zone exists
	return nil
}

// A returns A DNS records that matches with the query.
func (couchDB *CouchDB) A(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {

		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeA != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.A)

		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		var ip net.IP
		if ip = net.ParseIP(rec.Data); ip == nil {
			continue
		}
		r.A = ip
		answers = append(answers, r)
	}
	return
}

// AAAA returns AAAA DNS records that matches with the query.
func (couchDB *CouchDB) AAAA(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {
		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeAAAA != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.A)

		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		var ip net.IP
		if ip = net.ParseIP(rec.Data); ip == nil {
			continue
		}
		r.A = ip
		answers = append(answers, r)
	}
	return
}

// CNAME returns CNAME DNS records that matches with the query.
func (couchDB *CouchDB) CNAME(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {

		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeCNAME != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.CNAME)

		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		r.Target = dns.Fqdn(rec.Data)
		answers = append(answers, r)
	}
	return
}

// NS returns NS DNS records that matches with the query.
func (couchDB *CouchDB) NS(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {
		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeNS != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.NS)
		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeNS,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		r.Ns = dns.Fqdn(rec.Data)
		answers = append(answers, r)
		extras = append(extras, couchDB.hosts(r.Ns)...)
	}
	return
}

// TXT returns TXT DNS records that matches with the query.
// Data record format such as RFC 4408 https://tools.ietf.org/html/rfc4408
func (couchDB *CouchDB) TXT(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {
		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeTXT != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.TXT)
		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		r.Txt = split255(rec.Data)
		answers = append(answers, r)
	}
	return
}

// MX returns MX DNS records that matches with the query.
// Data record format such as RFC 974 https://tools.ietf.org/html/rfc974
func (couchDB *CouchDB) MX(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {
		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeMX != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.MX)
		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeMX,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		mx := strings.Fields(rec.Data)
		if len(mx) < 2 {
			continue
		}

		preference, err := strconv.Atoi(mx[0])
		if err != nil {
			continue
		}

		r.Preference = uint16(preference)
		r.Mx = mx[1]

		answers = append(answers, r)
		extras = append(extras, couchDB.hosts(r.Mx)...)
	}
	return
}

// SOA returns SOA record that matches with the query.
func (couchDB *CouchDB) SOA(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {
		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeSOA != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.SOA)
		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		soa := strings.Fields(rec.Data)
		if len(soa) < 7 {
			continue
		}

		r.Ns, r.Mbox, r.Serial, r.Refresh, r.Retry, r.Expire, r.Minttl = parseSOA(soa[0:]...)

		answers = append(answers, r)
	}
	return
}

// parseSOA unpack and parse SOA from json data record.
func parseSOA(args ...string) (string, string, uint32, uint32, uint32, uint32, uint32) {

	arr := []uint32{uint32(time.Now().Unix()), 86400, 7200, 3600, 14400}

	for i, v := range args[2:] {
		if d, err := strconv.Atoi(v); err == nil {
			arr[i] = uint32(d)
		}
	}

	return args[0], args[1], arr[0], arr[1], arr[2], arr[3], arr[4]

}

// SRV returns SRV DNS records that matches with the query.
// Data record format such as RFC 2782 https://tools.ietf.org/html/rfc2782
// RDATA: Priority Weight Port Target
func (couchDB *CouchDB) SRV(query string, z *Zone) (answers, extras []dns.RR) {
	for _, rec := range z.Data {
		name := query
		if rec.Name != name {
			continue
		}

		if dns.TypeSRV != rec.Type {
			continue
		}

		if name != "" {
			name = fmt.Sprintf("%s.%s", name, z.ID)
		} else {
			name = z.ID
		}

		r := new(dns.SRV)
		r.Hdr = dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    rec.TTL,
		}

		srv := strings.Fields(rec.Data)

		if len(srv) < 4 {
			continue
		}

		r.Priority, r.Weight, r.Port, r.Target = parseSRV(srv[0:]...)
		answers = append(answers, r)
		extras = append(extras, couchDB.hosts(r.Target)...)
	}
	return
}

// parseSRV unpack and parse SRV from json data record.
func parseSRV(args ...string) (uint16, uint16, uint16, string) {

	var arr [3]uint16

	for i, v := range args {
		if d, err := strconv.Atoi(v); err == nil {
			arr[i] = uint16(d)
		}
	}

	return arr[0], arr[1], arr[2], args[3]

}

// Split255 splits a string into 255 byte chunks.
func split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	sx := []string{}
	p, i := 0, 255
	for {
		if i <= len(s) {
			sx = append(sx, s[p:i])
		} else {
			sx = append(sx, s[p:])
			break

		}
		p, i = p+255, i+255
	}

	return sx
}

// Find DNS zone by qname.
func (couchDB *CouchDB) Find(qname string) (qrecord string, z Zone) {

	qnameParts := strings.Split(qname, ".")

	z = Zone{}
	var record []string

	for {
		z.ID = qname
		if err := couchDB.loadZone(&z); err == nil {
			break
		}

		if len(qnameParts) <= 2 {
			break
		}

		record = append(record, qnameParts[0])
		qnameParts = qnameParts[1:]
		qname = strings.Join(qnameParts, ".")
	}

	qrecord = strings.Join(record, ".") // example: www.

	return
}

// hosts returns additional  
func (couchDB *CouchDB) hosts(qname string) (answers []dns.RR) {

	qrecord, z := couchDB.Find(qname)
	if z.Rev == "" {
		return
	}

	a, _ := couchDB.A(qrecord, &z)
	answers = append(answers, a...)

	aaaa, _ := couchDB.AAAA(qrecord, &z)
	answers = append(answers, aaaa...)

	cname, _ := couchDB.CNAME(qrecord, &z)
	answers = append(answers, cname...)

	return
}
