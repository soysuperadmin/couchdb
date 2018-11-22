## Description

The **couchdb plugin** is a plugin to compile with CoreDNS that enables reading zone data from CouchDB server.

## Syntax

Use [rhinoman/couchdb-go](https://github.com/rhinoman/couchdb-go) to handle database.

~~~ txt
couchdb 
~~~

couchdb loads authoritative zones from couchdb server

Address and port will default to local couchdb server (localhost:5984/) and it's possible 
configure HTTP basic authentication by providing username and password.

~~~ txt
couchdb {
	address  COUCHDB ADDRESS
	port     COUCHDB PORT
    dbname   COUCHDB DBNAME
	username USERNAME BASIC AUTH
	password PASSWORD BASIC AUTH
}
~~~

all parameters 

* **address** defaults to localhost.
* **port** defaults to 5984.
* **dbname** defaults to zones.
* **username** username to basic auth. Defaults none.
* **password** password to basic auth. Defaults none.

This plugin is intended to appear at the end of the plugin list, usually
near the `proxy` plugin declaration.

## Example

Enable couchdb plugin at localhost with port 5984

~~~ corefile
. {
	couchdb {
		address 127.0.0.1
		port 5984
	}
	errors
}
~~~

### zone format in couchdb db

Data record format representing DNS Messages in JSON 
such as RFC 8427 (Draft) https://tools.ietf.org/html/draft-hoffman-dns-in-json-16

Each zones is stored in database "zones" as a document.

First, must create database zones:

```
# curl -H 'Content-Type: application/json' -X PUT http://127.0.0.1:5984/zones
{"ok":true}
```

Add some documents:

```
# curl -0 http://127.0.0.1:5984/zones1 \
-H 'Content-Type: application/json' \
-d @- << EOF
{
    "_id": "example.com.",
    "data": [
        {
            "name": "",
            "data": "127.0.0.2",
            "TTL": 3600,
            "type": 1
        },
        {
            "name": "ns1",
            "data": "127.0.0.2",
            "TTL": 3600,
            "type": 1
        },
        {
            "name": "ns2",
            "data": "127.0.0.2",
            "TTL": 3600,
            "type": 1
        },        
        {
            "name": "www",
            "data": "example.com.",
            "TTL": 3600,
            "type": 5
        },
        {
            "name": "",
            "data": "ns1.example.com.",
            "TTL": 3600,
            "type": 2
        },
        {
            "name": "",
            "data": "ns2.example.com.",
            "TTL": 3600,
            "type": 2
        },
        {
            "name": "",
            "data": "0 example.com.",
            "TTL": 3600,
            "type": 15
        },
        {
            "name": "",
            "data": "\"v=spf1 +a +mx +ip4:127.0.0.2 ~all\"",
            "TTL": 3600,
            "type": 16
        },
        {
            "name": "",
            "data": "ns1.example.com. webmaster.example.com. 2018071304 3600 7200 1209600 86400",
            "TTL": 3600,
            "type": 6
        },
        {
            "name": "_sip._tcp",
            "data": "20 0 5060 www.example.com.",
            "TTL": 3600,
            "type": 33
        }
    ]
}
EOF
{"ok":true,"id":"example.com.","rev":"1-223911e169ce700d846e62ee4ff2c087"}
```

Do some queries:

```
# --------------
# A 
# --------------
# dig @127.0.0.1 example.com a
...
;; ANSWER SECTION:
example.com.		3600	IN	A	127.0.0.2

# --------------
# MX
# --------------
# dig @127.0.0.1 example.com mx
...
;; ANSWER SECTION:
example.com.		3600	IN	MX	0 example.com.

;; ADDITIONAL SECTION:
example.com.		3600	IN	A	127.0.0.2

# --------------
# NS
# --------------
# dig @127.0.0.1 example.com ns
...
;; ANSWER SECTION:
example.com.		3600	IN	NS	ns1.example.com.
example.com.		3600	IN	NS	ns2.example.com.

;; ADDITIONAL SECTION:
ns1.example.com.	3600	IN	A	127.0.0.2
ns2.example.com.	3600	IN	A	127.0.0.2

# --------------
# SRV
# --------------
# dig @127.0.0.1 _sip._tcp.example.com srv
...
;; ANSWER SECTION:
_sip._tcp.example.com.	3600	IN	SRV	20 0 5060 www.example.com.

;; ADDITIONAL SECTION:
www.example.com.	3600	IN	CNAME	example.com.

# --------------
# SOA
# --------------
# dig @127.0.0.1 example.com soa
...
;; ANSWER SECTION:
example.com.		3600	IN	SOA	ns1.example.com. webmaster.example.com. 2018071304 3600 7200 1209600 86400

```
