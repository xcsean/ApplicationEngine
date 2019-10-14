package mysql

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestQueryCB(t *testing.T) {
	pool, err := New("root", "123456", "192.168.95.182", "3306", "app_ghost_1")
	if err != nil {
		t.Errorf("mysql init failed: %s", err.Error())
		return
	}

	stmt := fmt.Sprintf("SELECT * FROM t_userasset WHERE ghostid=%d AND uuid=%d", 1, 10001)
	err = pool.Query(stmt, func(rows *sql.Rows) error {
		for rows.Next() {
			var ghostID uint32 = 0
			var uuid uint64 = 0
			var revision uint64 = 0
			var asset []byte
			err := rows.Scan(&ghostID, &uuid, &revision, &asset)
			if err == nil {
				t.Logf("ghostid=%d, uuid=%d, revision=%d, asset len=%d", ghostID, uuid, revision, len(asset))
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
