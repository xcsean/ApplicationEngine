package conn

import (
	"bytes"
	"encoding/binary"
)

// IsPublicCmd tell whether it's a public command or not
func IsPublicCmd(cmdID uint16) bool {
	return cmdID < CmdPrivateBegin
}

// IsPrivateCmd tell whether it's a private command or not
func IsPrivateCmd(cmdID uint16) bool {
	return cmdID >= CmdPrivateBegin
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

// MakePkt make packet sent to client
func MakePkt(cmdID uint16, userData, timestamp uint32, body []byte) []byte {
	var buffer bytes.Buffer
	bufU16 := make([]byte, 2)
	bufU32 := make([]byte, 4)
	bodyLen := uint16(len(body))

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
