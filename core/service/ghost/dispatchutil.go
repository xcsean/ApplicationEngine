package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func parseUint64(s string) (uint64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	return uint64(i), err
}
func makeVMUnbind(sessionID, uuid uint64) *protocol.SessionPacket {
	rb := protocol.PacketReservedBody{Kv: map[string]string{"uuid": fmt.Sprintf("%d", uuid)}}
	body, _ := json.Marshal(rb)
	pkt := &protocol.SessionPacket{
		Sessions: []uint64{sessionID},
		Common: &protocol.Packet{
			CmdId:     int32(protocol.Packet_PRIVATE_NOTIFY_VM_UNBIND),
			UserData:  0,
			Timestamp: 0,
			Body:      string(body[:]),
		},
	}
	return pkt
}

func sendClientVerReply(sessionID uint64, result int32) {
	rb := protocol.PacketReservedBody{Kv: map[string]string{"result": fmt.Sprintf("%d", result)}}
	body, _ := json.Marshal(rb)
	pkt := &protocol.SessionPacket{
		Sessions: []uint64{sessionID},
		Common: &protocol.Packet{
			CmdId:     int32(protocol.Packet_PUBLIC_SESSION_VERREPLY),
			UserData:  0,
			Timestamp: 0,
			Body:      string(body[:]),
		},
	}
	connSend(pkt)
}

func setSessionVerCheck(sessionID uint64) {
	sm := getSessionMgr()
	sm.setSessionState(sessionID, timerCmdSessionWaitVerCheck)
	tmmAddDelayTask(timeoutWaitVerCheck, func(c chan *timerCmd) {
		c <- &timerCmd{Type: timerCmdSessionWaitVerCheck, Userdata1: sessionID}
	})
}

func setSessionWaitBind(sessionID uint64) {
	sm := getSessionMgr()
	sm.setSessionState(sessionID, timerCmdSessionWaitBind)
	tmmAddDelayTask(timeoutWaitBind, func(c chan *timerCmd) {
		c <- &timerCmd{Type: timerCmdSessionWaitBind, Userdata1: sessionID}
	})
}

func setSessionWaitUnbind(sessionID, uuid uint64) {
	sm := getSessionMgr()
	division, ok := sm.isSessionState(sessionID, timerCmdSessionWorking)
	if ok {
		// notify the vm to unbind
		pkt := makeVMUnbind(sessionID, uuid)
		result := getVMMgr().forward(division, pkt)
		if result != errno.OK {
			log.Debug("session=%d cmd=%d forward to %s failed: %d", sessionID, pkt.Common.CmdId, division, result)
		}
		// set state to WaitUnbind
		if sm.setSessionState(sessionID, timerCmdSessionWaitUnbind) {
			tmmAddDelayTask(timeoutWaitUnbind, func(c chan *timerCmd) {
				c <- &timerCmd{Type: timerCmdSessionWaitUnbind, Userdata1: sessionID, Userdata2: uint64(uuid)}
			})
		}
	}
}

func setSessionWaitDelete(sessionID uint64) {
	sm := getSessionMgr()
	if sm.setSessionState(sessionID, timerCmdSessionWaitDelete) {
		timeoutWaitDelete := 3 * time.Second
		tmmAddDelayTask(timeoutWaitDelete, func(c chan *timerCmd) {
			c <- &timerCmd{Type: timerCmdSessionWaitDelete, Userdata1: uint64(sessionID)}
		})
	}
}
