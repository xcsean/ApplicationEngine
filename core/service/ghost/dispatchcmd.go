package main

import (
	"encoding/json"
	"fmt"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

// all dispatchXXX functions run in the main routine context!!!

func dispatchRPC(cmd *innerCmd) bool {
	vmm := getVMMgr()
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdRegisterVM:
		division, version, addr, rspChannel := cmd.getRPCReq()
		uuid, result := vmm.addVM(division, version, addr)
		rspChannel <- newRPCRsp(innerCmdRegisterVM, result, fmt.Sprintf("%d", uuid))
	case innerCmdUnregisterVM:
		division, _, sUUID, rspChannel := cmd.getRPCReq()
		uuid, _ := parseUint64(sUUID)
		result := vmm.delVM(division, uuid)
		rspChannel <- newRPCRsp(innerCmdUnregisterVM, result, "")
	case innerCmdDebug:
		division, cmdOp, cmdParam, rspChannel := cmd.getRPCReq()
		desc, result := vmm.debug(division, cmdOp, cmdParam)
		ec, uc := getSessionMgr().getCount()
		desc = desc + fmt.Sprintf(" session[entity=%d, uuid=%d]", ec, uc)
		rspChannel <- newRPCRsp(innerCmdDebug, result, desc)
	case innerCmdBindSession:
		dispatchSessionBind(cmd)
	case innerCmdUnbindSession:
		dispatchSessionUnbind(cmd)
	}
	return false
}

func dispatchVMM(cmd *innerCmd) bool {
	vmm := getVMMgr()
	cmdID := cmd.getID()
	switch cmdID {
	case innerCmdVMStreamInitFault, innerCmdVMStreamSendFault, innerCmdVMStreamRecvFault:
		division, sVMID, _ := cmd.getVMMCmd()
		vmID, _ := parseUint64(sVMID)
		vmm.delVM(division, vmID)
	case innerCmdVMStreamRecvPkt:
		connSend(cmd.pkt)
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
		c := protocol.PacketType(cmd.pkt.Common.CmdId)
		if c == protocol.Packet_PRIVATE_SESSION_ENTER {
			dispatchSessionEnter(cmd.pkt)
		} else if c == protocol.Packet_PRIVATE_SESSION_LEAVE {
			dispatchSessionLeave(cmd.pkt)
		} else if c == protocol.Packet_PUBLIC_SESSION_VERCHECK {
			dispatchSessionVerCheck(cmd.pkt)
		} else {
			dispatchSessionForward(cmd.pkt)
		}
	}
	return exit
}

func dispatchTMM(cmd *timerCmd) bool {
	sm := getSessionMgr()
	cmdType := cmd.Type
	switch cmdType {
	case timerCmdVMMOnTick:
		getVMMgr().onTick()
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
			rb := protocol.PacketReservedBody{}
			body, _ := json.Marshal(rb)
			pkt := &protocol.SessionPacket{
				Sessions: []uint64{sessionID},
				Common: &protocol.Packet{
					CmdId:     int32(protocol.Packet_PRIVATE_SESSION_KICK),
					UserData:  0,
					Timestamp: 0,
					Body:      string(body[:]),
				},
			}
			connSend(pkt)
		} else {
			log.Debug("discard timer type='%s', userdata1=%d, userdata2=%d", getTimerDesc(cmdType), sessionID, cmd.Userdata2)
		}
	default:
		log.Debug("unknown timer type=%d u1=%d u2=%d", cmdType, cmd.Userdata1, cmd.Userdata2)
	}
	return false
}
