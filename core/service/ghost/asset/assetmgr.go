package asset

import (
	"sync"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
)

const (
	ownerTypeFree = iota
	ownerTypeSystem
	ownerTypeSession
)

type assetOwner struct {
	ownerType   uint8
	ownerID     uint64
	expiredTime int64
}

type assetData struct {
	owner assetOwner
	asset *protocol.GhostUserasset
}

var (
	assets map[uint64]*assetData
	reqC   chan *Req
	exitC  chan struct{}
	wg     sync.WaitGroup
)

func init() {
	assets = make(map[uint64]*assetData)
	reqC = make(chan *Req, 3000)
	exitC = make(chan struct{})
	wg.Add(1)
	go assetLoop()
}

// StopAssetLoop stop the asset loop
func StopAssetLoop() {
	close(exitC)
	wg.Wait()
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
