package gconnfmt

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// packet between gconn and other services
// the format is:
//  hdr ... session ... body

// MakeSessionPkt make a session packet
func MakeSessionPkt(sessions []uint64, cmdID uint16, userData, timestamp uint32, body []byte) ([]byte, error) {
	bodyLen := len(body)
	if bodyLen > LengthOfMaxBody {
		return nil, fmt.Errorf("body length=%d above max body=%d", bodyLen, LengthOfMaxBody)
	}

	bufU16 := make([]byte, 2)
	bufU32 := make([]byte, 4)
	bufU64 := make([]byte, 8)
	var buffer bytes.Buffer
	sessionNum := uint16(len(sessions))
	sessionBodyLen := len(body) + 2 + 8*int(sessionNum)
	if sessionBodyLen > LengthOfMaxBody {
		return nil, fmt.Errorf("session body length=%d above max body=%d", sessionBodyLen, LengthOfMaxBody)
	}

	// fill the header
	binary.BigEndian.PutUint16(bufU16, uint16(sessionBodyLen))
	buffer.Write(bufU16)
	binary.BigEndian.PutUint16(bufU16, cmdID)
	buffer.Write(bufU16)
	binary.BigEndian.PutUint32(bufU32, userData)
	buffer.Write(bufU32)
	binary.BigEndian.PutUint32(bufU32, timestamp)
	buffer.Write(bufU32)

	// fill the session
	binary.BigEndian.PutUint16(bufU16, sessionNum)
	buffer.Write(bufU16)
	for i := 0; i < int(sessionNum); i++ {
		binary.BigEndian.PutUint64(bufU64, sessions[i])
		buffer.Write(bufU64)
	}

	// fill the body
	buffer.Write(body)

	return buffer.Bytes(), nil
}

// MakeOneSessionPkt make a session packet by using sessionID
func MakeOneSessionPkt(sessionID uint64, cmdID uint16, userData, timestamp uint32, body []byte) ([]byte, error) {
	sessions := make([]uint64, 1)
	sessions[0] = sessionID
	return MakeSessionPkt(sessions, cmdID, userData, timestamp, body)
}

// ParseSessionBody parse the session packet body
func ParseSessionBody(body []byte) (uint16, []uint64, []byte) {
	sessionNum := binary.BigEndian.Uint16(body[:2])
	sessions := make([]uint64, sessionNum)
	for i := 0; i < int(sessionNum); i++ {
		b := 2 + i*8
		sessionID := binary.BigEndian.Uint64(body[b:(b + 8)])
		sessions[i] = sessionID
	}
	b := 2 + sessionNum*8
	return sessionNum, sessions, body[b:]
}

// ParseSessionPkt parse the session packet
func ParseSessionPkt(pkt []byte) ([]byte, uint16, []uint64, []byte) {
	hdr := pkt[:LengthOfHeader]
	sessionNum, sessions, body := ParseSessionBody(pkt[LengthOfHeader:])
	return hdr, sessionNum, sessions, body
}

// MakeMasterSet make MasterSet packet sent to gconn
func MakeMasterSet() []byte {
	return MakeHeader(0, CmdMasterSet, 0, 0)
}

// MakeMasterYou make MasterYou packet sent to gconn
func MakeMasterYou() []byte {
	return MakeHeader(0, CmdMasterYou, 0, 0)
}

// MakeMasterNot make MasterNot packet sent to gconn
func MakeMasterNot() []byte {
	return MakeHeader(0, CmdMasterNot, 0, 0)
}

// MakeKickAll make KickAll packet sent to gconn
func MakeKickAll() []byte {
	return MakeHeader(0, CmdKickAll, 0, 0)
}

// MakeBroadcastAll make BroadcastAll packet sent to gconn
func MakeBroadcastAll(pkt []byte) ([]byte, error) {
	pkt2 := MakePkt(CmdBroadcastAll, 0, 0, pkt)
	if len(pkt2) > LengthOfMaxBody {
		return nil, fmt.Errorf("broadcast all length=%d above max body=%d", len(pkt2), LengthOfMaxBody)
	}
	return pkt2, nil
}
