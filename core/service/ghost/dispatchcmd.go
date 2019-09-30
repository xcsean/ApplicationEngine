package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

// all dispatchXXX functions run in the main routine context!!!

func dispatchRPC(vmm *vmMgr, cmd *innerCmd) bool {
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

func dispatchVMM(vmm *vmMgr, cmd *innerCmd) bool {
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdVMStreamConnFault, innerCmdVMStreamSendFault:
		division, _, _, uuid := cmd.getVMMCmd()
		vmm.delVM(division, uuid)
	}

	return false
}

func dispatchConn(vmm *vmMgr, cmd *innerCmd) bool {
	defer dbg.Stacktrace()

	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdConnSessionUp:
		hdr, body := cmd.getConnCmd()
		header := conn.ParseHeader(hdr)
		_, sessionIDs, innerBody := conn.ParseSessionBody(body)
		sessionID := sessionIDs[0]
		if header.CmdID == conn.CmdSessionEnter {
			// get the client address
			var rb conn.ReservedBody
			json.Unmarshal(innerBody, &rb)
			log.Debug("session=%d addr='%s' enter", sessionID, rb.StrParam)
			// create a new session and monitor it
			getSessionMgr().addSession(sessionID, rb.StrParam, timerCmdSessionWaitVerCheck)
			tmmAddDelayTask(10*time.Second, func(c chan *timerCmd) {
				c <- &timerCmd{Type: timerCmdSessionWaitVerCheck, Userdata1: sessionID}
			})
		} else if header.CmdID == conn.CmdSessionLeave {
			log.Debug("session=%d leave", sessionID)
			sm := getSessionMgr()
			uuid, bind := sm.getBindUUID(sessionID)
			if bind {
				sm.unbindSession(sessionID, uuid)
				// TODO send a packet to the division which the uuid is in
			}
			sm.delSession(sessionID)
		} else if header.CmdID == conn.CmdVerCheck {
			// get the client version
			var rb conn.ReservedBody
			err := json.Unmarshal(innerBody, &rb)
			if err == nil {
				ver := rb.StrParam
				log.Debug("session=%d version=%s", sessionID, ver)
				division, result := vmm.pick(ver)
				if result == errno.OK {
					log.Debug("session=%d pick division=%s", sessionID, division)
					getSessionMgr().setSessionRouting(sessionID, ver, division, timerCmdSessionWaitBindUser)
					tmmAddDelayTask(10*time.Second, func(c chan *timerCmd) {
						c <- &timerCmd{Type: timerCmdSessionWaitBindUser, Userdata1: sessionID}
					})
				} else {
					// TODO kick the session after send a pkt
				}
			} else {
				log.Error("session=%d version check failed: %s", sessionID, err.Error())
				// TODO kick the session after send a pkt
			}
		} else {
			log.Debug("session=%d cmd=%d", header.CmdID)
		}
	}

	return false
}

func dispatchTMM(vmm *vmMgr, cmd *timerCmd) bool {
	cmdType := cmd.Type
	switch cmdType {
	case timerCmdVMMOnTick:
		vmm.onTick()
	case timerCmdSessionWaitVerCheck, timerCmdSessionWaitBindUser:
		sm := getSessionMgr()
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
