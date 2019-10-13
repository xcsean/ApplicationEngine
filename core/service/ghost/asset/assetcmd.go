package asset

import "github.com/xcsean/ApplicationEngine/core/protocol"

// Req the asset request
type Req struct {
	Type      uint8
	Userdata1 uint64
	Userdata2 uint64
	Userasset *protocol.GhostUserasset
}

// Rsp the asset response
type Rsp struct {
	Result    int32
	Userdata1 uint64
	Userdata2 uint64
	Userasset *protocol.GhostUserasset
}

const (
	assetCmdLock = iota
	assetCmdUnlock
)
