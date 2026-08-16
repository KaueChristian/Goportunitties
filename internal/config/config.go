package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Settings holds every knob the application reads from the environment.
// Each field has a default, so the binary starts with no configuration at all.
type Settings struct {
	Env             string
	Port            string
	DBPath          string
	CORSOrigins     []string
	ShutdownTimeout time.Duration
	DefaultPageSize int
	MaxPageSize     int
	Ingestion       IngestionSettings

	// AdminKey gates every write endpoint (create, replace, patch, delete)
	// behind a shared secret sent as the X-Admin-Key header. Listing stays
	// public regardless — this only stops a visitor from creating, editing or
	// deleting openings, including the ones ingestion brought in.
	//
	// Empty is a valid value in development: it makes the gate a no-op, so
	// `go run ./cmd/api` keeps working with zero setup. cmd/api refuses to
	// start in production without one, so an empty key never reaches a
	// deployment silently.
	AdminKey string
}

// IngestionSettings configures the worker. The API binary ignores all of it.
type IngestionSettings struct {
	// Sources are the board slugs to read, in the order they are configured.
	Sources []string
	// Interval between runs. Zero means run once and exit, which is what a cron
	// job or a manual invocation wants.
	Interval time.Duration
	// Timeout bounds a single source's fetch.
	Timeout time.Duration
	// MaxPerSource caps how many openings one board contributes per run.
	MaxPerSource int
	// UserAgent identifies this project to the boards it reads.
	UserAgent string
	// GitHubToken raises the rate limit on the community boards that run on
	// GitHub Issues. Optional: they are readable without it.
	GitHubToken string
}

// IsProduction reports whether the app should behave as a deployed instance
// (JSON logs, Gin release mode, no verbose SQL).
func (s Settings) IsProduction() bool {
	return s.Env == "production"
}

// Load reads the environment once, at startup. Nothing else in the codebase
// touches os.Getenv, so the full configuration surface is visible here.
func Load() Settings {
	return Settings{
		Env:  getEnv("APP_ENV", "development"),
		Port: getEnv("PORT", "8080"),
		// Same directory name the container mounts at /data, so the path means
		// the same thing whether the binary runs locally or in Docker.
		DBPath: getEnv("DB_PATH", "./data/main.db"),
		// Defaults cover the Vite dev server, which runs on a different port than the API.
		CORSOrigins:     splitAndTrim(getEnv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		DefaultPageSize: getInt("DEFAULT_PAGE_SIZE", 12),
		MaxPageSize:     getInt("MAX_PAGE_SIZE", 100),
		Ingestion: IngestionSettings{
			// The Brazilian boards come first: they are the ones publishing
			// openings in this project's own market.
			//
			// RemoteOK is deliberately absent. Its free feed measured five
			// usable postings in a hundred, the rest being retail work and
			// error pages served as if they were jobs. The adapter is still
			// registered, so adding "remoteok" here turns it back on.
			Sources: splitAndTrim(getEnv("INGESTION_SOURCES", "backend-br,frontend-br,remotive")),
			// Zero is meaningful here — "run once and exit" — so it cannot use
			// getDuration, which treats zero as "unset, take the default".
			Interval:     getIntervalOrOnce("INGESTION_INTERVAL", 6*time.Hour),
			Timeout:      getDuration("INGESTION_TIMEOUT", 30*time.Second),
			MaxPerSource: getInt("INGESTION_MAX_PER_SOURCE", 100),
			UserAgent:    getEnv("INGESTION_USER_AGENT", ""),
			GitHubToken:  getEnv("INGESTION_GITHUB_TOKEN", ""),
		},
		AdminKey: getEnv("ADMIN_KEY", ""),
	}
}

// getIntervalOrOnce parses a schedule where an explicit zero is a valid answer.
func getIntervalOrOnce(key string, fallback time.Duration) time.Duration {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	if raw == "0" {
		return 0
	}

	value, err := time.ParseDuration(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, ""))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(getEnv(key, ""))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
