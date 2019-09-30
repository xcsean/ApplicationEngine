package main

import (
	"encoding/json"
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
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
		if header.CmdID == conn.CmdSessionEnter {
			_, sessionIDs, innerBody := conn.ParseSessionBody(body)
			// get the client address
			var rb conn.ReservedBody
			json.Unmarshal(innerBody, &rb)
			log.Debug("session=%d addr='%s' enter", sessionIDs[0], rb.StrParam)
		} else if header.CmdID == conn.CmdSessionLeave {
			_, sessionIDs, _ := conn.ParseSessionBody(body)
			log.Debug("session=%d leave", sessionIDs[0])
		} else if header.CmdID == conn.CmdVersionCheck {
			_, sessionIDs, innerBody := conn.ParseSessionBody(body)
			var rb conn.ReservedBody
			err := json.Unmarshal(innerBody, &rb)
			if err == nil {
				log.Debug("session=%d version=%s", sessionIDs[0], rb.StrParam)
				tmmAddDelayTask(3*time.Second, func(c chan *timerCmd) {
					c <- &timerCmd{Type: timerCmdSessionWaitBindUser, Userdata1: sessionIDs[0], }
				})
			} else {
				log.Error("session=%d version check failed: %s", sessionIDs[0], err.Error())
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
	case timerCmdSessionWaitBindUser:
		log.Debug("wait bind user expired: userdata1=%d, userdata2=%d", cmd.Userdata1, cmd.Userdata2)
	default:
		log.Debug("type=%d u1=%d u2=%d", cmdType, cmd.Userdata1, cmd.Userdata2)
	}

	return false
}