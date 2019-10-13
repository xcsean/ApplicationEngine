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
	exitC  chan struct{}
	wg     sync.WaitGroup
)

func init() {
	assets = make(map[uint64]*assetData)
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
func LockAssetBySession(sessionID uint64, duration int64, asset *protocol.GhostUserasset) (*protocol.GhostUserasset, bool, int64, int32) {
	return nil, true, 0, errno.OK
}

// LockAssetBySystem lock the user asset by system
func LockAssetBySystem(duration int64) (*protocol.GhostUserasset, bool, int64, int32) {
	return nil, true, 0, errno.OK
}
