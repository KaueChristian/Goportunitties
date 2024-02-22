package config

import (
	"fmt"

	"gorm.io/gorm"
)

var (
	db *gorm.DB
	logger *Logger
)

// Init initializes the configuration settings.
func Init() error {
	var err error

	// Initialize SQLite database
	db, err = InitializeSQlite()
	if err != nil {
		return fmt.Errorf("error initializing SQLite: %v", err)
	}

	return nil
}

// GetSQlite returns the SQLite database instance.
func GetSQlite() *gorm.DB {
	return db
}

// GetLogger returns a logger instance with the provided prefix.
func GetLogger(p string) *Logger {
	// Initialize logger
	logger := NewLogger(p)
	return logger
}
