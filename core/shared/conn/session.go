package conn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// MakeSessionID make a session ID
func MakeSessionID(id uint16, seed uint16) uint64 {
	return uint64(time.Now().Unix())<<32 + uint64(seed)<<16 + uint64(id)
}

// packet between conn and other services
// the format is:
//  hdr ... session ... body

// MakeSessionPkt make a session packet
func MakeSessionPkt(sessionIDs []uint64, cmdID uint16, userData, timestamp uint32, body []byte) ([]byte, error) {
	bodyLen := len(body)
	if bodyLen > LengthOfMaxBody {
		return nil, fmt.Errorf("body length=%d above max body=%d", bodyLen, LengthOfMaxBody)
	}

	bufU16 := make([]byte, 2)
	bufU32 := make([]byte, 4)
	bufU64 := make([]byte, 8)
	sessionNum := uint16(len(sessionIDs))
	sessionBodyLen := len(body) + 2 + 8*int(sessionNum)
	if sessionBodyLen > LengthOfMaxBody {
		return nil, fmt.Errorf("session body length=%d above max body=%d", sessionBodyLen, LengthOfMaxBody)
	}

	var buffer bytes.Buffer
	buffer.Grow(LengthOfHeader + int(sessionBodyLen))

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
		binary.BigEndian.PutUint64(bufU64, sessionIDs[i])
		buffer.Write(bufU64)
	}

	// fill the body
	buffer.Write(body)

	return buffer.Bytes(), nil
}

// CopySessionPkt create a new session packet by using header & body
func CopySessionPkt(sessionIDs []uint64, hdr, body []byte) ([]byte, uint16) {
	header := ParseHeader(hdr)
	pkt, err := MakeSessionPkt(sessionIDs, header.CmdID, header.UserData, header.Timestamp, body)
	if err != nil {
		return nil, header.CmdID
	}
	return pkt, header.CmdID
}

// MakeOneSessionPkt make a session packet by using sessionID
func MakeOneSessionPkt(sessionID uint64, cmdID uint16, userData, timestamp uint32, body []byte) ([]byte, error) {
	return MakeSessionPkt([]uint64{sessionID}, cmdID, userData, timestamp, body)
}

// ParseSessionBody parse the session packet body
func ParseSessionBody(body []byte) (uint16, []uint64, []byte) {
	sessionNum := binary.BigEndian.Uint16(body[:2])
	sessionIDs := make([]uint64, sessionNum)
	for i := 0; i < int(sessionNum); i++ {
		b := 2 + i*8
		sessionID := binary.BigEndian.Uint64(body[b:(b + 8)])
		sessionIDs[i] = sessionID
	}
	b := 2 + sessionNum*8
	return sessionNum, sessionIDs, body[b:]
}

// ParseSessionPkt parse the session packet
func ParseSessionPkt(pkt []byte) ([]byte, uint16, []uint64, []byte) {
	hdr := pkt[:LengthOfHeader]
	sessionNum, sessionIDs, body := ParseSessionBody(pkt[LengthOfHeader:])
	return hdr, sessionNum, sessionIDs, body
}

// MakeCommonPkt make a new common packet (no session) by some data & body
func MakeCommonPkt(cmdID uint16, userData, timestamp uint32, body []byte) []byte {
	bufU16 := make([]byte, 2)
	bufU32 := make([]byte, 4)
	bodyLen := uint16(len(body))

	var buffer bytes.Buffer
	buffer.Grow(LengthOfHeader + int(bodyLen))

	// fill the header
	binary.BigEndian.PutUint16(bufU16, bodyLen)
	buffer.Write(bufU16)
	binary.BigEndian.PutUint16(bufU16, cmdID)
	buffer.Write(bufU16)
	binary.BigEndian.PutUint32(bufU32, userData)
	buffer.Write(bufU32)
	binary.BigEndian.PutUint32(bufU32, timestamp)
	buffer.Write(bufU32)

	// fill the body
	buffer.Write(body)

	return buffer.Bytes()
}

// CopyCommonPkt copy a common packet (no session) by header & body
func CopyCommonPkt(hdr, body []byte) ([]byte, uint16) {
	header := ParseHeader(hdr)
	return MakeCommonPkt(header.CmdID, header.UserData, header.Timestamp, body), header.CmdID
}
