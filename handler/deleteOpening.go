package handler

import (
	"fmt"
	"net/http"

	"github.com/KaueChristian/Goportunitties/schemas"
	"github.com/gin-gonic/gin"
)

func DeleteOpeningHandler(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, errparamIsRequired("id",
			"queryParameter").Error())
		return
	}
	opening := schemas.Opening{}
	//Find Opening
	if err := db.First(&opening, id).Error; err != nil {
		sendError(ctx, http.StatusFound, fmt.Sprintf("opening with id %s not found",
			id))
		return
	}
	//Delete Opening
	if err := db.Delete(&opening).Error; err != nil {
		sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("Error deleting opening with id %v", id))
		return
	}
	sendSucess(ctx,"delete-opening", opening)
}
