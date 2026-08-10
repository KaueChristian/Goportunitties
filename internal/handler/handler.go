// Package handler adapts HTTP to the service layer: it parses the request,
// calls a service, and maps the result — or the error — onto a status code.
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/KaueChristian/Goportunitties/internal/service"
	"github.com/gin-gonic/gin"
)

// OpeningHandler serves the opening endpoints.
type OpeningHandler struct {
	service         service.OpeningService
	log             *slog.Logger
	defaultPageSize int
	maxPageSize     int
}

// NewOpeningHandler wires a handler to its service and paging limits.
func NewOpeningHandler(svc service.OpeningService, log *slog.Logger, defaultPageSize, maxPageSize int) *OpeningHandler {
	return &OpeningHandler{
		service:         svc,
		log:             log,
		defaultPageSize: defaultPageSize,
		maxPageSize:     maxPageSize,
	}
}

// parseID reads the :id path parameter. A non-numeric id is a malformed
// request (400), not a missing record (404).
func parseID(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		dto.SendError(ctx, http.StatusBadRequest, "id inválido")
		return 0, false
	}
	return uint(id), true
}

// bind decodes and validates a request body, replying on failure. It separates
// "this isn't valid JSON" (400) from "these fields are wrong" (422).
func bind(ctx *gin.Context, target any) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		if fields := dto.FieldErrors(err); fields != nil {
			dto.SendValidationError(ctx, "há campos inválidos na requisição", fields)
			return false
		}
		dto.SendError(ctx, http.StatusBadRequest, "corpo da requisição inválido")
		return false
	}
	return true
}

// bindQuery decodes and validates the query string.
func bindQuery(ctx *gin.Context, target any) bool {
	if err := ctx.ShouldBindQuery(target); err != nil {
		if fields := dto.FieldErrors(err); fields != nil {
			dto.SendValidationError(ctx, "há parâmetros inválidos na consulta", fields)
			return false
		}
		dto.SendError(ctx, http.StatusBadRequest, "parâmetros de consulta inválidos")
		return false
	}
	return true
}

// fail maps a service error onto a response: a known domain error becomes its
// status code, anything else is logged and reported as a 500 without leaking
// the underlying message.
func (h *OpeningHandler) fail(ctx *gin.Context, op string, err error) {
	if errors.Is(err, service.ErrOpeningNotFound) {
		dto.SendError(ctx, http.StatusNotFound, "vaga não encontrada")
		return
	}

	h.log.ErrorContext(ctx.Request.Context(), "request failed",
		slog.String("operation", op),
		slog.String("error", err.Error()),
	)
	dto.SendError(ctx, http.StatusInternalServerError, "erro interno ao processar a requisição")
}
