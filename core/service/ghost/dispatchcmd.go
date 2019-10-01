package main

import (
	"fmt"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

// all dispatchXXX functions run in the main routine context!!!

func dispatchRPC(cmd *innerCmd) bool {
	vmm := getVMMgr()
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdRegisterVM:
		division, version, addr, _, rspChannel := cmd.getRPCReq()
		uuid, result := vmm.addVM(division, version, addr)
		rspChannel <- newRPCRsp(innerCmdRegisterVM, result, uuid, "")
	case innerCmdUnregisterVM:
		division, _, _, uuid, rspChannel := cmd.getRPCReq()
		result := vmm.delVM(division, uuid)
		rspChannel <- newRPCRsp(innerCmdUnregisterVM, result, 0, "")
	case innerCmdDebug:
		division, cmdOp, cmdParam, _, rspChannel := cmd.getRPCReq()
		desc, result := vmm.debug(division, cmdOp, cmdParam)
		ec, uc := getSessionMgr().getCount()
		desc = desc + fmt.Sprintf(" session[entity=%d, uuid=%d]", ec, uc)
		rspChannel <- newRPCRsp(innerCmdDebug, result, 0, desc)
	}
	return false
}

func dispatchVMM(cmd *innerCmd) bool {
	vmm := getVMMgr()
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdVMStreamConnFault, innerCmdVMStreamSendFault:
		division, _, _, uuid := cmd.getVMMCmd()
		vmm.delVM(division, uuid)
	}
	return false
}

func dispatchConn(cmd *innerCmd) bool {
	defer dbg.Stacktrace()

	exit := false
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdConnExit:
		finiConnMgr()
		exit = true
	case innerCmdConnSessionUp:
		hdr, body := cmd.getConnCmd()
		header := conn.ParseHeader(hdr)
		if header.CmdID == conn.CmdSessionEnter {
			dispatchSessionEnter(hdr, body)
		} else if header.CmdID == conn.CmdSessionLeave {
			dispatchSessionLeave(hdr, body)
		} else if header.CmdID == conn.CmdVerCheck {
			dispatchSessionVerCheck(hdr, body)
		} else {
			log.Debug("session=%d cmd=%d", header.CmdID)
		}
	}
	return exit
}

func dispatchTMM(cmd *timerCmd) bool {
	vmm := getVMMgr()
	sm := getSessionMgr()
	cmdType := cmd.Type
	switch cmdType {
	case timerCmdVMMOnTick:
		vmm.onTick()
	case timerCmdSessionWaitVerCheck, timerCmdSessionWaitBind:
		sessionID := cmd.Userdata1
		if sm.isSessionState(sessionID, cmdType) {
			// TODO kick the session because of timer expired
			log.Debug("should kick session=%d by timer type=%d", sessionID, cmdType)
		} else {
			log.Debug("discard timer type=%d, userdata1=%d, userdata2=%d", cmdType, sessionID, cmd.Userdata2)
		}
	default:
		log.Debug("type=%d u1=%d u2=%d", cmdType, cmd.Userdata1, cmd.Userdata2)
	}
	return false
}
