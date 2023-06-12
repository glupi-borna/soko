#!/bin/bash

BASEDIR="$(dirname "${BASH_SOURCE[0]}")"

cd $BASEDIR/..
go build wiggo.go
./wiggo -cpuprofile wiggo.prof -timeout 100000
go tool pprof -http localhost: wiggo wiggo.prof
rm -rf wiggo
rm -rf wiggo.prof
