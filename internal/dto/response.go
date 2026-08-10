package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the envelope every endpoint replies with.
//
// Message names the operation, Data carries the payload, Meta carries anything
// about the payload (pagination), and Error/Fields describe a failure —
// Fields being the per-field validation messages a form can render inline.
type APIResponse struct {
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    any               `json:"meta,omitempty"`
	Error   string            `json:"error,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Pagination describes the slice of a collection that Data represents.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// NewPagination computes the derived fields from a page window and a total.
func NewPagination(page, pageSize int, total int64) Pagination {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return Pagination{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
}

// SendError replies with a failure and a human-readable reason.
func SendError(ctx *gin.Context, code int, msg string) {
	ctx.JSON(code, APIResponse{Message: "operation failed", Error: msg})
}

// SendValidationError replies with 422 plus the offending fields, so the client
// can highlight the exact inputs instead of showing one generic message.
func SendValidationError(ctx *gin.Context, msg string, fields map[string]string) {
	ctx.JSON(http.StatusUnprocessableEntity, APIResponse{
		Message: "validation failed",
		Error:   msg,
		Fields:  fields,
	})
}

// SendSuccess replies 200 with the payload.
func SendSuccess(ctx *gin.Context, op string, data any) {
	ctx.JSON(http.StatusOK, APIResponse{Message: op + " successful", Data: data})
}

// SendPaginated replies 200 with the payload and its pagination window.
func SendPaginated(ctx *gin.Context, op string, data any, meta Pagination) {
	ctx.JSON(http.StatusOK, APIResponse{Message: op + " successful", Data: data, Meta: meta})
}

// SendCreated replies 201 and points the client at the new resource.
func SendCreated(ctx *gin.Context, op string, data any, location string) {
	ctx.Header("Location", location)
	ctx.JSON(http.StatusCreated, APIResponse{Message: op + " successful", Data: data})
}
