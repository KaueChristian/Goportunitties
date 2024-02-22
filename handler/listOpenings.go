package handler

import (
	"net/http"

	"github.com/KaueChristian/Goportunitties/schemas"
	"github.com/gin-gonic/gin"
)

// ListOpeningHandler handles the request to list all openings.
func ListOpeningHandler(ctx *gin.Context) {
	// Create a slice to hold the openings
	openings := []schemas.Opening{}

	// Find all openings from the database
	if err := db.Find(&openings).Error; err != nil {
		// If there's an error, send an internal server error response
		sendError(ctx, http.StatusInternalServerError, "error Listing openings")
		return
	}

	// Send a success response with the list of openings
	sendSuccess(ctx, "List-openings", openings)
}