#!/bin/bash

set -ex

go generate pkg/proto/proto.go

go build -o build/ra cmd/ra/main.go
go build -o build/raise cmd/raise/main.go
go build -o build/raised cmd/raised/main.go