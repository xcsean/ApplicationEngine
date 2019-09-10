@echo off

chcp 65001

rem convert the errno
%GOPATH%\bin\protoc.exe --plugin=protoc-gen-go=%GOPATH%\bin\protoc-gen-go.exe --go_out=..\core\protocol\gerrno\ gerrno.proto

rem convert the getcd
%GOPATH%\bin\protoc.exe --plugin=protoc-gen-go=%GOPATH%\bin\protoc-gen-go.exe --go_out=plugins=grpc:..\core\protocol\getcd\ getcd.proto