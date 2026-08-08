package handler

import (
	"errors"
	"net/http"

	"github.com/KaueChristian/Goportunitties/dto"
	"github.com/KaueChristian/Goportunitties/service"
	"github.com/gin-gonic/gin"
)

func (h *OpeningHandler) ShowOpeningHandler(ctx *gin.Context) {
	id := ctx.Param("id")

	opening, err := h.service.FindByID(id)
	if err != nil {
		if errors.Is(err, service.ErrOpeningNotFound) {
			dto.SendError(ctx, http.StatusNotFound, "opening not found")
			return
		}
		logger.Errorf("error fetching opening: %v", err.Error())
		dto.SendError(ctx, http.StatusInternalServerError, "error fetching opening")
		return
	}

	dto.SendSuccess(ctx, "show-opening", dto.NewOpeningResponse(opening))
}
