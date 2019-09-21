# Core Service

## Directory
* getcd, means 'go-etc-daemon', provides configuration service
* gconnd, means 'go-connection-daemon', provides connection-oriented features
* gconnless, means 'go-connection-less', provides connectionless features
* ghostd, means 'go-host-daemon', provides packet forwarding and storage features
* ws2tcp provides websocket to tcp conversion

## getcd
* service configuration
* server configuration
* global configuration
* protocol limit configuration

## gconnd
* connection-oriented
* packet forwarding to backend or frontend by sequence
* broadcast to the frontend(s) specified
* notify to the frontend(s) specified
* enter/leave notification to backend
* keep-alive check such as ping/pong
* traffic statistics

## ghostd
* packet forwarding between logic module or process with client(s)
* important data save to database and load interface
* protocol frequency limitation and protection
* overload protection
