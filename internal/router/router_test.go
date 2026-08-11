package router_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/KaueChristian/Goportunitties/internal/config"
	"github.com/KaueChristian/Goportunitties/internal/database"
	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/KaueChristian/Goportunitties/internal/middleware"
	"github.com/KaueChristian/Goportunitties/internal/repository"
	"github.com/KaueChristian/Goportunitties/internal/router"
	"github.com/KaueChristian/Goportunitties/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// newAPI builds the real engine over a throwaway database: same middleware,
// same routes, same wiring as production.
func newAPI(t *testing.T, spa *fstest.MapFS) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })

	deps := router.Deps{
		Settings: config.Settings{
			Env:             "test",
			CORSOrigins:     []string{"http://localhost:5173"},
			DefaultPageSize: 12,
			MaxPageSize:     100,
		},
		Log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:      db,
		Service: service.NewOpeningService(repository.NewOpeningRepository(db)),
		Version: "test",
	}
	if spa != nil {
		deps.SPA = spa
	}

	return router.New(deps), db
}

func request(t *testing.T, engine *gin.Engine, method, target, body string) (*httptest.ResponseRecorder, dto.APIResponse) {
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
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
	}
	return rec, response
}

const newOpening = `{"role":"Desenvolvedor Go","company":"Acme","location":"São Paulo, SP","remote":true,"link":"https://acme.com/vagas/1","salary":15000}`

// The whole contract, exercised the way a client does.
func TestOpeningLifecycle(t *testing.T) {
	engine, _ := newAPI(t, nil)

	rec, created := request(t, engine, http.MethodPost, "/api/v1/openings", newOpening)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: status %d, want 201 (%s)", rec.Code, rec.Body)
	}
	location := rec.Header().Get("Location")
	if location != "/api/v1/openings/1" {
		t.Fatalf("Location = %q, want /api/v1/openings/1", location)
	}

	opening := created.Data.(map[string]any)
	if opening["role"] != "Desenvolvedor Go" || opening["salary"].(float64) != 15000 {
		t.Fatalf("created opening = %+v", opening)
	}

	// The Location header must actually resolve.
	rec, _ = request(t, engine, http.MethodGet, location, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("following Location: status %d, want 200", rec.Code)
	}

	rec, _ = request(t, engine, http.MethodPatch, location, `{"salary":18000}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch: status %d, want 200 (%s)", rec.Code, rec.Body)
	}

	rec, shown := request(t, engine, http.MethodGet, location, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("show: status %d, want 200", rec.Code)
	}
	if shown.Data.(map[string]any)["salary"].(float64) != 18000 {
		t.Fatalf("patch did not persist: %+v", shown.Data)
	}
	if shown.Data.(map[string]any)["company"] != "Acme" {
		t.Fatal("patch overwrote a field it was not given")
	}

	rec, _ = request(t, engine, http.MethodDelete, location, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: status %d, want 200", rec.Code)
	}

	rec, _ = request(t, engine, http.MethodGet, location, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("show after delete: status %d, want 404", rec.Code)
	}
}

func TestListPaginatesAndFilters(t *testing.T) {
	engine, _ := newAPI(t, nil)

	for _, body := range []string{
		newOpening,
		`{"role":"Desenvolvedor React","company":"Globex","location":"São Paulo, SP","remote":false,"link":"https://globex.com/2","salary":9000}`,
		`{"role":"SRE","company":"Initech","location":"Curitiba, PR","remote":true,"link":"https://initech.com/3","salary":21500}`,
	} {
		if rec, _ := request(t, engine, http.MethodPost, "/api/v1/openings", body); rec.Code != http.StatusCreated {
			t.Fatalf("seeding: status %d (%s)", rec.Code, rec.Body)
		}
	}

	rec, response := request(t, engine, http.MethodGet, "/api/v1/openings?pageSize=2&sort=salary-desc", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list: status %d", rec.Code)
	}

	items := response.Data.([]any)
	if len(items) != 2 {
		t.Fatalf("page holds %d openings, want 2", len(items))
	}
	if items[0].(map[string]any)["role"] != "SRE" {
		t.Fatalf("sorting ignored: first item is %+v", items[0])
	}

	meta := response.Meta.(map[string]any)
	if meta["total"].(float64) != 3 || meta["totalPages"].(float64) != 2 {
		t.Fatalf("meta = %+v, want 3 items over 2 pages", meta)
	}

	rec, response = request(t, engine, http.MethodGet, "/api/v1/openings?remote=true&minSalary=20000", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("filtered list: status %d", rec.Code)
	}
	if items := response.Data.([]any); len(items) != 1 {
		t.Fatalf("filtered list returned %d openings, want 1", len(items))
	}

	rec, response = request(t, engine, http.MethodGet, "/api/v1/openings/facets?remote=true", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("facets: status %d", rec.Code)
	}
	facets := response.Data.(map[string]any)["remote"].(map[string]any)
	if facets["onsite"].(float64) != 1 {
		t.Fatalf("the on-site count must survive a remote-only query: %+v", facets)
	}

	rec, response = request(t, engine, http.MethodGet, "/api/v1/openings/stats", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("stats: status %d", rec.Code)
	}
	stats := response.Data.(map[string]any)
	if stats["total"].(float64) != 3 || stats["companies"].(float64) != 3 {
		t.Fatalf("stats = %+v", stats)
	}
	if len(stats["monthly"].([]any)) == 0 {
		t.Fatal("the monthly series should have at least the current month")
	}
}

// Creating the same posting twice through the API has to keep working: the
// unique index that ingestion relies on must not constrain what a person types.
func TestIdenticalOpeningsCanBeCreatedTwice(t *testing.T) {
	engine, _ := newAPI(t, nil)

	first, firstBody := request(t, engine, http.MethodPost, "/api/v1/openings", newOpening)
	second, secondBody := request(t, engine, http.MethodPost, "/api/v1/openings", newOpening)

	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("statuses %d and %d, want 201 twice (%s)", first.Code, second.Code, second.Body)
	}

	firstID := firstBody.Data.(map[string]any)["id"]
	secondID := secondBody.Data.(map[string]any)["id"]
	if firstID == secondID {
		t.Fatalf("both requests returned opening %v", firstID)
	}

	// Provenance is assigned by the server, not sent by the client.
	if source := secondBody.Data.(map[string]any)["source"]; source != "manual" {
		t.Fatalf("source = %v, want manual", source)
	}
}

// An empty index must answer with an empty array, never null: the frontend maps
// over this value.
func TestEmptyListIsAnArray(t *testing.T) {
	engine, _ := newAPI(t, nil)

	rec, _ := request(t, engine, http.MethodGet, "/api/v1/openings", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"data":[]`) {
		t.Fatalf("body = %s, want an empty array", rec.Body)
	}
}

func TestHealthProbes(t *testing.T) {
	engine, db := newAPI(t, nil)

	for _, path := range []string{"/healthz", "/readyz"} {
		if rec, _ := request(t, engine, http.MethodGet, path, ""); rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d, want 200", path, rec.Code)
		}
	}

	// With the database gone, liveness still passes — restarting the container
	// would not fix it — while readiness fails and traffic is withheld.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("reading sql.DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("closing database: %v", err)
	}

	if rec, _ := request(t, engine, http.MethodGet, "/healthz", ""); rec.Code != http.StatusOK {
		t.Fatalf("liveness with a dead database = %d, want 200", rec.Code)
	}
	if rec, _ := request(t, engine, http.MethodGet, "/readyz", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness with a dead database = %d, want 503", rec.Code)
	}
}

func TestUnknownAPIRouteAnswersJSON(t *testing.T) {
	engine, _ := newAPI(t, nil)

	rec, response := request(t, engine, http.MethodGet, "/api/v1/nope", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if response.Error == "" {
		t.Fatalf("a fetch() caller needs JSON, got %s", rec.Body)
	}
}

func TestRequestIDIsEchoedAndGenerated(t *testing.T) {
	engine, _ := newAPI(t, nil)

	rec, _ := request(t, engine, http.MethodGet, "/healthz", "")
	if rec.Header().Get(middleware.HeaderRequestID) == "" {
		t.Fatal("every response should carry a correlation id")
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(middleware.HeaderRequestID, "abc123")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Header().Get(middleware.HeaderRequestID); got != "abc123" {
		t.Fatalf("request id = %q, want the caller's own id", got)
	}
}

func TestCORS(t *testing.T) {
	engine, _ := newAPI(t, nil)

	preflight := httptest.NewRequest(http.MethodOptions, "/api/v1/openings", nil)
	preflight.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, preflight)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatal("the allowed origin should be echoed back")
	}

	blocked := httptest.NewRequest(http.MethodGet, "/api/v1/openings", nil)
	blocked.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, blocked)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("an unlisted origin must not be granted access")
	}
}

func TestSPAFallback(t *testing.T) {
	spa := &fstest.MapFS{
		"index.html":         {Data: []byte("<!doctype html><title>Goportunitties</title>")},
		"assets/app-a1b2.js": {Data: []byte("console.log('hi')")},
	}
	engine, _ := newAPI(t, spa)

	// A real file is served as itself.
	rec, _ := request(t, engine, http.MethodGet, "/assets/app-a1b2.js", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "console.log") {
		t.Fatalf("asset: status %d body %s", rec.Code, rec.Body)
	}
	if cache := rec.Header().Get("Cache-Control"); !strings.Contains(cache, "immutable") {
		t.Fatalf("fingerprinted assets should be cacheable forever, got %q", cache)
	}

	// A deep link is not a file, so the shell answers and the SPA routes it.
	rec, _ = request(t, engine, http.MethodGet, "/vagas/42", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<!doctype html>") {
		t.Fatalf("deep link: status %d body %s", rec.Code, rec.Body)
	}
	if cache := rec.Header().Get("Cache-Control"); cache != "no-cache" {
		t.Fatalf("the shell must not be cached, got %q", cache)
	}

	// API paths keep answering JSON even with a SPA mounted.
	rec, response := request(t, engine, http.MethodGet, "/api/v1/nope", "")
	if rec.Code != http.StatusNotFound || response.Error == "" {
		t.Fatalf("api 404 turned into HTML: status %d body %s", rec.Code, rec.Body)
	}
}
