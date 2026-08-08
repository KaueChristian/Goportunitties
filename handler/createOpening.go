package handler

import (
	"net/http"

	"github.com/KaueChristian/Goportunitties/dto"
	"github.com/gin-gonic/gin"
)

func (h *OpeningHandler) CreateOpeningHandler(ctx *gin.Context) {
	request := dto.CreateOpeningRequest{}

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

	opening, err := h.service.Create(&request)
	if err != nil {
		logger.Errorf("error creating opening: %v", err.Error())
		dto.SendError(ctx, http.StatusInternalServerError, "error creating opening on database")
		return
	}

	dto.SendSuccess(ctx, "create-opening", dto.NewOpeningResponse(opening))
}
