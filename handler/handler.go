package handler

import (
	"github.com/KaueChristian/Goportunitties/config"
	"gorm.io/gorm"
)

var (
	logger *config.Logger
	db     *gorm.DB
)

// InitializeHandler initializes the handler package by setting up the logger and database connection.
func InitializeHandler() {
	// Get the logger instance from the config package
	logger = config.GetLogger("handler")
	// Get the SQLite database connection from the config package
	db = config.GetSQlite()
}
