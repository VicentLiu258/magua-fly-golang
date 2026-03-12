package handler

import (
	"net/http"
	"strconv"

	"magua-fly-golang/internal/repository"

	"github.com/gin-gonic/gin"
)

// StudentHandler handles student endpoints.
type StudentHandler struct {
	studentRepo repository.StudentRepository
}

// NewStudentHandler creates a StudentHandler.
func NewStudentHandler(studentRepo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{studentRepo: studentRepo}
}

// List returns a paginated list of students.
func (h *StudentHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	list, err := h.studentRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
