package main

type innerCmd struct {
	cmdID uint8
	s1    string
	s2    string
	rsp   chan *rspRPC
}

type rspRPC struct {
	cmdID  uint8
	result int32
	data   []byte
}

const (
	innerCmdRegisterVM    = 101
	innerCmdUnregisterVM  = 102
	innerCmdLoadUserAsset = 103
	innerCmdSaveUserAsset = 104
	innerCmdSendPacket    = 105
	innerCmdDebug         = 106
)

func (ic *innerCmd) getID() uint8 {
	return ic.cmdID
}

func (ic *innerCmd) getRPCReq() (string, string, chan *rspRPC) {
	return ic.s1, ic.s2, ic.rsp
}

func (rsp *rspRPC) getRPCRsp() int32 {
	return rsp.result
}

func newRPCReq(cmdID uint8, s1, s2 string, ch chan *rspRPC) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		s1:    s1,
		s2:    s2,
		rsp:   ch,
	}
}

func newRPCRsp(cmdID uint8, result int32) *rspRPC {
	return &rspRPC{
		cmdID:  cmdID,
		result: result,
	}
}
