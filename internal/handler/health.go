package handler

import (
	"net/http"

	"magua-fly-golang/internal/mongodb"
	"magua-fly-golang/internal/postgres"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health checks.
type HealthHandler struct {
	pg *postgres.DB
	m  *mongodb.Client
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(pg *postgres.DB, m *mongodb.Client) *HealthHandler {
	return &HealthHandler{pg: pg, m: m}
}

// Check returns 200 if both DBs are reachable.
func (h *HealthHandler) Check(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.pg.Health(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "postgres unhealthy", "detail": err.Error()})
		return
	}
	if err := h.m.Health(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "mongodb unhealthy", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
