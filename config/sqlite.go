package config

import (
	"os"
	"path/filepath"

	"github.com/KaueChristian/Goportunitties/schemas"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// InitializeSQlite initializes the SQLite database at the given path.
func InitializeSQlite(dbPath string) (*gorm.DB, error) {
	// Get logger instance for SQLite
	logger := GetLogger("Sqlite")

	// Check if the database file exists
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		// If the database file doesn't exist, create it
		logger.Info("Database file not found. Creating...")
		// Create database file and directory
		err = os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)
		if err != nil {
			return nil, err
		}
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	// Create and connect to the database
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Errorf("SQLite Opening Error: %v", err)
		return nil, err
	}

	// Run auto migration to create necessary tables
	err = db.AutoMigrate(&schemas.Opening{})
	if err != nil {
		logger.Errorf("SQLite auto migration error: %v", err)
		return nil, err
	}

	return db, nil
}
