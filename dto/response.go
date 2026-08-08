package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for every API response.
type APIResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SendError sends a standardized error response with the given status code and message.
func SendError(ctx *gin.Context, code int, msg string) {
	ctx.JSON(code, APIResponse{
		Message: "operation failed",
		Error:   msg,
	})
}

// SendSuccess sends a standardized success response with an optional operation name and data.
func SendSuccess(ctx *gin.Context, op string, data any) {
	ctx.JSON(http.StatusOK, APIResponse{
		Message: op + " successful",
		Data:    data,
	})
}
