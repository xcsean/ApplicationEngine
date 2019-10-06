#!/bin/bash

PROTOC=$GOPATH/bin/protoc
PROTOCGO=$GOPATH/bin/protoc-gen-go

$PROTOC --plugin=protoc-gen-go=$PROTOCGO --go_out=plugins=grpc:../protocol/ -I../../protocol/ getcd.proto
$PROTOC --plugin=protoc-gen-go=$PROTOCGO --go_out=plugins=grpc:../protocol/ -I../../protocol/ gconnd.proto packet.proto
$PROTOC --plugin=protoc-gen-go=$PROTOCGO --go_out=plugins=grpc:../protocol/ -I../../protocol/ ghost.proto packet.proto
