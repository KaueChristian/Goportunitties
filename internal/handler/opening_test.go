package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/KaueChristian/Goportunitties/internal/handler"
	"github.com/KaueChristian/Goportunitties/internal/model"
	"github.com/KaueChristian/Goportunitties/internal/service"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// ctxType keeps the stub's long method set readable.
type ctxType = context.Context

// stubService lets each test decide what the business layer answers, so the
// handler is exercised on exactly one thing: turning that answer into HTTP.
type stubService struct {
	opening    *model.Opening
	pagination dto.Pagination
	facets     dto.OpeningFacets
	stats      dto.OpeningStats
	err        error

	lastQuery dto.ListOpeningsQuery
	lastID    uint
	lastReq   dto.OpeningRequest
	lastPatch dto.PatchOpeningRequest
}

func (s *stubService) Create(_ ctxType, req dto.OpeningRequest) (*model.Opening, error) {
	s.lastReq = req
	return s.opening, s.err
}

func (s *stubService) FindByID(_ ctxType, id uint) (*model.Opening, error) {
	s.lastID = id
	return s.opening, s.err
}

func (s *stubService) List(_ ctxType, query dto.ListOpeningsQuery) ([]model.Opening, dto.Pagination, error) {
	s.lastQuery = query
	if s.err != nil {
		return nil, dto.Pagination{}, s.err
	}
	if s.opening == nil {
		return []model.Opening{}, s.pagination, nil
	}
	return []model.Opening{*s.opening}, s.pagination, nil
}

func (s *stubService) Facets(_ ctxType, query dto.ListOpeningsQuery) (dto.OpeningFacets, error) {
	s.lastQuery = query
	return s.facets, s.err
}

func (s *stubService) Stats(ctxType) (dto.OpeningStats, error) {
	return s.stats, s.err
}

func (s *stubService) Replace(_ ctxType, id uint, req dto.OpeningRequest) (*model.Opening, error) {
	s.lastID, s.lastReq = id, req
	return s.opening, s.err
}

func (s *stubService) Patch(_ ctxType, id uint, req dto.PatchOpeningRequest) (*model.Opening, error) {
	s.lastID, s.lastPatch = id, req
	return s.opening, s.err
}

func (s *stubService) Delete(_ ctxType, id uint) (*model.Opening, error) {
	s.lastID = id
	return s.opening, s.err
}

func newServer(svc service.OpeningService) *gin.Engine {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	openings := handler.NewOpeningHandler(svc, log, 12, 100)

	engine := gin.New()
	engine.GET("/openings", openings.List)
	engine.GET("/openings/facets", openings.Facets)
	engine.GET("/openings/stats", openings.Stats)
	engine.POST("/openings", openings.Create)
	engine.GET("/openings/:id", openings.Show)
	engine.PUT("/openings/:id", openings.Replace)
	engine.PATCH("/openings/:id", openings.Patch)
	engine.DELETE("/openings/:id", openings.Delete)
	return engine
}

func do(t *testing.T, engine *gin.Engine, method, target, body string) (*httptest.ResponseRecorder, dto.APIResponse) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	response := dto.APIResponse{}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decoding response %q: %v", rec.Body.String(), err)
		}
	}
	return rec, response
}

func sampleOpening() *model.Opening {
	opening := &model.Opening{
		Role:     "Desenvolvedor Go",
		Company:  "Acme",
		Location: "São Paulo, SP",
		Remote:   true,
		Link:     "https://acme.com/vagas/1",
		Salary:   15000,
	}
	opening.ID = 7
	return opening
}

const validBody = `{"role":"Desenvolvedor Go","company":"Acme","location":"São Paulo, SP","remote":true,"link":"https://acme.com/vagas/1","salary":15000}`

func TestCreateReturns201AndLocation(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}

	rec, response := do(t, newServer(svc), http.MethodPost, "/openings", validBody)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/openings/7" {
		t.Fatalf("Location = %q, want /openings/7", got)
	}
	if response.Data == nil {
		t.Fatal("response should carry the created opening")
	}
}

func TestCreateAcceptsRemoteFalse(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}
	body := strings.Replace(validBody, `"remote":true`, `"remote":false`, 1)

	rec, _ := do(t, newServer(svc), http.MethodPost, "/openings", body)

	// A plain bool would make `required` reject every on-site opening; the
	// pointer is what keeps this a 201.
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if svc.lastReq.Remote == nil || *svc.lastReq.Remote {
		t.Fatalf("remote reached the service as %v, want a pointer to false", svc.lastReq.Remote)
	}
}

func TestCreateAcceptsSalaryZero(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}
	body := strings.Replace(validBody, `"salary":15000`, `"salary":0`, 1)

	rec, _ := do(t, newServer(svc), http.MethodPost, "/openings", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 — a salary of 0 means 'a combinar'", rec.Code)
	}
}

func TestCreateRejectsMissingFieldsWithPerFieldMessages(t *testing.T) {
	rec, response := do(t, newServer(&stubService{}), http.MethodPost, "/openings", `{"role":"Go"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	for _, field := range []string{"company", "location", "remote", "link"} {
		if _, ok := response.Fields[field]; !ok {
			t.Fatalf("missing field error for %q: %+v", field, response.Fields)
		}
	}
}

func TestCreateRejectsNonHTTPLink(t *testing.T) {
	tests := []string{
		"javascript:alert(1)",
		"ftp://acme.com/vaga",
		"nem-uma-url",
	}

	for _, link := range tests {
		t.Run(link, func(t *testing.T) {
			body := strings.Replace(validBody, "https://acme.com/vagas/1", link, 1)

			rec, response := do(t, newServer(&stubService{}), http.MethodPost, "/openings", body)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422 — the frontend renders this value as an href", rec.Code)
			}
			if _, ok := response.Fields["link"]; !ok {
				t.Fatalf("expected a link error, got %+v", response.Fields)
			}
		})
	}
}

func TestCreateRejectsMalformedJSON(t *testing.T) {
	rec, response := do(t, newServer(&stubService{}), http.MethodPost, "/openings", `{"role":`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 — broken JSON is not a validation failure", rec.Code)
	}
	if response.Fields != nil {
		t.Fatalf("a malformed body has no per-field errors, got %+v", response.Fields)
	}
}

func TestShowNotFound(t *testing.T) {
	svc := &stubService{err: service.ErrOpeningNotFound}

	rec, response := do(t, newServer(svc), http.MethodGet, "/openings/7", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if response.Error == "" {
		t.Fatal("a 404 should explain itself")
	}
}

func TestShowRejectsNonNumericID(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}

	rec, _ := do(t, newServer(svc), http.MethodGet, "/openings/abc", "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 — a malformed id never reaches the database", rec.Code)
	}
	if svc.lastID != 0 {
		t.Fatal("the service should not have been called")
	}
}

func TestUnexpectedErrorsBecome500WithoutLeakingDetails(t *testing.T) {
	svc := &stubService{err: errors.New("pq: connection refused to 10.0.0.4")}

	rec, response := do(t, newServer(svc), http.MethodGet, "/openings/7", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(response.Error, "10.0.0.4") {
		t.Fatalf("internal details leaked to the client: %q", response.Error)
	}
}

func TestPatchRejectsAnEmptyBody(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}

	rec, _ := do(t, newServer(svc), http.MethodPatch, "/openings/7", `{}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestPatchForwardsOnlyTheFieldsThatWereSent(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}

	rec, _ := do(t, newServer(svc), http.MethodPatch, "/openings/7", `{"salary":0,"remote":false}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.lastPatch.Role != nil || svc.lastPatch.Company != nil {
		t.Fatalf("absent fields reached the service: %+v", svc.lastPatch)
	}
	if svc.lastPatch.Salary == nil || *svc.lastPatch.Salary != 0 {
		t.Fatalf("salary = %v, want a pointer to 0", svc.lastPatch.Salary)
	}
	if svc.lastPatch.Remote == nil || *svc.lastPatch.Remote {
		t.Fatalf("remote = %v, want a pointer to false", svc.lastPatch.Remote)
	}
}

func TestReplaceRequiresEveryField(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}

	rec, _ := do(t, newServer(svc), http.MethodPut, "/openings/7", `{"role":"SRE"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 — PUT replaces the whole representation", rec.Code)
	}
}

func TestDeleteReturnsTheRemovedOpening(t *testing.T) {
	svc := &stubService{opening: sampleOpening()}

	rec, response := do(t, newServer(svc), http.MethodDelete, "/openings/7", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.lastID != 7 {
		t.Fatalf("service received id %d, want 7", svc.lastID)
	}
	if response.Data == nil {
		t.Fatal("delete should echo the removed record so the client can undo")
	}
}

func TestListAppliesDefaultsAndReturnsPaginationMeta(t *testing.T) {
	svc := &stubService{
		opening:    sampleOpening(),
		pagination: dto.Pagination{Page: 1, PageSize: 12, Total: 57, TotalPages: 5},
	}

	rec, response := do(t, newServer(svc), http.MethodGet, "/openings", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.lastQuery.Page != 1 || svc.lastQuery.PageSize != 12 || svc.lastQuery.Sort != dto.SortRecent {
		t.Fatalf("defaults were not applied: %+v", svc.lastQuery)
	}
	if response.Meta == nil {
		t.Fatal("a paginated listing must report its window")
	}
}

func TestListParsesTheFullQueryString(t *testing.T) {
	svc := &stubService{}

	rec, _ := do(t, newServer(svc), http.MethodGet,
		"/openings?search=go&location=Curitiba%2C+PR&remote=true&minSalary=8000&sort=salary-desc&page=2&pageSize=20", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	query := svc.lastQuery
	if query.Search != "go" || query.Location != "Curitiba, PR" || query.MinSalary != 8000 {
		t.Fatalf("filters = %+v", query)
	}
	if query.Remote == nil || !*query.Remote {
		t.Fatalf("remote = %v, want a pointer to true", query.Remote)
	}
	if query.Sort != dto.SortSalaryDesc || query.Page != 2 || query.PageSize != 20 {
		t.Fatalf("sort/paging = %+v", query)
	}
}

func TestListRejectsInvalidFilters(t *testing.T) {
	tests := []struct {
		name  string
		query string
		field string
	}{
		{"unknown sort", "?sort=random", "sort"},
		{"negative salary", "?minSalary=-1", "minSalary"},
		{"oversized search", "?search=" + strings.Repeat("a", 200), "search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, response := do(t, newServer(&stubService{}), http.MethodGet, "/openings"+tt.query, "")

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", rec.Code)
			}
			// The field name must be the one the client sent, not the Go one.
			if _, ok := response.Fields[tt.field]; !ok {
				t.Fatalf("expected an error on %q, got %+v", tt.field, response.Fields)
			}
		})
	}
}

// Paging is forgiving where filters are strict: a nonsensical window is clamped
// into range and reported back in the meta, never rejected.
func TestListClampsThePagingWindow(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		wantPage     int
		wantPageSize int
	}{
		{"page size above the cap", "?pageSize=5000", 1, 100},
		{"page zero", "?page=0", 1, 12},
		{"negative page", "?page=-4", 1, 12},
		{"page size zero", "?pageSize=0", 1, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubService{}

			rec, _ := do(t, newServer(svc), http.MethodGet, "/openings"+tt.query, "")

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if svc.lastQuery.Page != tt.wantPage || svc.lastQuery.PageSize != tt.wantPageSize {
				t.Fatalf("window = page %d size %d, want %d/%d",
					svc.lastQuery.Page, svc.lastQuery.PageSize, tt.wantPage, tt.wantPageSize)
			}
		})
	}
}

func TestFacetsAndStatsAreServed(t *testing.T) {
	svc := &stubService{
		facets: dto.OpeningFacets{Remote: dto.RemoteCounts{All: 3, Remote: 2, Onsite: 1}, SalaryCeiling: 22000},
		stats:  dto.OpeningStats{Total: 3, Remote: 2, Onsite: 1},
	}
	engine := newServer(svc)

	// These two sit under /openings alongside /openings/:id — this is also the
	// test that the static segments win over the wildcard.
	if rec, _ := do(t, engine, http.MethodGet, "/openings/facets", ""); rec.Code != http.StatusOK {
		t.Fatalf("facets status = %d, want 200", rec.Code)
	}
	if rec, _ := do(t, engine, http.MethodGet, "/openings/stats", ""); rec.Code != http.StatusOK {
		t.Fatalf("stats status = %d, want 200", rec.Code)
	}
}
