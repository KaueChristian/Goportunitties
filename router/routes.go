package router

import (
	"github.com/KaueChristian/Goportunitties/handler"
	"github.com/gin-gonic/gin"
)

// initializeRoutes initializes routes for the provided Gin router.
func initializeRoutes(router *gin.Engine) {
	// Initialize the handler
	handler.InitializeHandler()
	// Grouping routes under /api/v1
	v1 := router.Group("/api/v1")
	{
		// Define endpoints for opening CRUD operations
		v1.GET("/opening", handler.ShowOpeningHandler)    // Show a specific opening
		v1.POST("/opening", handler.CreateOpeningHandler) // Create a new opening
		v1.DELETE("/opening", handler.DeleteOpeningHandler) // Delete a specific opening
		v1.PUT("/opening", handler.UpdateOpeningHandler) // Update a specific opening
		v1.GET("/openings", handler.ListOpeningHandler)  // List all openings
	}
}
