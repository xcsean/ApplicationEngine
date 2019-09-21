# Core Service

## Directory
* getcd, means 'go-etc-daemon', provides configuration service
* gconnd, means 'go-connection-daemon', provides connection-oriented features such as broadcast/notification
* gconnless, means 'go-connection-less', provides connectionless features
* globby, means 'go-lobby', provides lobby features
* ws2tcp provides websocket to tcp conversion

## getcd
* service configuration
* server configuration
* global configuration
* protocol limit configuration

## gconnd
* connection-oriented
* packet forward to backend or frontend
* broadcast to the frontend(s) specified
* notify to the frontend(s) specified
* enter/leave notification to backend
* keep-alive such as ping/pong
* traffic statistics

## globby
* packet forwarding between logic module or process and client(s)
* player assets data save to database and load interface
* protocol frequency limitation and protection
* overload protection
* 3rd-sdk import
