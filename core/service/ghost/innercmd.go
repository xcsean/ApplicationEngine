package main

import "net"

type innerCmd struct {
	cmdID  uint8
	result int32
	err    error
	c      net.Conn
	s1     string
	s2     string
	rsp    chan *innerCmd
}

const (
	innerCmdRegisterVM    = 101
	innerCmdUnregisterVM  = 102
	innerCmdLoadUserAsset = 103
	innerCmdSaveUserAsset = 104
	innerCmdSendPacket    = 105
)

func (ic *innerCmd) getID() uint8 {
	return ic.cmdID
}

func (ic *innerCmd) getRPCReq() chan<- *innerCmd {
	return ic.rsp
}

func (ic *innerCmd) getRPCRsp() (int32, error) {
	return ic.result, ic.err
}

func newRPCReq(cmdID uint8, s1, s2 string, ch chan *innerCmd) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		s1:    s1,
		s2:    s2,
		rsp:   ch,
	}
}

func newRPCRsp(cmdID uint8, result int32, err error) *innerCmd {
	return &innerCmd{
		cmdID:  cmdID,
		result: result,
		err:    err,
	}
}
