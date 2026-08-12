package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/gin-gonic/gin"
)

// HeaderAdminKey carries the shared secret RequireAdminKey checks.
const HeaderAdminKey = "X-Admin-Key"

// RequireAdminKey gates a route behind a shared secret.
//
// Listing was always meant to be public; writing was not. Before this existed,
// any visitor could POST, PATCH or DELETE against the API directly — including
// deleting every opening ingestion had brought in — since nothing enforced in
// the UI (hiding the edit button on an ingested opening) reached the server.
//
// An empty key makes this a no-op, which is what keeps local development
// working with zero setup. config.Settings documents why an empty key can
// never reach a production deployment.
//
// The comparison is constant-time so a wrong guess cannot be narrowed down by
// timing how long the check takes to fail.
func RequireAdminKey(key string) gin.HandlerFunc {
	if key == "" {
		return func(ctx *gin.Context) { ctx.Next() }
	}

	expected := []byte(key)
	return func(ctx *gin.Context) {
		provided := []byte(ctx.GetHeader(HeaderAdminKey))

		// ConstantTimeCompare itself handles a length mismatch by returning 0
		// without panicking, so no length check is needed ahead of it — adding
		// one would leak the key's length through timing anyway.
		if subtle.ConstantTimeCompare(provided, expected) != 1 {
			dto.SendError(ctx, http.StatusUnauthorized, "chave de administrador ausente ou inválida")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
