package main

import (
	"fmt"
	"time"

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

func dispatchConn(cmd *connCmd) bool {
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
			dispatchSessionForward(hdr, body)
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
			log.Debug("should kick session=%d by timer type=%d", sessionID, cmdType)
			pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdSessionKick, 0, 0, nil)
			connSend(pkt)
		} else {
			log.Debug("discard timer type=%d, userdata1=%d, userdata2=%d", cmdType, sessionID, cmd.Userdata2)
		}
	case timerCmdSessionWaitUnbind:
		sessionID := cmd.Userdata1
		uuid := cmd.Userdata2
		if sm.isSessionState(sessionID, cmdType) {
			log.Debug("unbind session=%d uuid=%d by timer type=%d", sessionID, uuid, cmdType)
			sm.unbindSession(sessionID, uuid)
			// set the session wait delete state
			getSessionMgr().setSessionState(sessionID, timerCmdSessionWaitDelete)
			timeoutWaitDelete := 3 * time.Second
			tmmAddDelayTask(timeoutWaitDelete, func(c chan *timerCmd) {
				c <- &timerCmd{Type: timerCmdSessionWaitDelete, Userdata1: sessionID, Userdata2: uuid}
			})
		} else {
			log.Debug("discard timer type=%d, userdata1=%d, userdata2=%d", cmdType, sessionID, cmd.Userdata2)
		}
	case timerCmdSessionWaitDelete:
		sessionID := cmd.Userdata1
		if sm.isSessionState(sessionID, cmdType) {
			log.Debug("session=%d delete now", sessionID)
			sm.setSessionState(sessionID, timerCmdSessionDeleted)
			sm.delSession(sessionID)
		}
	default:
		log.Debug("type=%d u1=%d u2=%d", cmdType, cmd.Userdata1, cmd.Userdata2)
	}
	return false
}
