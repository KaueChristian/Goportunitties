package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger replaces gin.Default()'s printf-style access log with a structured
// one: same information, but each field is queryable once the logs leave the
// terminal.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery

		ctx.Next()

		attrs := []slog.Attr{
			slog.String("method", ctx.Request.Method),
			slog.String("path", path),
			slog.Int("status", ctx.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
			slog.String("requestId", RequestIDFrom(ctx)),
		}
		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}
		if errs := ctx.Errors.ByType(gin.ErrorTypePrivate).String(); errs != "" {
			attrs = append(attrs, slog.String("errors", errs))
		}

		level := slog.LevelInfo
		switch {
		case ctx.Writer.Status() >= 500:
			level = slog.LevelError
		case ctx.Writer.Status() >= 400:
			level = slog.LevelWarn
		}

		log.LogAttrs(ctx.Request.Context(), level, "http request", attrs...)
	}
}
