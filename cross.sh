#!/bin/bash

set -ex

export CGO_ENABLED=0

NAMEPREFIX=$1
EXESUFFIX=$2

go generate pkg/common/rev.go

CGO_ENABLED=0 go build -o build/$NAMEPREFIX/ra$EXESUFFIX cmd/ra/main.go
CGO_ENABLED=0 go build -o build/$NAMEPREFIX/raise$EXESUFFIX cmd/raise/main.go
CGO_ENABLED=0 go build -o build/$NAMEPREFIX/raised$EXESUFFIX cmd/raised/main.go