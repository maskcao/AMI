package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	dataDir string
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "store"), 0o755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "ami.db")
	sqlDB, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	d := &DB{DB: sqlDB, dataDir: dataDir}
	if err := d.Migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err := d.ensureSchemaPatches(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err := d.tune(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) DataDir() string { return d.dataDir }

func (d *DB) tune() error {
	_, err := d.Exec(`PRAGMA synchronous=NORMAL; PRAGMA cache_size=-64000; PRAGMA busy_timeout=5000;`)
	return err
}

func (d *DB) CountIntel() (int, error) {
	var n int
	err := d.QueryRow(`SELECT COUNT(*) FROM intel_report`).Scan(&n)
	return n, err
}

func (d *DB) CountKnowledge() (int, error) {
	var n int
	err := d.QueryRow(`SELECT COUNT(*) FROM attack_knowledge`).Scan(&n)
	return n, err
}

func (d *DB) IsEmpty() (bool, error) {
	n, err := d.CountIntel()
	if err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}
	k, err := d.CountKnowledge()
	if err != nil {
		return false, err
	}
	return k == 0, nil
}

func (d *DB) StoreRoot() string {
	return filepath.Join(d.dataDir, "store")
}

func IntelDir(intelID int64) string {
	return fmt.Sprintf("intel/%d", intelID)
}
