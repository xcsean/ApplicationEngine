# Core Service

## Directory
* getcd, means 'go-etc-daemon', provides configuration service
* gconnd, means 'go-connection-daemon', provides connection-oriented features such as broadcast/notification
* gconnless, means 'go-connection-less', provides connectionless features
* ghost, means 'go-host', provides packet forwarding and storage features
* ws2tcp provides websocket to tcp conversion

## getcd
* service configuration
* server configuration
* global configuration
* protocol limit configuration

## gconnd
* connection-oriented
* packet forward to backend or frontend by sequence
* broadcast to the frontend(s) specified
* notify to the frontend(s) specified
* enter/leave notification to backend
* keep-alive such as ping/pong
* traffic statistics

## ghost
* packet forwarding between logic module or process and client(s)
* assets data save to database and load interface
* protocol frequency limitation and protection
* overload protection
