package mysql

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbPoolMu sync.Mutex
	dbPool   = map[string]*sql.DB{}
)

func getDB(dsn string) (*sql.DB, error) {
	dbPoolMu.Lock()
	defer dbPoolMu.Unlock()
	if db, ok := dbPool[dsn]; ok {
		return db, nil
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql failed: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql failed: %w", err)
	}
	dbPool[dsn] = db
	return db, nil
}
