package mysql

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestQuery(t *testing.T) {
	pool, err := New("root", "123456", "127.0.0.1", "3306", "app_ghost_2")
	if err != nil {
		t.Errorf("mysql init failed: %s", err.Error())
		return
	}

	stmt := fmt.Sprintf("SELECT ghostid, uuid, revision, asset FROM t_userasset WHERE ghostid=%d AND uuid=%d", 2, 10001)
	err = pool.Query(stmt, func(rows *sql.Rows) error {
		for rows.Next() {
			var ghostID uint32 = 0
			var uuid uint64 = 0
			var revision uint64 = 0
			var asset []byte
			err := rows.Scan(&ghostID, &uuid, &revision, &asset)
			if err == nil {
				t.Logf("ghostid=%d, uuid=%d, revision=%d, assetLen=%d", ghostID, uuid, revision, len(asset))
			} else {
				t.Errorf("scan failed: %s", err.Error())
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("query failed: %s", err.Error())
	}
}
