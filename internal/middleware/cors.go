package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that allows browser clients from the given origins
// to call the API. Origins are configured via the CORS_ORIGINS env var.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAny := false
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			allowAny = true
		}
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")

		if origin != "" {
			_, ok := allowed[origin]
			if allowAny {
				ctx.Header("Access-Control-Allow-Origin", "*")
			} else if ok {
				ctx.Header("Access-Control-Allow-Origin", origin)
				// Tell caches the response varies per Origin.
				ctx.Header("Vary", "Origin")
			}
		}

		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Max-Age", "86400")

		// Preflight requests end here.
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
