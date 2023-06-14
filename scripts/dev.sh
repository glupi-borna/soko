#!/bin/bash

BASEDIR="$(dirname "${BASH_SOURCE[0]}")"
cd $BASEDIR/..

function build() {
    killall wiggo
    if go build ./wiggo.go ; then
        ./wiggo &
    fi
}

trap 'killall wiggo ; rm wiggo ; exit 0' EXIT HUP INT TERM
build

inotifywait . --monitor -e modify --include "\.go" -r |
    while read dir action file ; do
        echo "$file changed, recompiling..."
        build
    done
