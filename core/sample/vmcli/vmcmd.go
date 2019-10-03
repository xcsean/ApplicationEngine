package main

import "github.com/xcsean/ApplicationEngine/core/protocol/ghost"

type hostCmd struct {
	Type uint8
	Pkt  *ghost.GhostPacket
}

const (
	hostCmdRemoteClosed = 1
	hostCmdNotifyStatus = 2
	hostCmdNotifyPacket = 3
)
