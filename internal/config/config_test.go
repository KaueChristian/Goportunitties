package config_test

import (
	"testing"
	"time"

	"github.com/KaueChristian/Goportunitties/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	settings := config.Load()

	if settings.Port != "8080" || settings.DBPath != "./data/main.db" {
		t.Fatalf("defaults = %+v", settings)
	}
	if settings.DefaultPageSize != 12 || settings.MaxPageSize != 100 {
		t.Fatalf("paging defaults = %d/%d, want 12/100", settings.DefaultPageSize, settings.MaxPageSize)
	}
	if settings.ShutdownTimeout != 10*time.Second {
		t.Fatalf("shutdown timeout = %s, want 10s", settings.ShutdownTimeout)
	}
	if settings.IsProduction() {
		t.Fatal("an unconfigured process must not think it is in production")
	}
	if len(settings.CORSOrigins) != 2 {
		t.Fatalf("CORS origins = %v, want the two Vite dev origins", settings.CORSOrigins)
	}
}

func TestLoadReadsTheEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("DB_PATH", "/data/app.db")
	t.Setenv("CORS_ORIGINS", " https://vagas.dev , https://www.vagas.dev ")
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")
	t.Setenv("DEFAULT_PAGE_SIZE", "24")

	settings := config.Load()

	if !settings.IsProduction() || settings.Port != "9090" || settings.DBPath != "/data/app.db" {
		t.Fatalf("settings = %+v", settings)
	}
	if settings.ShutdownTimeout != 30*time.Second || settings.DefaultPageSize != 24 {
		t.Fatalf("settings = %+v", settings)
	}
	// Whitespace around a comma-separated origin is a typo, not a distinct origin.
	if len(settings.CORSOrigins) != 2 || settings.CORSOrigins[0] != "https://vagas.dev" {
		t.Fatalf("CORS origins = %q", settings.CORSOrigins)
	}
}

func TestLoadFallsBackOnUnusableValues(t *testing.T) {
	t.Setenv("PORT", "   ")
	t.Setenv("SHUTDOWN_TIMEOUT", "muito tempo")
	t.Setenv("DEFAULT_PAGE_SIZE", "-5")
	t.Setenv("MAX_PAGE_SIZE", "abc")

	settings := config.Load()

	if settings.Port != "8080" {
		t.Fatalf("port = %q, want the default", settings.Port)
	}
	if settings.ShutdownTimeout != 10*time.Second {
		t.Fatalf("shutdown timeout = %s, want the default", settings.ShutdownTimeout)
	}
	if settings.DefaultPageSize != 12 || settings.MaxPageSize != 100 {
		t.Fatalf("paging = %d/%d, want the defaults", settings.DefaultPageSize, settings.MaxPageSize)
	}
}
