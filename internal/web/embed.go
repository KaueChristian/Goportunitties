// Package web embeds the built frontend into the binary.
//
// `npm run build` writes to internal/web/dist (see frontend/vite.config.ts), so
// `go build` produces a single artifact that serves both the API and the SPA —
// no static file server, no second deployment target.
package web

import (
	"embed"
	"io/fs"
)

// The directory is committed with only a .gitkeep so the embed also compiles on
// a checkout where the frontend was never built. `all:` is what includes that
// dot-file, which is what keeps the pattern from failing on an empty build.
//
//go:embed all:dist
var dist embed.FS

// SPA returns the built frontend rooted at dist/, or nil when this binary was
// compiled without one — in which case the router serves the API only.
func SPA() fs.FS {
	root, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil
	}

	if _, err := fs.Stat(root, "index.html"); err != nil {
		return nil
	}
	return root
}
