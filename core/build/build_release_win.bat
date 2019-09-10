@echo off

rem convert proto firstly
call convert_win.bat

rem build for linux 64
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build ..\service\getcd