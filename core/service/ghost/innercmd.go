package main

import "github.com/xcsean/ApplicationEngine/core/protocol"

const (
	innerCmdRegisterVM        = 101
	innerCmdUnregisterVM      = 102
	innerCmdLoadUserAsset     = 103
	innerCmdSaveUserAsset     = 104
	innerCmdSendPacket        = 105
	innerCmdDebug             = 106
	innerCmdBindSession       = 107
	innerCmdUnbindSession     = 108
	innerCmdVMStart           = 111
	innerCmdVMStreamConnFault = 112
	innerCmdVMStreamSendFault = 113
	innerCmdVMShouldExit      = 114
	innerCmdConnSessionUp     = 121
	innerCmdConnSessionDn     = 122
	innerCmdConnExit          = 129
)

type innerCmd struct {
	cmdID uint8
	s1    string
	s2    string
	s3    string
	rsp   chan *rspRPC
	pkt   *protocol.SessionPacket
}

type rspRPC struct {
	cmdID  uint8
	result int32
	s1     string
}

func (ic *innerCmd) getID() uint8 {
	return ic.cmdID
}

// RPC methods
func (ic *innerCmd) getRPCReq() (string, string, string, chan *rspRPC) {
	return ic.s1, ic.s2, ic.s3, ic.rsp
}

func (rsp *rspRPC) getRPCRsp() (int32, string) {
	return rsp.result, rsp.s1
}

func newRPCReq(cmdID uint8, s1, s2, s3 string, ch chan *rspRPC) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		s1:    s1,
		s2:    s2,
		s3:    s3,
		rsp:   ch,
	}
}

func newRPCRsp(cmdID uint8, result int32, s1 string) *rspRPC {
	return &rspRPC{
		cmdID:  cmdID,
		result: result,
		s1:     s1,
	}
}

// VMM methods
func (ic *innerCmd) getVMMCmd() (string, string, string) {
	return ic.s1, ic.s2, ic.s3
}

func newVMMCmd(cmdID uint8, s1, s2, s3 string) *innerCmd {
	return &innerCmd{
		cmdID: cmdID,
		s1:    s1,
		s2:    s2,
		s3:    s3,
	}
}
