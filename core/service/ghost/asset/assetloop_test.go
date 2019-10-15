package asset

import (
	"testing"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/mysql"
)

func TestAsset(t *testing.T) {
	log.SetupMainLogger("./", "asset.log", "debug")
	pool, err := mysql.New("root", "123456", "192.168.95.182", "3306", "app_ghost_1")
	if err != nil {
		t.Errorf("mysql init failed: %s", err.Error())
		return
	}

	StartAssetLoop(1, pool)

	sessionID := uint64(123456)
	duration := int64(30)
	uuid := uint64(10001)

	time.Sleep(1 * time.Second)

	// test load
	asset := &protocol.GhostUserasset{
		Uuid:  uuid,
		Asset: "",
	}
	asset2, newbee, expiredTime, result := LockAssetBySession(sessionID, duration, asset, false)
	if result != errno.OK {
		t.Errorf("lock result=%d", result)
		return
	}
	t.Logf("lock result=%d newbee=%d expiredTime=%d asset=%v", result, newbee, expiredTime, asset2)

	time.Sleep(1 * time.Second)

	// test save & lock
	asset3 := &protocol.GhostUserasset{
		Uuid:     uuid,
		Revision: asset2.Revision,
		Asset:    "abc",
	}
	_, _, _, result = LockAssetBySession(sessionID, duration, asset3, false)
	if result != errno.OK {
		t.Errorf("save result=%d", result)
		return
	}
	t.Logf("save result=%d", result)

	time.Sleep(1 * time.Second)

	// test renew
	asset.Revision = asset2.Revision
	asset4, _, _, result := LockAssetBySession(sessionID, duration, asset, true)
	if result != errno.OK {
		t.Errorf("load result=%d", result)
		return
	}
	t.Logf("renew result=%d asset=%v", result, asset4)

	result = UnlockAssetBySession(sessionID, asset)
	if result != errno.OK {
		t.Errorf("unlock result=%d", result)
		return
	}
	t.Logf("unlock result=%d", result)

	StopAssetLoop()
}
