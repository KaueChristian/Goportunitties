package handler

import (
	"net/http"
	"strconv"

	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/gin-gonic/gin"
)

// List handles GET /openings — filtered, sorted and paginated.
func (h *OpeningHandler) List(ctx *gin.Context) {
	query := dto.ListOpeningsQuery{}
	if !bindQuery(ctx, &query) {
		return
	}
	query.Normalize(h.defaultPageSize, h.maxPageSize)

	openings, pagination, err := h.service.List(ctx.Request.Context(), query)
	if err != nil {
		h.fail(ctx, "list-openings", err)
		return
	}

	dto.SendPaginated(ctx, "list-openings", dto.NewOpeningResponseList(openings), pagination)
}

// Facets handles GET /openings/facets — the option counts for the filter
// sidebar, computed against the same query the listing received.
func (h *OpeningHandler) Facets(ctx *gin.Context) {
	query := dto.ListOpeningsQuery{}
	if !bindQuery(ctx, &query) {
		return
	}

	facets, err := h.service.Facets(ctx.Request.Context(), query)
	if err != nil {
		h.fail(ctx, "opening-facets", err)
		return
	}

	dto.SendSuccess(ctx, "opening-facets", facets)
}

// Stats handles GET /openings/stats — whole-index aggregates for the dashboard.
func (h *OpeningHandler) Stats(ctx *gin.Context) {
	stats, err := h.service.Stats(ctx.Request.Context())
	if err != nil {
		h.fail(ctx, "opening-stats", err)
		return
	}

	dto.SendSuccess(ctx, "opening-stats", stats)
}

// Show handles GET /openings/:id.
func (h *OpeningHandler) Show(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	opening, err := h.service.FindByID(ctx.Request.Context(), id)
	if err != nil {
		h.fail(ctx, "show-opening", err)
		return
	}

	dto.SendSuccess(ctx, "show-opening", dto.NewOpeningResponse(opening))
}

// Create handles POST /openings.
func (h *OpeningHandler) Create(ctx *gin.Context) {
	request := dto.OpeningRequest{}
	if !bind(ctx, &request) {
		return
	}

	opening, err := h.service.Create(ctx.Request.Context(), request)
	if err != nil {
		h.fail(ctx, "create-opening", err)
		return
	}

	dto.SendCreated(ctx, "create-opening", dto.NewOpeningResponse(opening), openingURL(ctx, opening.ID))
}

// Replace handles PUT /openings/:id — the full representation, every field
// required, exactly as REST defines PUT.
func (h *OpeningHandler) Replace(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	request := dto.OpeningRequest{}
	if !bind(ctx, &request) {
		return
	}

	opening, err := h.service.Replace(ctx.Request.Context(), id, request)
	if err != nil {
		h.fail(ctx, "replace-opening", err)
		return
	}

	dto.SendSuccess(ctx, "update-opening", dto.NewOpeningResponse(opening))
}

// Patch handles PATCH /openings/:id — only the fields that were sent.
func (h *OpeningHandler) Patch(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	request := dto.PatchOpeningRequest{}
	if !bind(ctx, &request) {
		return
	}
	if request.IsEmpty() {
		dto.SendError(ctx, http.StatusBadRequest, "envie ao menos um campo para atualizar")
		return
	}

	opening, err := h.service.Patch(ctx.Request.Context(), id, request)
	if err != nil {
		h.fail(ctx, "patch-opening", err)
		return
	}

	dto.SendSuccess(ctx, "update-opening", dto.NewOpeningResponse(opening))
}

// Delete handles DELETE /openings/:id. The deleted record comes back so the
// client can offer an undo without refetching.
func (h *OpeningHandler) Delete(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	opening, err := h.service.Delete(ctx.Request.Context(), id)
	if err != nil {
		h.fail(ctx, "delete-opening", err)
		return
	}

	dto.SendSuccess(ctx, "delete-opening", dto.NewOpeningResponse(opening))
}

// openingURL builds the absolute path of an opening, for the Location header.
func openingURL(ctx *gin.Context, id uint) string {
	return ctx.Request.URL.Path + "/" + strconv.FormatUint(uint64(id), 10)
}
