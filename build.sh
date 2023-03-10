#!/bin/bash

set -ex

go generate pkg/common/rev.go

CGO_ENABLED=0 go build -o build/ra cmd/ra/main.go
CGO_ENABLED=0 go build -o build/raise cmd/raise/main.go
CGO_ENABLED=0 go build -o build/raised cmd/raised/main.go