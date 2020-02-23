1.- Clone coredns repo

`git clone https://github.com/coredns/coredns.git`

2.- Install golang version >= 1.12

`https://golang.org/doc/install?download=go1.13.8.linux-amd64.tar.gz`

3.- Clone couchdb repo

`https://github.com/soysuperadmin/couchdb.git`

4.- Copy couchdb repo on coredns/plugin folder

5.- Enable plugin on `plugin.cfg`. Add `couchdb:couchdb` at the end of the file

```bash
# Directives are registered in the order they should be executed.
#
# Ordering is VERY important. Every plugin will feel the effects of all other
# plugin below (after) them during a request, but they must not care what plugin
# above them are doing.

# How to rebuild with updated plugin configurations: Modify the list below and
# run `go generate && go build`

# The parser takes the input format of:
#
#     <plugin-name>:<package-name>
# Or
#     <plugin-name>:<fully-qualified-package-name>
#
# External plugin example:
#
# log:github.com/coredns/coredns/plugin/log
# Local plugin example:
# log:log

metadata:metadata
cancel:cancel
tls:tls
reload:reload
nsid:nsid
bufsize:bufsize
root:root
bind:bind
debug:debug
trace:trace
ready:ready
health:health
pprof:pprof
prometheus:metrics
errors:errors
log:log
dnstap:dnstap
acl:acl
any:any
chaos:chaos
loadbalance:loadbalance
cache:cache
rewrite:rewrite
dnssec:dnssec
autopath:autopath
template:template
transfer:transfer
hosts:hosts
route53:route53
azure:azure
clouddns:clouddns
federation:github.com/coredns/federation
k8s_external:k8s_external
kubernetes:kubernetes
file:file
auto:auto
secondary:secondary
etcd:etcd
loop:loop
forward:forward
grpc:grpc
erratic:erratic
whoami:whoami
on:github.com/caddyserver/caddy/onevent
sign:sign
couchdb:couchdb

```

6.- Make sure you have build-essential

`sudo apt install build-essential`

7.- Run `make`

8.- Test coredns is build with couchdb plugin with `./coredns -plugins | grep couchdb`

```bash
root@dns:~/coredns-1.6.7$ ./coredns -plugins |grep couchdb
dns.couchdb

```

It's done!!