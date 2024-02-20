package config

import (
	"errors"

	"gorm.io/gorm"
)

var (
	db 		*gorm.DB
	logger	*Logger
)

func Init() error{
	return errors.New("critical error") 
}

func GetLogger(p string) *Logger {
	//Initialize Logger
	logger := NewLogger(p)
	return logger
}
