package asset

import "github.com/xcsean/ApplicationEngine/core/protocol"

const (
	assetCmdLock = iota
	assetCmdUnlock
	assetCmdSystemLock
	assetCmdSystemUnlock
)

const (
	ownerTypeFree = iota
	ownerTypeSystem
	ownerTypeSession
)

// Req the asset request
type Req struct {
	Type       uint8
	Userdata1  uint64
	Userdata2  uint64
	Userasset  *protocol.GhostUserasset
	RspChannel chan *Rsp
}

// Rsp the asset response
type Rsp struct {
	Result    int32
	Userdata1 uint64
	Userdata2 uint64
	Userasset *protocol.GhostUserasset
}
