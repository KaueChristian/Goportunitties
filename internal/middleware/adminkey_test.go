package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KaueChristian/Goportunitties/internal/middleware"
	"github.com/gin-gonic/gin"
)

func newGuardedEngine(key string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.POST("/protected", middleware.RequireAdminKey(key), func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	return engine
}

func do(engine *gin.Engine, headerValue string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	if headerValue != "" {
		req.Header.Set(middleware.HeaderAdminKey, headerValue)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestRequireAdminKeyIsANoOpWhenUnconfigured(t *testing.T) {
	engine := newGuardedEngine("")

	if rec := do(engine, ""); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — an unset key must not lock anything", rec.Code)
	}
}

func TestRequireAdminKeyRejectsAMissingHeader(t *testing.T) {
	engine := newGuardedEngine("s3cr3t")

	rec := do(engine, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAdminKeyRejectsTheWrongKey(t *testing.T) {
	engine := newGuardedEngine("s3cr3t")

	rec := do(engine, "not-it")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

// A key that merely starts with the right one must not pass — the comparison
// has to check the whole value, not a prefix.
func TestRequireAdminKeyRejectsAPrefixOfTheRealKey(t *testing.T) {
	engine := newGuardedEngine("s3cr3t")

	rec := do(engine, "s3cr3")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAdminKeyAcceptsTheRightKey(t *testing.T) {
	engine := newGuardedEngine("s3cr3t")

	rec := do(engine, "s3cr3t")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

// The handler behind the gate must never run on a rejected request — this is
// what stops a bad key from reaching the database at all.
func TestRequireAdminKeyNeverCallsTheNextHandlerOnRejection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false

	engine := gin.New()
	engine.POST("/protected", middleware.RequireAdminKey("s3cr3t"), func(ctx *gin.Context) {
		called = true
		ctx.Status(http.StatusOK)
	})

	do(engine, "wrong")

	if called {
		t.Fatal("the protected handler ran despite the wrong key")
	}
}
