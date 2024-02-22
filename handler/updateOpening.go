package handler

import (
	"net/http"

	"github.com/KaueChristian/Goportunitties/schemas"
	"github.com/gin-gonic/gin"
)

// UpdateOpeningHandler handles the request to update an opening.
func UpdateOpeningHandler(ctx *gin.Context) {
	request := UpdateOpeningrequest{}
	
	ctx.BindJSON(&request)

	if err := request.Validate(); err != nil{
		logger.Errorf("validation error: %v", err.Error())
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	id := ctx.Query("id")
	if id == "" {
		sendError(ctx, http.StatusBadRequest, errparamIsRequired("id", "queryParameter").Error())
		return
	}	
	opening := schemas.Opening{}

	if err := db.First(&opening, id).Error; err != nil {
		sendError(ctx, http.StatusNotFound, "Opening not Found")
		return
	}
	// Update Opening
	if request.Role != "" {
		opening.Role = request.Role
	}
	if request.Company != "" {
		opening.Company = request.Company
	}
	if request.Location != "" {
		opening.Location = request.Location
	}
	if request.Remote != nil {
		opening.Remote = *request.Remote
	}
	if request.Link != "" {
		opening.Link = request.Link
	}
	if request.Salary > 0 {
		opening.Salary = request.Salary
	}
	// Save Opening
	if err := db.Save(&opening).Error;err != nil {
		logger.Errorf("error updating opening: %v", err.Error())
		sendError(ctx, http.StatusInternalServerError, "error updating opening")
		return
	}

	sendSuccess(ctx, "Update-opening", opening)
}