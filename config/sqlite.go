package config

import (
	"os"

	"github.com/KaueChristian/Goportunitties/schemas"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitializeSQlite initializes the SQLite database.
func InitializeSQlite() (*gorm.DB, error) {
	// Get logger instance for SQLite
	logger := GetLogger("Sqlite")
	// Path to the SQLite database file
	dbPath := "./database/main.db"

	// Check if the database file exists
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		// If the database file doesn't exist, create it
		logger.Info("Database file not found. Creating...")
		// Create database file and directory
		err = os.MkdirAll("./database", os.ModePerm)
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
