package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// HeaderRequestID is the header carrying the correlation id, both inbound and
// outbound.
const HeaderRequestID = "X-Request-ID"

// contextKeyRequestID is where the id lives inside the Gin context.
const contextKeyRequestID = "requestID"

// RequestID gives every request a correlation id — reusing the caller's when it
// sends one — so a log line, a response and a client-side error report can all
// be tied to the same request.
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetHeader(HeaderRequestID)
		if id == "" {
			id = newRequestID()
		}

		ctx.Set(contextKeyRequestID, id)
		ctx.Header(HeaderRequestID, id)
		ctx.Next()
	}
}

// RequestIDFrom reads the id set by the RequestID middleware.
func RequestIDFrom(ctx *gin.Context) string {
	if id, ok := ctx.Get(contextKeyRequestID); ok {
		if value, ok := id.(string); ok {
			return value
		}
	}
	return ""
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		// An id is a debugging aid, never a reason to fail the request.
		return "unknown"
	}
	return hex.EncodeToString(buf)
}
