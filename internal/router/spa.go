package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/gin-gonic/gin"
)

// registerSPA serves the built frontend from the same binary as the API.
//
// Anything that isn't a real file falls back to index.html, so the SPA keeps
// working on a deep link or a reload — while /api paths still get a JSON 404
// instead of an HTML page, which is what a fetch() caller can actually handle.
func registerSPA(engine *gin.Engine, spa fs.FS) {
	if spa == nil {
		engine.NoRoute(notFound)
		return
	}

	fileServer := http.FileServer(http.FS(spa))

	engine.NoRoute(func(ctx *gin.Context) {
		if isAPIPath(ctx.Request.URL.Path) {
			notFound(ctx)
			return
		}

		name := strings.TrimPrefix(ctx.Request.URL.Path, "/")
		if info, err := fs.Stat(spa, name); err == nil && !info.IsDir() {
			// Vite fingerprints filenames, so a hit on /assets can be cached
			// forever: a new build produces a new name.
			if strings.HasPrefix(name, "assets/") {
				ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		index, err := fs.ReadFile(spa, "index.html")
		if err != nil {
			notFound(ctx)
			return
		}

		// The shell must never be cached, or a deploy leaves browsers asking for
		// asset filenames that no longer exist.
		ctx.Header("Cache-Control", "no-cache")
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api/") ||
		path == "/healthz" ||
		path == "/readyz"
}

func notFound(ctx *gin.Context) {
	dto.SendError(ctx, http.StatusNotFound, "rota não encontrada")
}
