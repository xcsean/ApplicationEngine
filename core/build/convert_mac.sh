#!/bin/bash

PROTOC=$GOPATH/bin/protoc
PROTOCGO=$GOPATH/bin/protoc-gen-go

$PROTOC --plugin=protoc-gen-go=$PROTOCGO --go_out=plugins=grpc:../protocol/getcd/ -I../../protocol/ getcd.proto
$PROTOC --plugin=protoc-gen-go=$PROTOCGO --go_out=plugins=grpc:../protocol/ghost/ -I../../protocol/ ghost.proto