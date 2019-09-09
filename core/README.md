## Application Engine Core

### Features
* aaa
* bbb
* ccc

### Compiler
* Go 1.11+
* Legacy GOPATH mode

### Download
```sh
go get -u github.com/xcsean/ApplicationEngine
```

### Dependencies Installation
* Zerolog
```sh
go get -u github.com/rs/zerolog/log
```
* Radix
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
* gRPC
```sh
go get -u google.golang.org/grpc
cd $GOPATH/src/
go install google.golang.org/grpc
go install github.com/golang/protobuf/protoc-gen-go
```
* JWT
```sh
go get -u github.com/dgrijalva/jwt-go
```

### Compile
* Linux
```sh
cd $GOPATH/src/github.com/xcsean/ApplicationEngine
cd core/bin && sh -ex build.sh
```
* Windows
```cmd
cd %GOPATH%\src\github.com\xcsean\ApplicatonEngine
cd protocol
convert_win.bat
cd ..
cd core\bin
build_win.bat
```
* Mac
```sh
cd $GOPATH/src/github.com/xcsean/ApplicationEngine
cd core/bin && sh -ex build_mac.sh
```

### Deployment
* Centos 7.x
* Ansible