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
	CmdPrivateBegin   = 60000
	CmdSessionEnter   = 60001 // conn --> host
	CmdSessionLeave   = 60002 // conn --> host
	CmdSessionRoute   = 60003 // conn <-- host
	CmdSessionKick    = 60004 // conn <-- host
	CmdSessionPing    = 60005 // conn <-- host
	CmdSessionPong    = 60006 // conn --> host
	CmdMasterSet      = 60011 // conn <-- host
	CmdMasterYou      = 60012 // conn --> host
	CmdMasterNot      = 60013 // conn --> host
	CmdBroadcastAll   = 60021 // conn <-- host
	CmdKickAll        = 60031 // conn <-- host
	CmdNotifyClient   = 65001 // client <-- conn
	CmdNotifyVMBind   = 65011 // host --> vm
	CmdNotifyVMUnbind = 65012 // host --> vm
	CmdPrivateEnd     = 65500
	CmdPing           = 65501 // client <-> host
	CmdPong           = 65502 // client <-> host
	CmdVerCheck       = 65535 // client <-> host
)

// ReservedBody reserved command body format
type ReservedBody struct {
	StrParam string
	Kv       map[string]string
}
