package handler

import (
	"net/http"

	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SubjectExamHandler exposes training-subject quiz APIs.
type SubjectExamHandler struct {
	uc *usecase.SubjectExamUseCase
}

// NewSubjectExamHandler constructs SubjectExamHandler.
func NewSubjectExamHandler(uc *usecase.SubjectExamUseCase) *SubjectExamHandler {
	return &SubjectExamHandler{uc: uc}
}

type subjectStartQuestionsRequest struct {
	SubjectLearningID string `json:"subject_learning_id" binding:"required"`
}

// StartQuestions POST /api/v1/subject-learning/questions/start
func (h *SubjectExamHandler) StartQuestions(c *gin.Context) {
	var req subjectStartQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.StartQuestions(c.Request.Context(), userID, req.SubjectLearningID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

// SubmitAnswers POST /api/v1/subject-learning/questions/submit
func (h *SubjectExamHandler) SubmitAnswers(c *gin.Context) {
	var body struct {
		SubjectLearningID string `json:"subject_learning_id" binding:"required"`
		usecase.SubjectSubmitAnswersRequest
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.SubmitAnswers(c.Request.Context(), userID, body.SubjectLearningID, &body.SubjectSubmitAnswersRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

type subjectListAttemptsRequest struct {
	SubjectLearningID string `json:"subject_learning_id" binding:"required"`
	Page              int    `json:"page"`
	PageSize          int    `json:"page_size"`
}

// ListMyAttempts POST /api/v1/subject-learning/attempts/my-list
func (h *SubjectExamHandler) ListMyAttempts(c *gin.Context) {
	var req subjectListAttemptsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.ListMyAttempts(c.Request.Context(), userID, req.SubjectLearningID, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

// AdminListQuestions POST /api/v1/subject-learning/questions/admin/list
func (h *SubjectExamHandler) AdminListQuestions(c *gin.Context) {
	var req usecase.AdminListSubjectQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	resp, err := h.uc.AdminListQuestions(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

type subjectAdminQuestionBody struct {
	SubjectLearningID string               `json:"subject_learning_id" binding:"required"`
	Question          adminQuestionPayload `json:"question" binding:"required"`
}

// AdminCreateQuestion POST /api/v1/subject-learning/questions/admin/create
func (h *SubjectExamHandler) AdminCreateQuestion(c *gin.Context) {
	var body subjectAdminQuestionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	q := toRepositoryQuestion(body.Question)
	if err := h.uc.AdminCreateQuestion(c.Request.Context(), body.SubjectLearningID, q); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

// AdminUpdateQuestion POST /api/v1/subject-learning/questions/admin/update
func (h *SubjectExamHandler) AdminUpdateQuestion(c *gin.Context) {
	var body subjectAdminQuestionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	q := toRepositoryQuestion(body.Question)
	if err := h.uc.AdminUpdateQuestion(c.Request.Context(), body.SubjectLearningID, q); err != nil {
		if err == usecase.ErrQuestionNotFound {
			c.JSON(http.StatusBadRequest, errorMessage("question not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

type subjectAdminDeleteQuestionBody struct {
	SubjectLearningID string `json:"subject_learning_id" binding:"required"`
	QuestionID        string `json:"question_id" binding:"required"`
}

// AdminDeleteQuestion POST /api/v1/subject-learning/questions/admin/delete
func (h *SubjectExamHandler) AdminDeleteQuestion(c *gin.Context) {
	var body subjectAdminDeleteQuestionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := h.uc.AdminDeleteQuestion(c.Request.Context(), body.SubjectLearningID, body.QuestionID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}
