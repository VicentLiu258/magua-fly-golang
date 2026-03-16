package handler

import (
	"net/http"

	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/repository"
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
	PaperID    string `json:"paper_id"`
}

func (h *TheoryHandler) StartQuestions(c *gin.Context) {
	var req startQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	resp, err := h.uc.StartQuestions(c.Request.Context(), userID, req.DocumentID, req.PaperID)
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

// ----- Document management -----

type createDocumentBody struct {
	Title     string `json:"title" binding:"required"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
	Category  string `json:"category"`
	VideoURL  string `json:"video_url"`
}

func (h *TheoryHandler) CreateDocument(c *gin.Context) {
	var body createDocumentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	doc, err := h.uc.CreateDocument(c.Request.Context(), &usecase.CreateDocumentRequest{
		Title:     body.Title,
		Slug:      body.Slug,
		SortOrder: body.SortOrder,
		Category:  body.Category,
		VideoURL:  body.VideoURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(doc))
}

type updateDocumentBody struct {
	ID        string `json:"id" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
	Category  string `json:"category"`
	VideoURL  string `json:"video_url"`
}

func (h *TheoryHandler) UpdateDocument(c *gin.Context) {
	var body updateDocumentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	doc, err := h.uc.UpdateDocument(c.Request.Context(), &usecase.UpdateDocumentRequest{
		ID:        body.ID,
		Title:     body.Title,
		Slug:      body.Slug,
		SortOrder: body.SortOrder,
		Category:  body.Category,
		VideoURL:  body.VideoURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(doc))
}

type deleteDocumentBody struct {
	ID string `json:"id" binding:"required"`
}

func (h *TheoryHandler) DeleteDocument(c *gin.Context) {
	var body deleteDocumentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := h.uc.DeleteDocument(c.Request.Context(), body.ID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

// ----- Admin question management -----

func (h *TheoryHandler) AdminListQuestions(c *gin.Context) {
	var req usecase.AdminListQuestionsRequest
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

type adminQuestionBody struct {
	DocumentID string                         `json:"document_id" binding:"required"`
	Question   repository.TheoryQuestionItem `json:"question" binding:"required"`
}

func (h *TheoryHandler) AdminCreateQuestion(c *gin.Context) {
	var body adminQuestionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := h.uc.AdminCreateQuestion(c.Request.Context(), body.DocumentID, body.Question); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

func (h *TheoryHandler) AdminUpdateQuestion(c *gin.Context) {
	var body adminQuestionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := h.uc.AdminUpdateQuestion(c.Request.Context(), body.DocumentID, body.Question); err != nil {
		if err == usecase.ErrQuestionNotFound {
			c.JSON(http.StatusBadRequest, errorMessage("question not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

type adminDeleteQuestionBody struct {
	DocumentID string `json:"document_id" binding:"required"`
	QuestionID string `json:"question_id" binding:"required"`
}

func (h *TheoryHandler) AdminDeleteQuestion(c *gin.Context) {
	var body adminDeleteQuestionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := h.uc.AdminDeleteQuestion(c.Request.Context(), body.DocumentID, body.QuestionID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

// ----- Paper management -----

type listPapersRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

func (h *TheoryHandler) ListPapers(c *gin.Context) {
	var req listPapersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	resp, err := h.uc.ListPapers(c.Request.Context(), &usecase.ListPapersRequest{
		DocumentID: req.DocumentID,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(resp))
}

type savePaperBody struct {
	ID               *string                    `json:"id"`
	TheoryDocumentID string                     `json:"theory_document_id" binding:"required"`
	Name             string                     `json:"name" binding:"required"`
	Description      string                     `json:"description"`
	Enabled          bool                       `json:"enabled"`
	Questions        []usecase.SavePaperQuestion `json:"questions"`
}

func (h *TheoryHandler) SavePaper(c *gin.Context) {
	var body savePaperBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	userID := getUserIDFromContext(c)
	paper, err := h.uc.SavePaper(c.Request.Context(), userID, &usecase.SavePaperRequest{
		ID:               body.ID,
		TheoryDocumentID: body.TheoryDocumentID,
		Name:             body.Name,
		Description:      body.Description,
		Enabled:          body.Enabled,
		Questions:        body.Questions,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(paper))
}

type deletePaperBody struct {
	PaperID string `json:"paper_id" binding:"required"`
}

func (h *TheoryHandler) DeletePaper(c *gin.Context) {
	var body deletePaperBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := h.uc.DeletePaper(c.Request.Context(), body.PaperID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

