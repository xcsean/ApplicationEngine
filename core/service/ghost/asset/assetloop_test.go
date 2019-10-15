package asset

import (
	"testing"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/mysql"
)

func TestAssetLoop(t *testing.T) {
	log.SetupMainLogger("./", "asset.log", "debug")
	pool, err := mysql.New("root", "123456", "192.168.95.182", "3306", "app_ghost_1")
	if err != nil {
		t.Errorf("mysql init failed: %s", err.Error())
		return
	}

	StartAssetLoop(pool)

	sessionID := uint64(123456)
	duration := int64(30)
	uuid := uint64(10001)
	asset := &protocol.GhostUserasset{
		Uuid:     uuid,
		Revision: uint64(0),
		Asset:    "",
	}
	asset2, newbee, expiredTime, result := LockAssetBySession(sessionID, duration, asset)
	if result != errno.OK {
		t.Errorf("lock result=%d", result)
		return
	}
	t.Logf("lock result=%d newbee=%d expiredTime=%d asset=%v", result, newbee, expiredTime, asset2)

	result = UnlockAssetBySession(sessionID, nil)
	if result != errno.OK {
		t.Errorf("unlock result=%d", result)
		return
	}
	t.Logf("unlock result=%d", result)

	StopAssetLoop()
}
