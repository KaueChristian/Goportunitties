package handler

import (
	"errors"
	"net/http"

	"github.com/KaueChristian/Goportunitties/dto"
	"github.com/KaueChristian/Goportunitties/service"
	"github.com/gin-gonic/gin"
)

func (h *OpeningHandler) UpdateOpeningHandler(ctx *gin.Context) {
	request := dto.UpdateOpeningRequest{}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		logger.Errorf("invalid request body: %v", err.Error())
		dto.SendError(ctx, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := request.Validate(); err != nil {
		logger.Errorf("validation error: %v", err.Error())
		dto.SendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	id := ctx.Param("id")

	opening, err := h.service.Update(id, &request)
	if err != nil {
		if errors.Is(err, service.ErrOpeningNotFound) {
			dto.SendError(ctx, http.StatusNotFound, "opening not found")
			return
		}
		logger.Errorf("error updating opening: %v", err.Error())
		dto.SendError(ctx, http.StatusInternalServerError, "error updating opening")
		return
	}

	dto.SendSuccess(ctx, "update-opening", dto.NewOpeningResponse(opening))
}
