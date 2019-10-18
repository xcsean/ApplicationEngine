package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// for need of database/sql
	_ "github.com/go-sql-driver/mysql"
)

// DB database wrapper
type DB struct {
	pool   *sql.DB
	queryT time.Duration
	execT  time.Duration
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
	pool.SetConnMaxLifetime(10 * time.Minute)
	for i := 0; i < 5; i++ {
		if err = pool.Ping(); err != nil {
			return nil, fmt.Errorf("pool.Ping(%d) failed: %s", i, err.Error())
		}
	}

	return &DB{pool: pool, queryT: 3 * time.Second, execT: 3 * time.Second}, nil
}

// SetQueryTimeout set the timeout of query
func (db *DB) SetQueryTimeout(timeout time.Duration) {
	db.queryT = timeout
}

// Query do a query statement with callback
func (db *DB) Query(stmt string, cb func(*sql.Rows) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.queryT)
	defer cancel()

	rows, err := db.pool.QueryContext(ctx, stmt)
	if err != nil {
		return err
	}
	defer rows.Close()

	return cb(rows)
}

// SetExecTimeout set the timeout of exec
func (db *DB) SetExecTimeout(timeout time.Duration) {
	db.execT = timeout
}

// Exec do a exec statement
func (db *DB) Exec(stmt string, cb func(sql.Result) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.execT)
	defer cancel()

	result, err := db.pool.ExecContext(ctx, stmt)
	if err != nil {
		return err
	}

	return cb(result)
}

// ExecDirect exec directly
func (db *DB) ExecDirect(query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.execT)
	defer cancel()

	return db.pool.ExecContext(ctx, query, args...)
}
