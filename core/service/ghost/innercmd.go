package main

import "net"

type innerCmd struct {
	cmdID uint8
	err   error
	c     net.Conn
	ch    chan<- *innerCmd
}

const (
	innerCmdRegisterVM    = 101
	innerCmdUnregisterVM  = 102
	innerCmdLoadUserAsset = 103
	innerCmdSaveUserAsset = 104
	innerCmdSendPacket    = 105
)

func (ic *innerCmd) getRPCCmd() chan<- *innerCmd {
	return ic.ch
}

func newRPCCmd(cmdID uint8, ch chan<- *innerCmd) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		ch:    ch,
	}
}
