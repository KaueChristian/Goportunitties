package handler

import (
	
	"github.com/KaueChristian/Goportunitties/config"
	"gorm.io/gorm"
)

var(
	logger	*config.Logger
	db		*gorm.DB
)

func InitializeHandler() {
	logger = config.GetLogger("handler")
	db = config.GetSQlite()
}