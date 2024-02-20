package main

import (
	"github.com/KaueChristian/Goportunitties/config"
	"github.com/KaueChristian/Goportunitties/router"
)

var (
	logger *config.Logger
)

func main() {
	logger = config.GetLogger("main")

	// Initialize Config
	err := config.Init()
	if err != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}
	
	// Initialize Router
	router.Initialize()
}
