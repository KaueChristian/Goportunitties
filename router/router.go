package router

import (
	"github.com/KaueChristian/Goportunitties/config"
	"github.com/KaueChristian/Goportunitties/middleware"
	"github.com/gin-gonic/gin"
)

// Initialize sets up and starts the router.
func Initialize() {
	// Initialize the Gin router with default middleware
	router := gin.Default()

	// Allow the web frontend (served from another origin in development) to call the API
	router.Use(middleware.CORS(config.GetSettings().CORSOrigins))

	// Initialize routes
	initializeRoutes(router)

	// Running API on the configured port
	router.Run(":" + config.GetSettings().Port)
}
