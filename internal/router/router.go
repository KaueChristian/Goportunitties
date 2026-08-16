// Package router assembles the HTTP engine: middleware, routes and, in a
// production build, the embedded frontend.
package router

import (
	"io/fs"
	"log/slog"

	"github.com/KaueChristian/Goportunitties/internal/config"
	"github.com/KaueChristian/Goportunitties/internal/handler"
	"github.com/KaueChristian/Goportunitties/internal/middleware"
	"github.com/KaueChristian/Goportunitties/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Deps is everything the router needs, handed to it by main. Nothing here is
// resolved through a package-level global, which is what lets a test build the
// same engine against an in-memory database.
type Deps struct {
	Settings config.Settings
	Log      *slog.Logger
	DB       *gorm.DB
	Service  service.OpeningService
	Version  string

	// SPA is the built frontend. When nil (development, where Vite serves the
	// app), the router only answers the API.
	SPA fs.FS
}

// New builds the HTTP handler.
func New(deps Deps) *gin.Engine {
	if deps.Settings.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.Logger(deps.Log),
		middleware.CORS(deps.Settings.CORSOrigins),
	)

	openings := handler.NewOpeningHandler(
		deps.Service,
		deps.Log,
		deps.Settings.DefaultPageSize,
		deps.Settings.MaxPageSize,
	)
	health := handler.NewHealthHandler(deps.DB, deps.Version)

	registerRoutes(engine, openings, health, deps.Settings.AdminKey)
	registerSPA(engine, deps.SPA)

	return engine
}
