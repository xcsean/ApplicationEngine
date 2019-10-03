package main

import (
	"encoding/json"
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

const (
	timeoutWaitVerCheck = time.Duration(30) * time.Second
	timeoutWaitBind     = time.Duration(30) * time.Second
	timeoutWaitUnbind   = time.Duration(30) * time.Second
)

// all dispatchXXX functions run in the main routine context!!!

func dispatchSessionEnter(hdr, body []byte) {
	_, sessionIDs, innerBody := conn.ParseSessionBody(body)
	sessionID := sessionIDs[0]

	// get the client address
	var rb conn.ReservedBody
	json.Unmarshal(innerBody, &rb)
	log.Debug("session=%d addr='%s' enter", sessionID, rb.StrParam)

	// create a new session and monitor it
	sm := getSessionMgr()
	sm.addSession(sessionID, rb.StrParam)
	sm.setSessionState(sessionID, timerCmdSessionWaitVerCheck)
	tmmAddDelayTask(timeoutWaitVerCheck, func(c chan *timerCmd) {
		c <- &timerCmd{Type: timerCmdSessionWaitVerCheck, Userdata1: sessionID}
	})
}

func dispatchSessionLeave(hdr, body []byte) {
	_, sessionIDs, _ := conn.ParseSessionBody(body)
	sessionID := sessionIDs[0]

	log.Debug("session=%d leave", sessionID)
	sm := getSessionMgr()
	uuid, bind := sm.getBindUUID(sessionID)
	if bind {
		// TODO send a packet to the division which the uuid is in
		//  and wait the unbind call
		getSessionMgr().setSessionState(sessionID, timerCmdSessionWaitUnbind)
		tmmAddDelayTask(timeoutWaitUnbind, func(c chan *timerCmd) {
			c <- &timerCmd{Type: timerCmdSessionWaitUnbind, Userdata1: sessionID, Userdata2: uuid}
		})
	} else {
		// set the session wait delete state
		getSessionMgr().setSessionState(sessionID, timerCmdSessionWaitDelete)
		timeoutWaitDelete := 3 * time.Second
		tmmAddDelayTask(timeoutWaitDelete, func(c chan *timerCmd) {
			c <- &timerCmd{Type: timerCmdSessionWaitDelete, Userdata1: sessionID, Userdata2: uuid}
		})
	}
}

func dispatchSessionVerCheck(hdr, body []byte) {
	header := conn.ParseHeader(hdr)
	_, sessionIDs, innerBody := conn.ParseSessionBody(body)
	sessionID := sessionIDs[0]

	// check the session exist and its state
	sm := getSessionMgr()
	if !sm.isSessionState(sessionID, timerCmdSessionWaitVerCheck) {
		log.Debug("session=%d shouldn't send ver-check when not in ver-check state, or session not found", sessionID)
		return
	}

	// get the client version & try to pick a division
	shouldKick := false
	var rb conn.ReservedBody
	err := json.Unmarshal(innerBody, &rb)
	if err == nil {
		ver := rb.StrParam
		log.Debug("session=%d version=%s", sessionID, ver)
		// pick a division to serve this client
		vmm := getVMMgr()
		division, result := vmm.pick(ver)
		if result == errno.OK {
			log.Debug("session=%d pick division=%s", sessionID, division)
			sm.setSessionRouting(sessionID, ver, division)
			sm.setSessionState(sessionID, timerCmdSessionWaitBind)
			tmmAddDelayTask(timeoutWaitBind, func(c chan *timerCmd) {
				c <- &timerCmd{Type: timerCmdSessionWaitBind, Userdata1: sessionID}
			})
			// send a ver-check ack to client
			var ack conn.ReservedBody
			ack.StrParam = "ver-check ok"
			body, _ := json.Marshal(ack)
			pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdVerCheck, header.UserData, header.Timestamp, body)
			connSend(pkt)
		} else {
			var ack conn.ReservedBody
			ack.StrParam = "no available division can serve you"
			body, _ := json.Marshal(ack)
			pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdVerCheck, header.UserData, header.Timestamp, body)
			connSend(pkt)
			shouldKick = true
		}
	} else {
		log.Error("session=%d version check failed: %s", sessionID, err.Error())
		var ack conn.ReservedBody
		ack.StrParam = err.Error()
		body, _ := json.Marshal(ack)
		pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdVerCheck, header.UserData, header.Timestamp, body)
		connSend(pkt)
		shouldKick = true
	}

	if shouldKick {
		log.Debug("kick the session=%d by ver-check failed", sessionID)
		pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdSessionKick, header.UserData, header.Timestamp, nil)
		connSend(pkt)
	}
}

func dispatchSessionForward(hdr, body []byte) {
	header := conn.ParseHeader(hdr)
	_, sessionIDs, innerBody := conn.ParseSessionBody(body)
	sessionID := sessionIDs[0]
	sm := getSessionMgr()

	division, ok := sm.isSessionStateOf(sessionID, []uint8{timerCmdSessionWaitBind, timerCmdSessionWorking})
	if ok {
		result := getVMMgr().forward(division, sessionID, header, innerBody)
		if result != errno.OK {
			log.Debug("session=%d cmd=%d forward to %s failed: %d", sessionID, header.CmdID, division, result)
		}
	} else {
		log.Warn("session=%d discard cmd=%d by state", sessionID, header.CmdID)
	}
}