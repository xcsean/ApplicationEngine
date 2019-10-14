package asset

import (
	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/mysql"
)

// StartAssetLoop start the asset loop
func StartAssetLoop(pool *mysql.DB) {
	if !flag {
		flag = true
		start(pool)
	}
}

// StopAssetLoop stop the asset loop
func StopAssetLoop() {
	stop()
}

// LockAssetBySession lock the user asset by session id
func LockAssetBySession(sessionID uint64, duration int64, asset *protocol.GhostUserasset) (*protocol.GhostUserasset, int64, int64, int32) {
	rspChannel := make(chan *Rsp, 1)
	reqC <- &Req{
		Type:       assetCmdLock,
		Userdata1:  sessionID,
		Userdata2:  uint64(duration),
		Userasset:  asset,
		RspChannel: rspChannel,
	}
	rsp := <-rspChannel
	return rsp.Userasset, int64(rsp.Userdata1), int64(rsp.Userdata2), rsp.Result
}

// LockAssetBySystem lock the user asset by system
func LockAssetBySystem(duration int64) (*protocol.GhostUserasset, bool, int64, int32) {
	return nil, true, 0, errno.OK
}

// UnlockAssetBySession unlock the user asset by session id
func UnlockAssetBySession(sessionID uint64, asset *protocol.GhostUserasset) int32 {
	rspChannel := make(chan *Rsp, 1)
	reqC <- &Req{
		Type:       assetCmdUnlock,
		Userdata1:  sessionID,
		Userasset:  asset,
		RspChannel: rspChannel,
	}
	rsp := <-rspChannel
	return rsp.Result
}

// UnlockAssetBySystem unlock the user asset by system
func UnlockAssetBySystem(asset *protocol.GhostUserasset) int32 {
	return errno.OK
}
