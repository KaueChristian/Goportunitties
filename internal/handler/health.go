package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler answers the probes an orchestrator uses to decide whether the
// container is alive and whether it should receive traffic.
type HealthHandler struct {
	db      *gorm.DB
	version string
}

// NewHealthHandler builds a HealthHandler.
func NewHealthHandler(db *gorm.DB, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

// Live handles GET /healthz: the process is up. It touches no dependency on
// purpose — a failing database must not get the container restarted.
func (h *HealthHandler) Live(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "version": h.version})
}

// Ready handles GET /readyz: the process can actually serve requests, which
// here means the database answers.
func (h *HealthHandler) Ready(ctx *gin.Context) {
	sqlDB, err := h.db.DB()
	if err == nil {
		err = sqlDB.PingContext(ctx.Request.Context())
	}
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "database unreachable"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "version": h.version})
}
