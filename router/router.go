package router

import "github.com/gin-gonic/gin"

// Initialize sets up and starts the router.
func Initialize() {
	// Initialize the Gin router with default middleware
	router := gin.Default()

	// Initialize routes
	initializeRoutes(router)

	// Running API on port 8080
	router.Run(":8080")
}
