package handler

import (
	"net/http"

	"magua-fly-golang/internal/repository"

	"github.com/gin-gonic/gin"
)

// InstructorHandler handles instructor endpoints.
type InstructorHandler struct {
	instructorRepo repository.InstructorRepository
}

// NewInstructorHandler creates an InstructorHandler.
func NewInstructorHandler(instructorRepo repository.InstructorRepository) *InstructorHandler {
	return &InstructorHandler{instructorRepo: instructorRepo}
}

// List returns all instructors.
func (h *InstructorHandler) List(c *gin.Context) {
	list, err := h.instructorRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
