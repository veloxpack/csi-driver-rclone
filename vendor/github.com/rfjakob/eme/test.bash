#!/bin/bash -eu

go build
go test -v "$@"
go vet -all .
if command -v staticcheck >/dev/null ; then
    staticcheck ./...
else
    echo "staticcheck not installed, skipping"
fi
