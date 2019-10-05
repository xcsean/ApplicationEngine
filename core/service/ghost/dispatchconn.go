package main

import (
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/packet"
)

const (
	timeoutWaitVerCheck = time.Duration(30) * time.Second
	timeoutWaitBind     = time.Duration(30) * time.Second
	timeoutWaitUnbind   = time.Duration(30) * time.Second
)

// all dispatchXXX functions run in the main routine context!!!

func dispatchSessionEnter(pkt *protocol.SessionPacket) {
	sessionID := pkt.Sessions[0]
	addr := packet.ParseSessionEnterBody([]byte(pkt.Common.Body))
	log.Debug("session=%d addr='%s' enter", sessionID, addr)

	// create a new session and monitor it
	sm := getSessionMgr()
	sm.addSession(sessionID, addr)
	setSessionVerCheck(sessionID)
}

func dispatchSessionLeave(pkt *protocol.SessionPacket) {
	sessionID := pkt.Sessions[0]
	sm := getSessionMgr()
	if !sm.isSessionExist(sessionID) {
		return
	}

	log.Debug("session=%d leave", sessionID)
	uuid, bind := sm.getBindUUID(sessionID)
	if bind {
		setSessionWaitUnbind(sessionID, uuid)
	} else {
		setSessionWaitDelete(sessionID)
	}
}

func dispatchSessionVerCheck(pkt *protocol.SessionPacket) {
	ver, err := packet.ParseSessionVerCheckBody([]byte(pkt.Common.Body))
	if err != nil {
		log.Debug("parse ver-check body failed: %s", err.Error())
		return
	}
	sessionID := pkt.Sessions[0]

	// check the session exist and its state
	sm := getSessionMgr()
	_, ok := sm.isSessionState(sessionID, timerCmdSessionWaitVerCheck)
	if !ok {
		log.Debug("session=%d shouldn't do ver-check when not in ver-check state, or session not found", sessionID)
		return
	}

	// get the client version & try to pick a vm to serve this client
	shouldKick := false
	log.Debug("session=%d version=%s", sessionID, ver)
	division, result := getVMMgr().pick(ver)
	if result == errno.OK {
		log.Debug("session=%d pick division=%s", sessionID, division)
		sm.setSessionRouting(sessionID, ver, division)
		setSessionWaitBind(sessionID)
		sendClientVerReply(sessionID, result)
	} else {
		shouldKick = true
		sendClientVerReply(sessionID, result)
	}

	if shouldKick {
		setSessionWaitDelete(sessionID)
	}
}

func dispatchSessionBind(cmd *innerCmd) {
	vmm := getVMMgr()
	sm := getSessionMgr()
	division, sSessionID, sUUID, rspChannel := cmd.getRPCReq()
	result := vmm.exist(division)
	if result != errno.OK {
		rspChannel <- newRPCRsp(innerCmdBindSession, result, "")
		return
	}
	uuid, _ := parseUint64(sUUID)
	sessionID, _ := parseUint64(sSessionID)

	// check the uuid bind or not?
	bindSessionID, bind := sm.getBindSession(uuid)
	if bind {
		if bindSessionID == sessionID {
			rspChannel <- newRPCRsp(innerCmdBindSession, errno.OK, "")
		} else {
			// conflict detected, just notify the caller to retry later
			rspChannel <- newRPCRsp(innerCmdBindSession, errno.HOSTVMBINDNEEDRETRY, "")
			_, ok := sm.isSessionState(bindSessionID, timerCmdSessionWorking)
			if ok {
				setSessionWaitUnbind(bindSessionID, uuid)
			}
		}
	} else {
		// check the session bind or not?
		bindUUID, bind := sm.getBindUUID(sessionID)
		if bind {
			if bindUUID == uuid {
				rspChannel <- newRPCRsp(innerCmdBindSession, errno.OK, "")
			} else {
				// conflict detected, just notify the caller don't bind another uuid
				rspChannel <- newRPCRsp(innerCmdBindSession, errno.HOSTVMSESSIONALREADYBIND, "")
			}
		} else {
			_, ok := sm.isSessionState(sessionID, timerCmdSessionWaitBind)
			if ok {
				sm.bindSession(sessionID, uuid)
				sm.setSessionState(sessionID, timerCmdSessionWorking)
				rspChannel <- newRPCRsp(innerCmdBindSession, errno.OK, "")
			} else {
				rspChannel <- newRPCRsp(innerCmdBindSession, errno.HOSTVMSESSIONNOTWAITBIND, "")
			}
		}
	}
}

func dispatchSessionForward(pkt *protocol.SessionPacket) {
	sm := getSessionMgr()
	sessionID := pkt.Sessions[0]

	division, ok := sm.isSessionStateOf(sessionID, []uint8{timerCmdSessionWaitBind, timerCmdSessionWorking})
	if ok {
		result := getVMMgr().forward(division, pkt)
		if result != errno.OK {
			log.Debug("session=%d cmd=%d forward to %s failed: %d", sessionID, pkt.Common.CmdId, division, result)
		}
	} else {
		log.Warn("session=%d discard cmd=%d by state", sessionID, pkt.Common.CmdId)
	}
}

func dispatchSessionUnbind(cmd *innerCmd) {
	vmm := getVMMgr()
	sm := getSessionMgr()
	division, sSessionID, sUUID, rspChannel := cmd.getRPCReq()
	result := vmm.exist(division)
	if result != errno.OK {
		rspChannel <- newRPCRsp(innerCmdUnbindSession, result, "")
		return
	}
	uuid, _ := parseUint64(sUUID)
	sessionID, _ := parseUint64(sSessionID)
	bindSessionID, bind := sm.getBindSession(uuid)
	if bind && bindSessionID == sessionID {
		_, ok := sm.isSessionState(sessionID, timerCmdSessionWaitUnbind)
		if ok {
			sm.unbindSession(sessionID, uuid)
			setSessionWaitDelete(sessionID)
		}
		rspChannel <- newRPCRsp(innerCmdUnbindSession, errno.OK, "")
	} else {
		rspChannel <- newRPCRsp(innerCmdUnbindSession, errno.OK, "")
	}
}
