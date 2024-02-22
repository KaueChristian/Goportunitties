package handler

import (
	"net/http"

	"github.com/KaueChristian/Goportunitties/schemas"
	"github.com/gin-gonic/gin"
)

func ListOpeningHandler(ctx *gin.Context) {
	openings := []schemas.Opening{}

	if err := db.Find(&openings).Error; err != nil {
		sendError(ctx, http.StatusInternalServerError, "error Listing openings")
		return
	}

	sendSucess(ctx, "List-openings", openings)
}
