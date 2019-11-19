@echo off

call convert_win.bat

go build ../service/getcd
go build ../service/gconnd
go build ../service/ghost
go build ../sample/vmcli
