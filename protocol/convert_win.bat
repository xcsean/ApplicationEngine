@echo off

chcp 65001

%GOPATH%\bin\protoc.exe --plugin=protoc-gen-go=%GOPATH%\bin\protoc-gen-go.exe --go_out=..\core\proto\getcd\ getcd.proto