package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// sendError sends an error response with the given status code and message.
func sendError(ctx *gin.Context, code int, msg string) {
	// Set response content type
	ctx.Header("Content-Type", "application/json")
	// Send JSON response with error message and code
	ctx.JSON(code, gin.H{
		"MSG":       msg,
		"ERRORCODE": code,
	})
}

// sendSuccess sends a success response with the given operation name and data.
func sendSuccess(ctx *gin.Context, op string, data interface{}) {
	// Set response content type
	ctx.Header("Content-Type", "application/json")
	// Send JSON response with success message, operation name, and data
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  fmt.Sprintf("operation from handler: %s successful", op),
		"data": data,
	})
}
