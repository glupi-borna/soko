#!/bin/bash

BASEDIR="$(dirname "${BASH_SOURCE[0]}")"
cd $BASEDIR/..

function build() {
    killall -s TERM soko
    if go build ./soko.go ; then
        ./soko -x -8 -y 25 -anchor top-right test &
    fi
}

trap 'killall -s TERM soko ; rm soko ; exit 0' EXIT HUP INT TERM
build

inotifywait . --monitor -e modify --include "\.go" -r |
    while read dir action file ; do
        echo "$file changed, recompiling..."
        build
    done
