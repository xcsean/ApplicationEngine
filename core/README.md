# Application Engine Core

## Directory

* 'database' contain registry & ghost sql scripts
* 'protocol' contain services protocol go files
* 'sample' contain some samples such as host/vm/client
* 'service' contain services source, such as getcd/gconnd/ghost
* 'shared' contain go source shared by services

## Compiler

* Go 1.11+
* Legacy GOPATH mode

## Download

```sh
go get -u github.com/xcsean/ApplicationEngine
```

## Dependencies Installation

* Log
```sh
go get -u github.com/rs/zerolog/log
```
* Redis
```sh
go get -u github.com/mediocregopher/radix
```
* MySQL
```sh
go get -u github.com/go-sql-driver/mysql
```
* Protobuf
```sh
cd $GOPATH/bin
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.9.1/protoc-3.9.1-linux-x86_64.zip -O protoc.zip
mkdir -p protoc && unzip protoc.zip -d protoc && cp protoc/bin/protoc . && rm -rf protoc && rm -f protoc.zip
```
* RPC
```sh
go get -u google.golang.org/grpc
cd $GOPATH/src/
go install google.golang.org/grpc
go install github.com/golang/protobuf/protoc-gen-go
```
* Token
```sh
go get -u github.com/dgrijalva/jwt-go
```
* Timer
```sh
go get -u github.com/RussellLuo/timingwheel
```
* Cross-Compile
```sh
go get -u golang.org/x/sys
```

## Compile for testing

* Linux
```sh
cd $GOPATH/src/github.com/xcsean/ApplicationEngine
cd core/build && sh -ex build.sh
```
* Windows
```cmd
cd %GOPATH%\src\github.com\xcsean\ApplicatonEngine
cd core\build
build_win.bat
```
* Mac
```sh
cd $GOPATH/src/github.com/xcsean/ApplicationEngine
cd core/build && sh -ex build_mac.sh
```

## Compile for release

* Linux
```sh
cd $GOPATH/src/github.com/xcsean/ApplicationEngine
cd core/build && sh -ex build_release.sh
```
* Windows
```cmd
cd %GOPATH%\src\github.com\xcsean\ApplicatonEngine
cd core\build
build_release_win.bat
```
* Mac
```sh
cd $GOPATH/src/github.com/xcsean/ApplicationEngine
cd core/build && sh -ex build_release_mac.sh
```

## Deployment

* Centos 7.6
* Ansible 2.8