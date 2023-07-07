#!/bin/bash

BASEDIR="$(dirname "${BASH_SOURCE[0]}")"

cd $BASEDIR/..
go build -tags "debug" soko.go
./soko -profile test &

sleep 1
go tool pprof -http localhost: soko localhost:6060/debug/pprof/heap

while true ; do
    if dialog --yesno "Reopen profile?" 10 60 ; then
        go tool pprof -http localhost: soko localhost:6060/debug/pprof/heap
    else
        break
    fi
done

clear
rm -rf soko
killall soko
