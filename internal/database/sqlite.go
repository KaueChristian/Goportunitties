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

// identityIndex is the unique key behind idempotent ingestion: re-running a
// source updates the matching row instead of inserting a copy.
const identityIndex = `CREATE UNIQUE INDEX IF NOT EXISTS idx_openings_source_identity
	ON openings(source, external_id)`

// Migrate applies the schema. Exported so tests can build a throwaway database
// through the same path production uses.
//
// The three steps run in this order on purpose. Creating the unique index
// before the backfill would fail on any database that already has rows: they
// would all share an empty external_id, and a unique index cannot be built over
// duplicates. Columns first, identities second, constraint last.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Opening{}); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	if err := backfillIdentities(db); err != nil {
		return err
	}

	if err := db.Exec(identityIndex).Error; err != nil {
		return fmt.Errorf("creating the identity index: %w", err)
	}

	return nil
}

// backfillIdentities gives an identity to rows that predate provenance.
//
// Those rows were all typed in by hand, and their primary key is already unique
// within the table, so it doubles as their identity. Raw SQL is deliberate: it
// ignores the soft-delete scope, and deleted rows still occupy the index.
func backfillIdentities(db *gorm.DB) error {
	err := db.Exec(`
		UPDATE openings
		SET source = ?, external_id = 'legacy-' || id
		WHERE external_id IS NULL OR external_id = ''`,
		model.SourceManual,
	).Error
	if err != nil {
		return fmt.Errorf("backfilling opening identities: %w", err)
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
