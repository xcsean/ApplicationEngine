package main

import (
	"net"
)

const (
	innerCmdClientListenStart = 1
	innerCmdClientListenStop  = 2
	innerCmdClientIncoming    = 3
	innerCmdClientLeave       = 4
	innerCmdClientUp          = 5

	innerCmdServerListenStart  = 101
	innerCmdServerListenStop   = 102
	innerCmdServerIncoming     = 103
	innerCmdServerLeave        = 104
	innerCmdServerMasterSet    = 105
	innerCmdServerBroadcast    = 106
	innerCmdServerKick         = 107
	innerCmdServerKickAll      = 108
	innerCmdServerSetRoute     = 109
	innerCmdServerBroadcastAll = 110

	innerCmdAdminListenStart = 201
	innerCmdAdminListenStop  = 202
	innerCmdAdminKick        = 203
	innerCmdAdminKickAll     = 204
)

type innerCmd struct {
	cmdID  uint8
	c      net.Conn
	n      uint64
	err    error
	str    string
	hdr    []byte
	body   []byte
}

func (ic *innerCmd) getCmdID() uint8 {
	return ic.cmdID
}

func (ic *innerCmd) getServerCmd() (net.Conn, []byte, []byte) {
	return ic.c, ic.hdr, ic.body
}

func (ic *innerCmd) getClientCmd() (uint64, []byte, []byte) {
	return ic.n, ic.hdr, ic.body
}

func (ic *innerCmd) getAdminCmd() (uint64) {
	return ic.n
}

func (ic *innerCmd) getNotifyCmd() (net.Conn, string, uint64, error) {
	return ic.c, ic.str, ic.n, ic.err
}

func newServerCmd(cmdID uint8, c net.Conn, hdr, body []byte) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		c: c,
		hdr: hdr,
		body: body,
	}
}

func newClientCmd(cmdID uint8, sessionID uint64, hdr, body []byte) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		n: sessionID,
		hdr: hdr,
		body: body,
	}
}

func newAdminCmd(cmdID uint8, sessionID uint64) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		n: sessionID,
	}
}

func newNotifyCmd(cmdID uint8, c net.Conn, str string, n uint64, err error) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		c: c,
		str: str,
		n: n,
		err: err,
	}
}
