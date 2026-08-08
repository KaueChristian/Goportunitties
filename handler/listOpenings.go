package handler

import (
	"net/http"

	"github.com/KaueChristian/Goportunitties/dto"
	"github.com/gin-gonic/gin"
)

// ListOpeningHandler handles the request to list all openings.
func (h *OpeningHandler) ListOpeningHandler(ctx *gin.Context) {
	openings, err := h.service.FindAll()
	if err != nil {
		logger.Errorf("error listing openings: %v", err.Error())
		dto.SendError(ctx, http.StatusInternalServerError, "error listing openings")
		return
	}

	dto.SendSuccess(ctx, "list-openings", dto.NewOpeningResponseList(openings))
}
