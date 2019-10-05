package main

import (
	"fmt"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/packet"
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
	case innerCmdSendPacket:
		s, _, _, _ := cmd.getRPCReq()
		connSend([]byte(s))
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
	case innerCmdVMStreamConnFault, innerCmdVMStreamSendFault:
		division, sVMID, _ := cmd.getVMMCmd()
		vmID, _ := parseUint64(sVMID)
		vmm.delVM(division, vmID)
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
		c := protocol.PacketType(header.CmdID)
		if c == protocol.Packet_PRIVATE_SESSION_ENTER {
			dispatchSessionEnter(hdr, body)
		} else if c == protocol.Packet_PRIVATE_SESSION_LEAVE {
			dispatchSessionLeave(hdr, body)
		} else if c == protocol.Packet_PUBLIC_SESSION_VERCHECK {
			dispatchSessionVerCheck(hdr, body)
		} else {
			dispatchSessionForward(hdr, body)
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
			pkt, _ := packet.MakeSessionKick([]uint64{sessionID})
			connSend(pkt)
		} else {
			log.Debug("discard timer type='%s', userdata1=%d, userdata2=%d", getTimerDesc(cmdType), sessionID, cmd.Userdata2)
		}
	default:
		log.Debug("unknown timer type=%d u1=%d u2=%d", cmdType, cmd.Userdata1, cmd.Userdata2)
	}
	return false
}
