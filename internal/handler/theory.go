package handler

import (
	"net/http"

	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// TheoryHandler holds handlers for theory learning APIs.
type TheoryHandler struct {
	uc *usecase.TheoryUseCase
}

func NewTheoryHandler(uc *usecase.TheoryUseCase) *TheoryHandler {
	return &TheoryHandler{uc: uc}
}

func okResponse(data interface{}) gin.H {
	return gin.H{"code": 200, "msg": "ok", "data": data}
}

func errorResponse(err error) gin.H {
	return gin.H{"code": 500, "msg": err.Error(), "data": nil}
}

func errorMessage(msg string) gin.H {
	return gin.H{"code": 400, "msg": msg, "data": nil}
}

func getUserIDFromContext(c *gin.Context) string {
	claims := middleware.GetGinClaims(c)
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// listDocumentsRequest mirrors usecase.ListDocumentsRequest for binding.
type listDocumentsRequest struct {
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

func (h *TheoryHandler) ListDocuments(c *gin.Context) {
	var req listDocumentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.ListDocuments(c.Request.Context(), userID, &usecase.ListDocumentsRequest{
		Keyword:  req.Keyword,
		Category: req.Category,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

type documentDetailRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
}

func (h *TheoryHandler) DocumentDetail(c *gin.Context) {
	var req documentDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	resp, err := h.uc.GetDocumentDetail(c.Request.Context(), req.DocumentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

type startQuestionsRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
}

func (h *TheoryHandler) StartQuestions(c *gin.Context) {
	var req startQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.StartQuestions(c.Request.Context(), userID, req.DocumentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

func (h *TheoryHandler) SubmitAnswers(c *gin.Context) {
	var body struct {
		DocumentID string                         `json:"document_id" binding:"required"`
		usecase.SubmitAnswersRequest
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.SubmitAnswers(c.Request.Context(), userID, body.DocumentID, &body.SubmitAnswersRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

type listAttemptsRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

func (h *TheoryHandler) ListMyAttempts(c *gin.Context) {
	var req listAttemptsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.ListMyAttempts(c.Request.Context(), userID, req.DocumentID, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

