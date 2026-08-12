package router

import (
	"github.com/KaueChristian/Goportunitties/internal/handler"
	"github.com/gin-gonic/gin"
)

// registerRoutes wires every endpoint the API exposes.
//
// The collection is /openings and a single opening is /openings/:id — the
// aggregate views (facets, stats) are sub-resources of the collection, which is
// why they sit under it rather than at the root.
func registerRoutes(engine *gin.Engine, openings *handler.OpeningHandler, health *handler.HealthHandler) {
	engine.GET("/healthz", health.Live)
	engine.GET("/readyz", health.Ready)

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/openings", openings.List)
		v1.GET("/openings/facets", openings.Facets)
		v1.GET("/openings/stats", openings.Stats)
		v1.GET("/openings/suggestions", openings.Suggest)
		v1.POST("/openings", openings.Create)

		v1.GET("/openings/:id", openings.Show)
		v1.PUT("/openings/:id", openings.Replace)
		v1.PATCH("/openings/:id", openings.Patch)
		v1.DELETE("/openings/:id", openings.Delete)
	}
}
