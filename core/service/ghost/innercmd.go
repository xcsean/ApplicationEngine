package main

type innerCmd struct {
	cmdID uint8
	s1    string
	s2    string
	s3    string
	rsp   chan *rspRPC
	i1    uint64
	b1    []byte
	b2    []byte
}

type rspRPC struct {
	cmdID  uint8
	result int32
	i1     uint64
	s1     string
	data   []byte
}

const (
	innerCmdRegisterVM        = 101
	innerCmdUnregisterVM      = 102
	innerCmdLoadUserAsset     = 103
	innerCmdSaveUserAsset     = 104
	innerCmdSendPacket        = 105
	innerCmdDebug             = 106
	innerCmdVMStart           = 111
	innerCmdVMStreamConnFault = 112
	innerCmdVMStreamSendFault = 113
	innerCmdVMShouldExit      = 114
	innerCmdConnSessionUp     = 121
	innerCmdConnExit          = 129
)

func (ic *innerCmd) getID() uint8 {
	return ic.cmdID
}

// RPC methods
func (ic *innerCmd) getRPCReq() (string, string, string, uint64, chan *rspRPC) {
	return ic.s1, ic.s2, ic.s3, ic.i1, ic.rsp
}

func (rsp *rspRPC) getRPCRsp() (int32, uint64, string) {
	return rsp.result, rsp.i1, rsp.s1
}

func newRPCReq(cmdID uint8, s1, s2, s3 string, i1 uint64, ch chan *rspRPC) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		s1:    s1,
		s2:    s2,
		s3:    s3,
		i1:    i1,
		rsp:   ch,
	}
}

func newRPCRsp(cmdID uint8, result int32, i1 uint64, s1 string) *rspRPC {
	return &rspRPC{
		cmdID:  cmdID,
		result: result,
		s1:     s1,
		i1:     i1,
	}
}

// VMM methods
func (ic *innerCmd) getVMMCmd() (string, string, string, uint64) {
	return ic.s1, ic.s2, ic.s3, ic.i1
}

func newVMMCmd(cmdID uint8, s1, s2, s3 string, i1 uint64) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		s1:    s1,
		s2:    s2,
		s3:    s3,
		i1:    i1,
	}
}

// Conn methods
func (ic *innerCmd) getConnCmd() ([]byte, []byte) {
	return ic.b1, ic.b2
}

func newConnCmd(cmdID uint8, b1, b2 []byte) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		b1: b1,
		b2: b2,
	}
}
