package handler

import (
	"errors"
	"net/http"

	"github.com/KaueChristian/Goportunitties/dto"
	"github.com/KaueChristian/Goportunitties/service"
	"github.com/gin-gonic/gin"
)

func (h *OpeningHandler) DeleteOpeningHandler(ctx *gin.Context) {
	id := ctx.Param("id")

	opening, err := h.service.Delete(id)
	if err != nil {
		if errors.Is(err, service.ErrOpeningNotFound) {
			dto.SendError(ctx, http.StatusNotFound, "opening not found")
			return
		}
		logger.Errorf("error deleting opening: %v", err.Error())
		dto.SendError(ctx, http.StatusInternalServerError, "error deleting opening")
		return
	}

	dto.SendSuccess(ctx, "delete-opening", dto.NewOpeningResponse(opening))
}
