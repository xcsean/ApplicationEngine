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
	sm := getSessionMgr()
	_, sessionIDs, _ := conn.ParseSessionBody(body)
	sessionID := sessionIDs[0]
	if !sm.isSessionExist(sessionID) {
		return
	}

	log.Debug("session=%d leave", sessionID)
	uuid, bind := sm.getBindUUID(sessionID)
	if bind {
		notifyVMUnbind(sessionID, uuid)
	} else {
		setSessionWaitDelete(sessionID)
	}
}

func dispatchSessionVerCheck(hdr, body []byte) {
	_, sessionIDs, innerBody := conn.ParseSessionBody(body)
	sessionID := sessionIDs[0]

	// check the session exist and its state
	sm := getSessionMgr()
	_, ok := sm.isSessionState(sessionID, timerCmdSessionWaitVerCheck)
	if !ok {
		log.Debug("session=%d shouldn't do ver-check when not in ver-check state, or session not found", sessionID)
		return
	}

	// get the client version & try to pick a vm
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
			notifyClientVerCheck(sessionID, "ver-check ok")
		} else {
			notifyClientVerCheck(sessionID, "no available division can serve you")
			shouldKick = true
		}
	} else {
		notifyClientVerCheck(sessionID, err.Error())
		shouldKick = true
	}

	if shouldKick {
		setSessionWaitDelete(sessionID)
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
