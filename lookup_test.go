package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	couchdb "github.com/rhinoman/couchdb-go"

	"github.com/miekg/dns"
)

var zone = Zone{
	ID: "example.com.",
	Data: []Record{
		{
			Name: "",
			Data: "127.0.0.2",
			Type: dns.TypeA,
			TTL:  3600,
		},
		{
			Name: "www",
			Data: "example.com.",
			Type: dns.TypeCNAME,
			TTL:  3600,
		},
		{
			Name: "",
			Data: "10 mail.example.com.",
			Type: dns.TypeMX,
			TTL:  3600,
		},
		{
			Name: "",
			Data: "ns1.example.com.",
			Type: dns.TypeNS,
			TTL:  3600,
		},
		{
			Name: "",
			Data: "ns2.example.com.",
			Type: dns.TypeNS,
			TTL:  3600,
		},
		{
			Name: "",
			Data: "v=spf1 +a +mx +ip4:127.0.0.2 ~all",
			Type: dns.TypeTXT,
			TTL:  3600,
		},
		{
			Name: "",
			Data: "ns.example.com. webmaster.example.com. 2018071304 3600 7200 1209600 86400",
			Type: dns.TypeSOA,
			TTL:  3600,
		},
		{
			Name: "_sip._tcp",
			Data: "20 0 5060 backup.example.com.",
			Type: dns.TypeSRV,
			TTL:  3600,
		},
	},
}

var tests = []test.Case{
	{
		Qname: "example.com.",
		Qtype: dns.TypeA,
		Answer: []dns.RR{
			test.A("example.com. 3600 IN A 127.0.0.2"),
		},
	},
	{
		Qname: "www.example.com.",
		Qtype: dns.TypeCNAME,
		Answer: []dns.RR{
			test.CNAME("www.example.com. 3600 IN CNAME example.com."),
		},
	},
	{
		Qname: "example.com.",
		Qtype: dns.TypeMX,
		Answer: []dns.RR{
			test.MX("example.com. 3600 IN MX 10 mail.example.com."),
		},
	},
	{
		Qname: "example.com.",
		Qtype: dns.TypeNS,
		Answer: []dns.RR{
			test.NS("example.com. 3600 IN NS ns1.example.com."),
			test.NS("example.com. 3600 IN NS ns2.example.com."),
		},
	},
	{
		Qname: "example.com.",
		Qtype: dns.TypeTXT,
		Answer: []dns.RR{
			test.TXT("example.com. 3600 IN TXT \"v=spf1 +a +mx +ip4:127.0.0.2 ~all\""),
		},
	},
	{
		Qname: "example.com.",
		Qtype: dns.TypeSOA,
		Answer: []dns.RR{
			test.SOA("example.com. 3600 IN SOA ns.example.com. webmaster.example.com. 2018071304 3600 7200 1209600 86400"),
		},
	},
	{
		Qname: "_sip._tcp.example.com.",
		Qtype: dns.TypeSRV,
		Answer: []dns.RR{
			test.SRV("_sip._tcp.example.com. 3600 IN SRV 20 0 5060 backup.example.com."),
		},
	},
}

func TestAnswer(t *testing.T) {

	fmt.Println("lookup test")

	couchDB := New()
	couchDB.Address = "127.0.0.1"
	couchDB.Port = 5984
	couchDB.DBname = "testzones"

	connection, err := couchdb.NewConnection(couchDB.Address, couchDB.Port, time.Duration(500*time.Millisecond))
	if err != nil {
		log.Println("[ERROR] Unable to connect")
		t.Fail()
		return
	}

	couchDB.Connection = connection
	couchDB.DB = connection.SelectDB(couchDB.DBname, nil)

	if err := couchDB.DB.DbExists(); err == nil {
		log.Println("[ERROR] already exists database testzones")
		t.Fail()
		return
	}

	if err := couchDB.Connection.CreateDB(couchDB.DBname, nil); err != nil {
		log.Println("[ERROR] unable to create database")
		t.Fail()
		return
	}

	if err = connection.Ping(); err != nil {
		log.Println("[ERROR] unable to connect")
		t.Fail()
		return
	}

	_, z := couchDB.Find(zone.ID)
	if z.Rev != "" {
		log.Println("[ERROR] already exists example.com zone")
		t.Fail()
		return
	}

	// save test zone example.com if not exists
	b, _ := json.Marshal(zone)
	body := strings.NewReader(string(b))

	r := fmt.Sprintf("http://%s:%d/%s", couchDB.Address, couchDB.Port, couchDB.DBname)
	req, err := http.NewRequest("POST", r, body)
	if err != nil {
		log.Println("[ERROR] error en couchdb", err)
		t.Fail()
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("[ERROR] error en couchdb", err)
		t.Fail()
	}
	defer resp.Body.Close()

	// tests
	for _, tc := range tests {

		m := tc.Msg()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		couchDB.ServeDNS(ctxt, rec, m)

		resp := rec.Msg
		if resp == nil {
			resp = new(dns.Msg)
		}
		test.SortAndCheck(t, resp, tc)
	}

	if err := couchDB.Connection.DeleteDB(couchDB.DBname, nil); err != nil {
		log.Println("[ERROR] unable to delete database")
		t.Fail()
		return
	}
}

var ctxt context.Context
