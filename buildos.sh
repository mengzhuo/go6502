#!/bin/bash

set -ue

rm -rf zhuos/src/*.out

for f in zhuos/src/*.s
do
    	if [ "$f" != 'zhuos/src/symbols.s' ]; then
	    echo $f
	    go run cmd/asm/main.go -i $f
	fi
done

go run cmd/link/main.go z.dat zhuos/src/*.out
