package couchdb

import (
	"log"
	"strconv"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/mholt/caddy"
	couchdb "github.com/rhinoman/couchdb-go"
)

func init() {
	caddy.RegisterPlugin("couchdb", caddy.Plugin{
		// ServerType is "dns" because couchDB is a DNS server;
		ServerType: "dns",
		// Action is the name of the setup function tthat takes
		// care of the parsing of the Corefile. Its job is to return
		// a type that implements the plugin.Handler interface.
		// So whenever the Corefile parser sees "couchdb", setupCouchDB is called.
		Action: setupCouchDB,
	})
}

// setupCouchDB configures the couchdb plugin, the format is:
//	couchdb {
// 		address  COUCHDB ADDRESS
// 		port     COUCHDB PORT
// 		username USERNAME BASIC AUTH
// 		password PASSWORD BASIC AUTH
//	}
func setupCouchDB(c *caddy.Controller) error {

	couchDB, err := couchDBParse(c)
	if err != nil {
		return plugin.Error("couchdb", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		couchDB.Next = next
		return couchDB
	})

	return nil

}

// couchDBParse parse arguments
func couchDBParse(c *caddy.Controller) (*CouchDB, error) {

	couchDB := New()
	var username, password string

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "address":
					if !c.NextArg() {
						return couchDB, c.ArgErr()
					}
					couchDB.Address = c.Val()
				case "port":
					if !c.NextArg() {
						return couchDB, c.ArgErr()
					}
					port, err := strconv.Atoi(c.Val())
					if err == nil {
						couchDB.Port = port
					}
				case "dbname":
					if !c.NextArg() {
						return couchDB, c.ArgErr()
					}
					couchDB.DBname = c.Val()
				case "username":
					if !c.NextArg() {
						return couchDB, c.ArgErr()
					}
					username = c.Val()
				case "password":
					if !c.NextArg() {
						return couchDB, c.ArgErr()
					}
					password = c.Val()
				default:
					if c.Val() != "}" {
						return couchDB, c.Errf("unknown property '%s'", c.Val())
					}
				}
				if !c.Next() {
					break
				}

			}
		}
	}

	if username != "" && password != "" {
		couchDB.BasicAuth = &couchdb.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	if err := couchDB.connect(); err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, c.Err("Unable to connect")
	}

	log.Printf("[INFO] couchDB connected")
	return couchDB, nil
}
