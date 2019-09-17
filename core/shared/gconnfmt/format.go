package gconnfmt

// Header header format
type Header struct {
	BodyLen   uint16
	CmdID     uint16
	UserData  uint32
	Timestamp uint32
}

// const vars in header format
const (
	LengthOfBodyLen   = 2
	LengthOfCmdID     = 2
	LengthOfUserData  = 4
	LengthOfTimestamp = 4
)

// const vars in header format
const (
	BodyLenStart   = 0
	BodyLenEnd     = 2
	CmdIDStart     = 2
	CmdIDEnd       = 4
	UserDataStart  = 4
	UserDataEnd    = 8
	TimestampStart = 8
	TimestampEnd   = 12
)

// const vars in header data
const (
	LengthOfHeader    = LengthOfBodyLen + LengthOfCmdID + LengthOfUserData + LengthOfTimestamp
	LengthOfMaxBody   = 60000
	LengthOfMaxPacket = LengthOfHeader + LengthOfMaxBody
)

// private command
const (
	CmdPrivateBegin = 60000
	CmdSessionEnter = 60001
	CmdSessionLeave = 60002
	CmdSessionRoute = 60003
	CmdSessionKick  = 60004
	CmdMasterSet    = 60011
	CmdMasterYou    = 60012
	CmdMasterNot    = 60013
	CmdBroadcastAll = 60021
	CmdKickAll      = 60031
	CmdPrivateEnd   = 65535
)

// PrivateBody private command body format
type PrivateBody struct {
	StrParam string
}
