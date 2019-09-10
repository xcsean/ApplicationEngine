#!/bin/bash

# convert proto firstly
./convert_mac.sh

# build for linux 64
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build ../service/getcd
