package router

import (
	"github.com/KaueChristian/Goportunitties/handler"
	"github.com/gin-gonic/gin"
)

// initializeRoutes initializes routes for the provided Gin router.
func initializeRoutes(router *gin.Engine) {
	openingHandler := handler.NewOpeningHandler()

	// Grouping routes under /api/v1
	v1 := router.Group("/api/v1")
	{
		// Define endpoints for opening CRUD operations
		v1.GET("/opening/:id", openingHandler.ShowOpeningHandler)      // Show a specific opening
		v1.POST("/opening", openingHandler.CreateOpeningHandler)       // Create a new opening
		v1.DELETE("/opening/:id", openingHandler.DeleteOpeningHandler) // Delete a specific opening
		v1.PUT("/opening/:id", openingHandler.UpdateOpeningHandler)    // Update a specific opening
		v1.GET("/openings", openingHandler.ListOpeningHandler)         // List all openings
	}
}
