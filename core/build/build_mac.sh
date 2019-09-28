#!/bin/bash

./convert_mac.sh

go build ../service/getcd
go build ../service/gconnd
go build ../service/ghost
go build ../sample/loop
go build ../sample/vmcli
