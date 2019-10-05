package main

import "github.com/xcsean/ApplicationEngine/core/protocol"

type hostCmd struct {
	Type uint8
	Pkt  *protocol.GhostPacket
}

const (
	hostCmdRemoteClosed = 1
	hostCmdNotifyStatus = 2
	hostCmdNotifyPacket = 3
)
