package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateOpeningHandler handles the request to update an opening.
func UpdateOpeningHandler(ctx *gin.Context) {
	// Send a JSON response indicating that the opening is being updated
	ctx.JSON(http.StatusOK, gin.H{
		"message": "UPDATE Opening",
	})
}
