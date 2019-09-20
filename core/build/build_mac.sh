#!/bin/bash

./convert_mac.sh

go build ../service/getcd
go build ../service/gconnd
go build ../sample/loop
