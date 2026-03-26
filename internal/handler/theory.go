package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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

// UploadTheoryVideo POST multipart: document_id + file
func (h *TheoryHandler) UploadTheoryVideo(c *gin.Context) {
	if getUserIDFromContext(c) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	documentID := c.PostForm("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, errorMessage("document_id required"))
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := h.uc.AttachTheoryVideo(c.Request.Context(), documentID, header.Filename, contentType, file); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

type removeTheoryVideoBody struct {
	DocumentID string `json:"document_id" binding:"required"`
}

func (h *TheoryHandler) RemoveTheoryVideo(c *gin.Context) {
	if getUserIDFromContext(c) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	var body removeTheoryVideoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := h.uc.RemoveTheoryVideo(c.Request.Context(), body.DocumentID); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, okResponse(gin.H{"status": "ok"}))
}

// parseBytesRange 解析单个 Range: bytes=…；partial=false 表示返回整文件。
func parseBytesRange(rangeHeader string, size int64) (start, end int64, partial bool, bad bool) {
	if size <= 0 {
		return 0, 0, false, true
	}
	if rangeHeader == "" {
		return 0, size - 1, false, false
	}
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, false, false
	}
	spec := strings.TrimSpace(strings.TrimPrefix(rangeHeader, "bytes="))
	dash := strings.IndexByte(spec, '-')
	if dash < 0 {
		return 0, 0, false, true
	}
	left, right := spec[:dash], spec[dash+1:]

	if left == "" && right != "" {
		n, err := strconv.ParseInt(right, 10, 64)
		if err != nil || n <= 0 || n > size {
			return 0, 0, false, true
		}
		start = size - n
		end = size - 1
		return start, end, true, false
	}
	if left != "" && right == "" {
		start, err := strconv.ParseInt(left, 10, 64)
		if err != nil || start < 0 || start >= size {
			return 0, 0, false, true
		}
		return start, size - 1, true, false
	}
	if left != "" && right != "" {
		start, err1 := strconv.ParseInt(left, 10, 64)
		end, err2 := strconv.ParseInt(right, 10, 64)
		if err1 != nil || err2 != nil || start < 0 || start > end {
			return 0, 0, false, true
		}
		if start >= size {
			return 0, 0, false, true
		}
		if end >= size {
			end = size - 1
		}
		return start, end, true, false
	}
	return 0, 0, false, true
}

func theoryVideoStreamError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, usecase.ErrTheoryNoHostedVideo):
		c.JSON(http.StatusNotFound, errorMessage("no hosted video"))
		return true
	case errors.Is(err, usecase.ErrTheoryVideoCOSNotReady):
		c.JSON(http.StatusNotImplemented, errorMessage("COS video not implemented"))
		return true
	case errors.Is(err, usecase.ErrTheoryVideoFileMissing):
		c.JSON(http.StatusNotFound, errorMessage("视频文件在存储中不存在，可能 MongoDB GridFS 与库表不一致，请重新上传"))
		return true
	default:
		return false
	}
}

// StreamTheoryVideo GET，支持 Accept-Ranges / Content-Range（GridFS 托管视频）。
func (h *TheoryHandler) StreamTheoryVideo(c *gin.Context) {
	if getUserIDFromContext(c) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	documentID := c.Query("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, errorMessage("document_id required"))
		return
	}
	ctx := c.Request.Context()
	total, ct, err := h.uc.GetTheoryVideoMeta(ctx, documentID)
	if err != nil {
		if theoryVideoStreamError(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	start, end, partial, bad := parseBytesRange(c.GetHeader("Range"), total)
	if bad {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", total))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", ct)

	if !partial {
		rc, _, _, err := h.uc.OpenTheoryVideoContent(ctx, documentID, 0, 0)
		if err != nil {
			if theoryVideoStreamError(c, err) {
				return
			}
			c.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		defer rc.Close()
		c.Header("Content-Length", strconv.FormatInt(total, 10))
		c.Status(http.StatusOK)
		_, _ = io.Copy(c.Writer, rc)
		return
	}

	length := end - start + 1
	rc, _, _, err := h.uc.OpenTheoryVideoContent(ctx, documentID, start, length)
	if err != nil {
		if theoryVideoStreamError(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	defer rc.Close()
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, total))
	c.Header("Content-Length", strconv.FormatInt(length, 10))
	c.Status(http.StatusPartialContent)
	_, _ = io.Copy(c.Writer, rc)
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

// adminQuestionOptionBody is the payload schema for a single option in admin question APIs.
type adminQuestionOptionBody struct {
	ID   string `json:"id" binding:"required"`
	Text string `json:"text" binding:"required"`
}

// adminQuestionPayload is the payload schema for a question in admin create/update APIs.
// It mirrors repository.TheoryQuestionItem but keeps correct answers/explanation visible
// to admin clients while still preventing them from being exposed in student APIs.
type adminQuestionPayload struct {
	ID               string                    `json:"id"`
	Type             string                    `json:"type" binding:"required"` // single | multi
	Stem             string                    `json:"stem" binding:"required"`
	Options          []adminQuestionOptionBody `json:"options" binding:"required"`
	CorrectOptionIDs []string                  `json:"correct_option_ids"` // admin only
	Explanation      string                    `json:"explanation"`
	Tags             []string                  `json:"tags"`
	Score            int                       `json:"score"`
}

// toRepositoryQuestion converts adminQuestionPayload into repository.TheoryQuestionItem
// so that sensitive fields (correct answers, explanation) are stored in MongoDB
// but not leaked to student-facing JSON responses.
func toRepositoryQuestion(q adminQuestionPayload) repository.TheoryQuestionItem {
	opts := make([]repository.TheoryQuestionOption, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, repository.TheoryQuestionOption{
			ID:   o.ID,
			Text: o.Text,
		})
	}
	return repository.TheoryQuestionItem{
		ID:               q.ID,
		Type:             q.Type,
		Stem:             q.Stem,
		Options:          opts,
		CorrectOptionIDs: q.CorrectOptionIDs,
		Explanation:      q.Explanation,
		Tags:             q.Tags,
		Score:            q.Score,
	}
}

type adminQuestionBody struct {
	DocumentID string               `json:"document_id" binding:"required"`
	Question   adminQuestionPayload `json:"question" binding:"required"`
}

func (h *TheoryHandler) AdminCreateQuestion(c *gin.Context) {
	var body adminQuestionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	q := toRepositoryQuestion(body.Question)
	if err := h.uc.AdminCreateQuestion(c.Request.Context(), body.DocumentID, q); err != nil {
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
	q := toRepositoryQuestion(body.Question)
	if err := h.uc.AdminUpdateQuestion(c.Request.Context(), body.DocumentID, q); err != nil {
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

