package asset

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/mysql"
)

// StartAssetLoop start the asset loop
func StartAssetLoop(id int64, pool *mysql.DB) {
	start(id, pool)
}

// StopAssetLoop stop the asset loop
func StopAssetLoop() {
	stop()
}

// LockAssetBySession lock the user asset by session id
func LockAssetBySession(sessionID uint64, duration int64, assetReq *protocol.GhostUserasset, isRenew bool) (*protocol.GhostUserasset, int64, int64, int32) {
	if assetReq == nil || assetReq.Uuid == 0 {
		return nil, 0, 0, errno.HOSTASSETUUIDNOTSET
	}

	now := time.Now().Unix()
	expiredTime := now + duration
	if len(assetReq.Asset) == 0 {
		if isRenew {
			// just only renew
			renewOk := false
			stmt := fmt.Sprintf("UPDATE t_userasset SET expiredtime=%d WHERE uuid=%d AND ghostid=%d AND revision=%d",
				expiredTime, assetReq.Uuid, ghostID, assetReq.Revision)
			dbpool.ExecCB(stmt, func(result sql.Result) error {
				n, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if n > 0 {
					renewOk = true
				}
				return nil
			})
			if !renewOk {
				return nil, 0, 0, errno.HOSTASSETLOCKRENEWFAILED
			}
			log.Debug("ghostid=%d uuid=%d renew ok, expiredtime=%d", ghostID, assetReq.Uuid, expiredTime)
			return nil, 0, 0, errno.OK
		}

		// not renew, try to insert
		var revision uint64 = 0
		insertOk := false
		stmt := fmt.Sprintf("INSERT INTO t_userasset (ghostid, uuid, revision, lockerid, expiredtime) VALUES (%d, %d, %d, %d, %d)", ghostID, assetReq.Uuid, revision, sessionID, expiredTime)
		dbpool.ExecCB(stmt, func(result sql.Result) error {
			n, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if n > 0 {
				insertOk = true
			}
			return nil
		})
		if insertOk {
			newbee := int64(1)
			assetRsp := &protocol.GhostUserasset{
				Uuid:     assetReq.Uuid,
				Revision: revision,
				Asset:    "",
			}
			log.Debug("ghostid=%d uuid=%d insert ok, lockerid=%d revision=%d expiredtime=%d", ghostID, assetReq.Uuid, sessionID, revision, expiredTime)
			return assetRsp, newbee, expiredTime, errno.OK
		}

		// try to lock
		//  1. now > expiredtime
		//  2. lockerid is equal to sessionID
		lockOk := false
		stmt = fmt.Sprintf("UPDATE t_userasset SET lockerid=%d, expiredtime=%d, revision=revision+1 WHERE uuid=%d AND ghostid=%d AND (lockerid=%d OR expiredtime<%d)",
			sessionID, expiredTime, assetReq.Uuid, ghostID, sessionID, expiredTime)
		dbpool.Exec(stmt, func(result sql.Result) error {
			n, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if n > 0 {
				lockOk = true
			}
			return nil
		})
		if !lockOk {
			return nil, 0, 0, errno.HOSTASSETALREADYLOCKED
		}
		log.Debug("ghostid=%d uuid=%d lock ok, lockerid=%d expiredtime=%d", ghostID, assetReq.Uuid, sessionID, expiredTime)

		// try to load
		var data []byte
		stmt = fmt.Sprintf("SELECT revision, asset FROM t_userasset WHERE uuid=%d AND ghostid=%d", assetReq.Uuid, ghostID)
		err := dbpool.QueryCB(stmt, func(rows *sql.Rows) error {
			for rows.Next() {
				return rows.Scan(&revision, &data)
			}
			return nil
		})
		if err != nil {
			log.Error("load ghost=%d uuid=%d failed: %s", ghostID, assetReq.Uuid, err.Error())
			return nil, 0, 0, errno.SYSINTERNALERROR
		}

		log.Debug("ghostid=%d uuid=%d load ok, revision=%d assetLen=%d", ghostID, assetReq.Uuid, revision, len(data))
		assetRsp := &protocol.GhostUserasset{
			Uuid:     assetReq.Uuid,
			Revision: revision,
			Asset:    string(data[:]),
		}
		return assetRsp, 0, expiredTime, errno.OK
	}

	// try to save with lock
	result, err := dbpool.Exec("UPDATE t_userasset SET expiredtime=?, asset=? WHERE uuid=? AND ghostid=? AND revision=?",
		expiredTime, []byte(assetReq.Asset), assetReq.Uuid, ghostID, assetReq.Revision)
	if err != nil {
		log.Error("save ghost=%d uuid=%d failed: %s", ghostID, assetReq.Uuid, err.Error())
		return nil, 0, 0, errno.SYSINTERNALERROR
	}

	n, err := result.RowsAffected()
	if err != nil {
		log.Error("save ghost=%d uuid=%d failed: %s", ghostID, assetReq.Uuid, err.Error())
		return nil, 0, 0, errno.SYSINTERNALERROR
	}

	if n <= 0 {
		// save failed
		return nil, 0, 0, errno.HOSTASSETLOCKLOST
	}

	log.Debug("ghostid=%d uuid=%d revision=%d save assetLen=%d expiredtime=%d", ghostID, assetReq.Uuid, assetReq.Revision, len(assetReq.Asset), expiredTime)
	return nil, int64(0), int64(expiredTime), errno.OK
}

// UnlockAssetBySession unlock the user asset by session id
func UnlockAssetBySession(sessionID uint64, assetReq *protocol.GhostUserasset) int32 {
	if assetReq == nil || assetReq.Uuid == 0 {
		return errno.HOSTASSETUUIDNOTSET
	}

	if len(assetReq.Asset) == 0 {
		// no need to save, just unlock
		stmt := fmt.Sprintf("UPDATE t_userasset SET lockerid=0, expiredtime=0 WHERE uuid=%d AND ghostid=%d AND revision=%d AND lockerid=%d",
			assetReq.Uuid, ghostID, assetReq.Revision, sessionID)
		dbpool.ExecCB(stmt, func(result sql.Result) error {
			return nil
		})
		log.Debug("ghostid=%d uuid=%d unlock revision=%d without save", ghostID, assetReq.Uuid, assetReq.Revision)
		return errno.OK
	}

	// save and unlock
	_, err := dbpool.Exec("UPDATE t_userasset SET lockerid=0, expiredtime=0, asset=? WHERE uuid=? AND ghostid=? AND revision=? AND lockerid=?",
		[]byte(assetReq.Asset), assetReq.Uuid, ghostID, assetReq.Revision, sessionID)
	if err != nil {
		log.Error("save ghost=%d uuid=%d failed: %s", ghostID, assetReq.Uuid, err.Error())
		return errno.SYSINTERNALERROR
	}

	log.Debug("ghostid=%d uuid=%d unlock revision=%d with save assetLen=%d", ghostID, assetReq.Uuid, assetReq.Revision, len(assetReq.Asset))
	return errno.OK
}
