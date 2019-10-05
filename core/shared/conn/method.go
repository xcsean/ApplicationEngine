package conn

import (
	"encoding/binary"
)

// MakeHeader make the header
func MakeHeader(bodyLen, cmdID uint16, userData, timestamp uint32) []byte {
	hdr := make([]byte, LengthOfHeader)
	bodyLenBuf := hdr[BodyLenStart:BodyLenEnd]
	binary.BigEndian.PutUint16(bodyLenBuf, bodyLen)
	cmdIDBuf := hdr[CmdIDStart:CmdIDEnd]
	binary.BigEndian.PutUint16(cmdIDBuf, cmdID)
	userDataBuf := hdr[UserDataStart:UserDataEnd]
	binary.BigEndian.PutUint32(userDataBuf, userData)
	timestampBuf := hdr[TimestampStart:TimestampEnd]
	binary.BigEndian.PutUint32(timestampBuf, timestamp)
	return hdr
}

// ParseHeader parse the header
func ParseHeader(hdr []byte) *Header {
	bodyLen := binary.BigEndian.Uint16(hdr[BodyLenStart:BodyLenEnd])
	cmdID := binary.BigEndian.Uint16(hdr[CmdIDStart:CmdIDEnd])
	userData := binary.BigEndian.Uint32(hdr[UserDataStart:UserDataEnd])
	timestamp := binary.BigEndian.Uint32(hdr[TimestampStart:TimestampEnd])
	return &Header{
		BodyLen:   bodyLen,
		CmdID:     cmdID,
		UserData:  userData,
		Timestamp: timestamp,
	}
}
