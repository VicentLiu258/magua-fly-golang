package handler

import (
	"net/http"

	"magua-fly-golang/internal/repository"

	"github.com/gin-gonic/gin"
)

// AircraftHandler handles aircraft endpoints.
type AircraftHandler struct {
	aircraftRepo repository.AircraftRepository
}

// NewAircraftHandler creates an AircraftHandler.
func NewAircraftHandler(aircraftRepo repository.AircraftRepository) *AircraftHandler {
	return &AircraftHandler{aircraftRepo: aircraftRepo}
}

// List returns all aircraft.
func (h *AircraftHandler) List(c *gin.Context) {
	list, err := h.aircraftRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{"items": list},
	})
}
