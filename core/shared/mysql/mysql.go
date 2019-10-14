package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB database wrapper
type DB struct {
	pool    *sql.DB
	timeout time.Duration
}

// New new a mysql instance
func New(username, password, ip, port, database string) (*DB, error) {
	dbStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", username, password, ip, port, database)
	pool, err := sql.Open("mysql", dbStr)
	if err != nil {
		return nil, fmt.Errorf("sql.Open(%s) failed: %v", dbStr, pool)
	}

	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)
	for i := 0; i < 5; i++ {
		if err = pool.Ping(); err != nil {
			return nil, fmt.Errorf("pool.Ping(%d) failed: %s", i, err.Error())
		}
	}

	return &DB{pool: pool, timeout: 5}, nil
}

// Query do a query statement
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	return db.pool.QueryContext(ctx, query, args...)
}

// Exec do a exec statement
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.timeout)
	defer cancel()
	return db.pool.ExecContext(ctx, query, args...)
}
