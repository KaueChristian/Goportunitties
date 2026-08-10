// Package database owns the SQLite connection and its schema migration.
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KaueChristian/Goportunitties/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Open connects to the SQLite file at path, creating it (and its directory) on
// first run, and brings the schema up to date.
func Open(path string, verbose bool) (*gorm.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating database directory: %w", err)
		}
	}

	level := gormlogger.Silent
	if verbose {
		level = gormlogger.Warn
	}

	// WAL lets readers work while a write is in flight, and busy_timeout makes
	// a contended write wait instead of failing outright. Both matter the moment
	// a background job writes while the API is serving requests.
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", path)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(level),
	})
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("reading sql.DB handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)

	if err := Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate applies the schema. Exported so tests can build an in-memory database
// through the same path production uses.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Opening{}); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}

// Close releases the underlying connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
