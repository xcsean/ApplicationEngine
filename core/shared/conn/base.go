package conn

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

// reserved command
const (
	CmdPrivateBegin = 60000
	CmdSessionEnter = 60001 // conn --> backend
	CmdSessionLeave = 60002 // conn --> backend
	CmdSessionRoute = 60003 // conn <-- backend
	CmdSessionKick  = 60004 // conn <-- backend
	CmdSessionPing  = 60005 // conn <-- backend
	CmdSessionPong  = 60006 // conn --> backend
	CmdMasterSet    = 60011 // conn <-- backend
	CmdMasterYou    = 60012 // conn --> backend
	CmdMasterNot    = 60013 // conn --> backend
	CmdBroadcastAll = 60021 // conn <-- backend
	CmdKickAll      = 60031 // conn <-- backend
	CmdNotifyClient = 65001 // client <-- conn
	CmdPrivateEnd   = 65500
	CmdPing         = 65501 // client <-> backend
	CmdPong         = 65502 // client <-> backend
	CmdVersionCheck = 65535 // client <-> backend
)

// ReservedBody reserved command body format
type ReservedBody struct {
	StrParam string
}
