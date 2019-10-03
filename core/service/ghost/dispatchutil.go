package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func makeVMUnbind(uuid uint64) (*conn.Header, []byte) {
	rb := &conn.ReservedBody{
		Kv: make(map[string]string),
	}
	rb.Kv["uuid"] = fmt.Sprintf("%d", uuid)
	body, _ := json.Marshal(rb)
	header := &conn.Header{
		BodyLen:   uint16(len(body)),
		CmdID:     conn.CmdNotifyVMUnbind,
		UserData:  0,
		Timestamp: 0,
	}
	return header, body
}

func notifyVMUnbind(sessionID, uuid uint64) {
	sm := getSessionMgr()
	division, ok := sm.isSessionState(sessionID, timerCmdSessionWorking)
	if ok {
		header, body := makeVMUnbind(uint64(uuid))
		result := getVMMgr().forward(division, sessionID, header, body)
		if result != errno.OK {
			log.Debug("session=%d cmd=%d forward to %s failed: %d", sessionID, header.CmdID, division, result)
		}
		sm.setSessionState(sessionID, timerCmdSessionWaitUnbind)
		tmmAddDelayTask(timeoutWaitUnbind, func(c chan *timerCmd) {
			c <- &timerCmd{Type: timerCmdSessionWaitUnbind, Userdata1: sessionID, Userdata2: uint64(uuid)}
		})
	}
}

func notifyClientVerCheck(sessionID uint64, s string) {
	var ack conn.ReservedBody
	ack.StrParam = s
	body, _ := json.Marshal(ack)
	pkt, _ := conn.MakeOneSessionPkt(sessionID, conn.CmdVerCheck, 0, 0, body)
	connSend(pkt)
}

func setSessionWaitDelete(sessionID uint64) {
	sm := getSessionMgr()
	sm.setSessionState(uint64(sessionID), timerCmdSessionWaitDelete)
	timeoutWaitDelete := 3 * time.Second
	tmmAddDelayTask(timeoutWaitDelete, func(c chan *timerCmd) {
		c <- &timerCmd{Type: timerCmdSessionWaitDelete, Userdata1: uint64(sessionID)}
	})
}
