#!/bin/bash

BASEDIR="$(dirname "${BASH_SOURCE[0]}")"

cd $BASEDIR/..
go build soko.go
./soko -profile test &
sleep 1
go tool pprof -http localhost: soko localhost:6060/debug/pprof/heap
rm -rf soko
killall soko
