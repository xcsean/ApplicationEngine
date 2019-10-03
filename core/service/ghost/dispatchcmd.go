package main

import (
	"fmt"
	"strconv"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

// all dispatchXXX functions run in the main routine context!!!

func dispatchRPC(cmd *innerCmd) bool {
	vmm := getVMMgr()
	sm := getSessionMgr()
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
	case innerCmdBindSession:
		division, sSessionID, sUUID, _, rspChannel := cmd.getRPCReq()
		result := vmm.exist(division)
		if result != errno.OK {
			rspChannel <- newRPCRsp(innerCmdBindSession, result, 0, "")
			return false
		}
		uuid, _ := strconv.ParseInt(sUUID, 10, 64)
		sessionID, _ := strconv.ParseInt(sSessionID, 10, 64)
		// check the uuid bind or not?
		bindSessionID, bind := sm.getBindSession(uint64(uuid))
		if bind {
			if bindSessionID == uint64(sessionID) {
				rspChannel <- newRPCRsp(innerCmdBindSession, errno.OK, 0, "")
			} else {
				// notify caller to retry later
				// TODO add the caller into bind pending list of session manager
				rspChannel <- newRPCRsp(innerCmdBindSession, errno.HOSTVMBINDNEEDRETRY, 0, "")
				// notify the vm to unbind the session with the uuid
				notifyVMUnbind(bindSessionID, uint64(uuid))
			}
		} else {
			// check the session bind or not?
			bindUUID, bind := sm.getBindUUID(uint64(sessionID))
			if bind {
				if bindUUID == uint64(uuid) {
					rspChannel <- newRPCRsp(innerCmdBindSession, errno.OK, 0, "")
				} else {
					rspChannel <- newRPCRsp(innerCmdBindSession, errno.HOSTVMSESSIONALREADYBIND, 0, "")
				}
			} else {
				sm.bindSession(uint64(sessionID), uint64(uuid))
				sm.setSessionState(uint64(sessionID), timerCmdSessionWorking)
				rspChannel <- newRPCRsp(innerCmdBindSession, errno.OK, 0, "")
			}
		}
	case innerCmdUnbindSession:
		division, sSessioinID, sUUID, _, rspChannel := cmd.getRPCReq()
		result := vmm.exist(division)
		if result != errno.OK {
			rspChannel <- newRPCRsp(innerCmdUnbindSession, result, 0, "")
			return false
		}
		uuid, _ := strconv.ParseInt(sUUID, 10, 64)
		sessionID, _ := strconv.ParseInt(sSessioinID, 10, 64)
		bindSessionID, bind := sm.getBindSession(uint64(uuid))
		if bind && bindSessionID == uint64(sessionID) {
			_, ok := sm.isSessionState(uint64(sessionID), timerCmdSessionWaitUnbind)
			if ok {
				sm.unbindSession(uint64(sessionID), uint64(uuid))
				setSessionWaitDelete(uint64(sessionID))
			}
			rspChannel <- newRPCRsp(innerCmdUnbindSession, errno.OK, 0, "")
		} else {
			rspChannel <- newRPCRsp(innerCmdUnbindSession, errno.OK, 0, "")
		}
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
		_, ok := sm.isSessionState(sessionID, cmdType)
		if ok {
			setSessionWaitDelete(sessionID)
		} else {
			log.Debug("discard timer type='%s', userdata1=%d, userdata2=%d", getTimerDesc(cmdType), sessionID, cmd.Userdata2)
		}
	case timerCmdSessionWaitUnbind:
		sessionID := cmd.Userdata1
		uuid := cmd.Userdata2
		_, ok := sm.isSessionState(sessionID, cmdType)
		if ok {
			log.Debug("unbind session=%d uuid=%d by timer type='%s'", sessionID, uuid, getTimerDesc(cmdType))
			sm.unbindSession(sessionID, uuid)
			setSessionWaitDelete(sessionID)
		} else {
			log.Debug("discard timer type='%s', userdata1=%d, userdata2=%d", getTimerDesc(cmdType), sessionID, cmd.Userdata2)
		}
	case timerCmdSessionWaitDelete:
		sessionID := cmd.Userdata1
		_, ok := sm.isSessionState(sessionID, cmdType)
		if ok {
			log.Debug("session=%d delete now", sessionID)
			sm.setSessionState(sessionID, timerCmdSessionDeleted)
			sm.delSession(sessionID)
			// kick the session
			pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdSessionKick, 0, 0, nil)
			connSend(pkt)
		}
	default:
		log.Debug("type=%d u1=%d u2=%d", cmdType, cmd.Userdata1, cmd.Userdata2)
	}
	return false
}
