#!/bin/bash

BASEDIR="$(dirname "${BASH_SOURCE[0]}")"

cd $BASEDIR/..
go build wiggo.go
./wiggo -profile test &
sleep 1
go tool pprof -http localhost: wiggo localhost:6060/debug/pprof/heap
rm -rf wiggo
killall wiggo
