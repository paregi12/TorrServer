module github.com/paregi12/torrentserver

go 1.21.0

require github.com/paregi12/torrentserver/server v0.0.0
replace github.com/paregi12/torrentserver/server => ./server
replace github.com/wlynxg/anet => ./server/anet-dummy
