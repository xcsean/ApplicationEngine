## Application Engine Core

### Features
* aaa
* bbb
* ccc

### Compiler
* Go 1.11+
* Legacy GOPATH mode

### Dependencies Installation
* Zerolog
```sh
go get -u github.com/rs/zerolog/log
```
* Radix
```sh
go get -u github.com/mediocregopher/radix 
```
* JWT
```sh
go get -u github.com/dgrijalva/jwt-go
```

### Compile & Test
```
go get -u github.com/xcsean/ApplicationEngine
cd $GOPATH/src/github.com/xcsean/ApplicationEngine
cd core/bin && sh -ex build.sh
```

### Deployment
* Centos 7.x
* Ansible