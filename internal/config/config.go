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
	}
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
