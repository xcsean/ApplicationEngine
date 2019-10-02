package main

import "github.com/xcsean/ApplicationEngine/core/shared/log"

const (
	uuidDefaultValue uint64 = 0
)

type sessionEntity struct {
	state    uint8
	ver      []byte
	uuid     uint64 // bind or not
	division string
	addr     string
}

type sessionMgr struct {
	s2e map[uint64]*sessionEntity
	u2s map[uint64]uint64 // fast-index, uuid -> session, bind or not
}

var (
	sessMgr *sessionMgr
)

func init() {
	sessMgr = &sessionMgr{
		s2e: make(map[uint64]*sessionEntity),
		u2s: make(map[uint64]uint64),
	}
}

func getSessionMgr() *sessionMgr {
	return sessMgr
}

func (sm *sessionMgr) getCount() (int, int) {
	return len(sm.s2e), len(sm.u2s)
}

func (sm *sessionMgr) isSessionExist(sessionID uint64) bool {
	_, ok := sm.s2e[sessionID]
	return ok
}

func (sm *sessionMgr) isUserExist(uuid uint64) bool {
	_, ok := sm.u2s[uuid]
	return ok
}

func (sm *sessionMgr) addSession(sessionID uint64, addr string) {
	e := &sessionEntity{
		state:    0,
		ver:      []byte{0, 0, 0, 0},
		uuid:     uuidDefaultValue,
		division: "",
		addr:     addr,
	}
	sm.s2e[sessionID] = e
}

func (sm *sessionMgr) delSession(sessionID uint64) {
	delete(sm.s2e, sessionID)
}

func (sm *sessionMgr) isSessionState(sessionID uint64, state uint8) bool {
	e, ok := sm.s2e[sessionID]
	if !ok {
		return false
	}
	return e.state == state
}

func (sm *sessionMgr) isSessionStateOf(sessionID uint64, states []uint8) (string, bool) {
	e, ok := sm.s2e[sessionID]
	if !ok {
		return "", false
	}

	for i := 0; i < len(states); i++ {
		if e.state == states[i] {
			return e.division, true
		}
	}
	return "", false
}

func (sm *sessionMgr) setSessionState(sessionID uint64, state uint8) {
	e, ok := sm.s2e[sessionID]
	if ok && e.state != state {
		log.Debug("session=%d change state from %d to %d", sessionID, e.state, state)
		e.state = state
	}
}

func (sm *sessionMgr) setSessionRouting(sessionID uint64, ver, division string) {
	e, ok := sm.s2e[sessionID]
	if !ok {
		return
	}

	e.ver = []byte{1, 1, 1, 1}
	e.division = division
}

func (sm *sessionMgr) bindSession(sessionID, uuid uint64) bool {
	e, ok := sm.s2e[sessionID]
	if !ok {
		return false
	}
	if e.uuid != uuidDefaultValue {
		log.Warn("session=%d uuid=%d already binded, new uuid=%d", sessionID, e.uuid, uuid)
		return false
	}

	// bind the session with uuid
	e.uuid = uuid
	sm.u2s[uuid] = sessionID
	return true
}

func (sm *sessionMgr) unbindSession(sessionID, uuid uint64) {
	s, ok := sm.u2s[uuid]
	if ok && s == sessionID {
		delete(sm.u2s, uuid)
	}
	e, ok := sm.s2e[sessionID]
	if ok && e.uuid != uuidDefaultValue {
		e.uuid = uuidDefaultValue
	}
}

func (sm *sessionMgr) getBindSession(uuid uint64) (uint64, bool) {
	s, ok := sm.u2s[uuid]
	if !ok {
		return 0, false
	}
	return s, true
}

func (sm *sessionMgr) getBindUUID(sessionID uint64) (uint64, bool) {
	e, ok := sm.s2e[sessionID]
	if !ok {
		return 0, false
	}
	if e.uuid == uuidDefaultValue {
		return 0, false
	}
	return e.uuid, true
}
