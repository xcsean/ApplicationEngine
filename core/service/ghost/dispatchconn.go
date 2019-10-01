package main

import (
	"encoding/json"
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
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
	getSessionMgr().addSession(sessionID, rb.StrParam, timerCmdSessionWaitVerCheck)
	tmmAddDelayTask(10*time.Second, func(c chan *timerCmd) {
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
		sm.unbindSession(sessionID, uuid)
		// TODO send a packet to the division which the uuid is in
	}
	sm.delSession(sessionID)
}

func dispatchSessionVerCheck(hdr, body []byte) {
	vmm := getVMMgr()

	_, sessionIDs, innerBody := conn.ParseSessionBody(body)
	sessionID := sessionIDs[0]

	// TODO check the session exist and its state

	// get the client version
	var rb conn.ReservedBody
	err := json.Unmarshal(innerBody, &rb)
	if err == nil {
		ver := rb.StrParam
		log.Debug("session=%d version=%s", sessionID, ver)
		// pick a division to serve this client
		division, result := vmm.pick(ver)
		if result == errno.OK {
			log.Debug("session=%d pick division=%s", sessionID, division)
			getSessionMgr().setSessionRouting(sessionID, ver, division, timerCmdSessionWaitBind)
			tmmAddDelayTask(10*time.Second, func(c chan *timerCmd) {
				c <- &timerCmd{Type: timerCmdSessionWaitBind, Userdata1: sessionID}
			})
			// send a ver-check ack to client
			var ack conn.ReservedBody
			ack.StrParam = "ver-check ok"
			body, _ := json.Marshal(ack)
			pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdVerCheck, 0, 0, body)
			connSend(pkt)
		} else {
			// TODO kick the session after send a pkt
		}
	} else {
		log.Error("session=%d version check failed: %s", sessionID, err.Error())
		// TODO kick the session after send a pkt
	}
}
