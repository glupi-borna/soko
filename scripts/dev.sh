#!/bin/bash

BASEDIR="$(dirname "${BASH_SOURCE[0]}")"
cd $BASEDIR/..

function build() {
    killall soko
    if go build ./soko.go ; then
        ./soko test &
    fi
}

trap 'killall soko ; rm soko ; exit 0' EXIT HUP INT TERM
build

inotifywait . --monitor -e modify --include "\.go" -r |
    while read dir action file ; do
        echo "$file changed, recompiling..."
        build
    done
