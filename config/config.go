package config

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	logger *Logger

	settings Settings
)

// Settings holds environment-driven configuration for the application.
type Settings struct {
	Port        string
	DBPath      string
	CORSOrigins []string
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadSettings() Settings {
	// Defaults cover the Vite dev server, which runs on a different port than the API.
	origins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")

	return Settings{
		Port:        getEnv("PORT", "8080"),
		DBPath:      getEnv("DB_PATH", "./database/main.db"),
		CORSOrigins: origins,
	}
}

// Init initializes the configuration settings.
func Init() error {
	var err error

	settings = loadSettings()

	// Initialize SQLite database
	db, err = InitializeSQlite(settings.DBPath)
	if err != nil {
		return fmt.Errorf("error initializing SQLite: %v", err)
	}

	return nil
}

// GetSQlite returns the SQLite database instance.
func GetSQlite() *gorm.DB {
	return db
}

// GetSettings returns the loaded application settings.
func GetSettings() Settings {
	return settings
}

// GetLogger returns a logger instance with the provided prefix.
func GetLogger(p string) *Logger {
	// Initialize logger
	logger := NewLogger(p)
	return logger
}
