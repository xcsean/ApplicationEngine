@echo off

chcp 65001

rem convert the getcd
%GOPATH%\bin\protoc.exe --plugin=protoc-gen-go=%GOPATH%\bin\protoc-gen-go.exe --go_out=plugins=grpc:..\protocol\getcd\ -I..\..\protocol\ getcd.proto