package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// TheoryUseCase encapsulates theory learning flows: document reading and quizzes.
type TheoryUseCase struct {
	docRepo        repository.TheoryDocumentRepository
	attemptRepo    repository.TheoryAttemptRepository
	contentRepo    repository.TheoryContentRepository
	questionRepo   repository.TheoryQuestionSetRepository
	attemptDetail  repository.TheoryAttemptDetailRepository
	paperRepo      repository.TheoryPaperRepository
	paperQuestionRepo repository.TheoryPaperQuestionRepository
}

func NewTheoryUseCase(
	docRepo repository.TheoryDocumentRepository,
	attemptRepo repository.TheoryAttemptRepository,
	contentRepo repository.TheoryContentRepository,
	questionRepo repository.TheoryQuestionSetRepository,
	attemptDetail repository.TheoryAttemptDetailRepository,
	paperRepo repository.TheoryPaperRepository,
	paperQuestionRepo repository.TheoryPaperQuestionRepository,
) *TheoryUseCase {
	return &TheoryUseCase{
		docRepo:       docRepo,
		attemptRepo:   attemptRepo,
		contentRepo:   contentRepo,
		questionRepo:  questionRepo,
		attemptDetail: attemptDetail,
		paperRepo:     paperRepo,
		paperQuestionRepo: paperQuestionRepo,
	}
}

// ListDocumentsRequest lists theory documents with simple filters.
type ListDocumentsRequest struct {
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// ListDocumentsItem represents one row in the list.
type ListDocumentsItem struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Category       string `json:"category,omitempty"`
	SortOrder      int    `json:"sort_order"`
	Status         string `json:"status,omitempty"`
	HasQuestions   bool   `json:"has_questions"`
	HasVideo       bool   `json:"has_video"`
	MyLatestScore  int    `json:"my_latest_score"`
	MyAttemptCount int    `json:"my_attempt_count"`
}

// ListDocumentsResponse is the response for listing theory documents.
type ListDocumentsResponse struct {
	Items []ListDocumentsItem `json:"items"`
	Total int                 `json:"total"`
}

// ListDocuments currently implements a very simple behaviour: it ignores keyword/category
// and just returns all documents ordered by sort_order. Filters and stats can be enriched later.
func (uc *TheoryUseCase) ListDocuments(ctx context.Context, userID string, req *ListDocumentsRequest) (*ListDocumentsResponse, error) {
	docs, err := uc.docRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]ListDocumentsItem, 0, len(docs))
	for _, d := range docs {
		items = append(items, ListDocumentsItem{
			ID:       d.ID,
			Title:    d.Title,
			SortOrder: d.SortOrder,
			HasVideo: d.VideoURL != "",
		})
	}
	return &ListDocumentsResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// DocumentDetailResponse aggregates document metadata and content blocks.
type DocumentDetailResponse struct {
	Document *domain.TheoryDocument          `json:"document"`
	Content  []repository.TheoryContentBlock `json:"content"`
}

func (uc *TheoryUseCase) GetDocumentDetail(ctx context.Context, documentID string) (*DocumentDetailResponse, error) {
	doc, err := uc.docRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	content, err := uc.contentRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	return &DocumentDetailResponse{
		Document: doc,
		Content:  content.Blocks,
	}, nil
}

// StartQuestionsResponse bundles attempt id and questions for client.
type StartQuestionsResponse struct {
	AttemptID        string                          `json:"attempt_id"`
	TimeLimitSeconds int                             `json:"time_limit_seconds"`
	Questions        []repository.TheoryQuestionItem `json:"questions"`
}

// StartQuestions can optionally take a paperID; when paperID is non-empty, questions
// will be drawn according to the paper configuration, otherwise all questions
// under the document will be used.
func (uc *TheoryUseCase) StartQuestions(ctx context.Context, userID, documentID, paperID string) (*StartQuestionsResponse, error) {
	// For now we re-use a single row to represent the attempt "shell".
	// The detailed answers and grading will be saved on submit.
	now := time.Now()
	attempt := &domain.TheoryAttempt{
		UserID:           userID,
		TheoryDocumentID: documentID,
		SubmittedAt:      now,
	}
	// Create initial attempt record with zero score.
	if err := uc.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, err
	}

	qs, err := uc.questionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	questions := qs.Questions
	// If a paper is specified, slice and reorder questions by paper config.
	if paperID != "" {
		paperQuestions, err := uc.paperQuestionRepo.ListByPaperID(ctx, paperID)
		if err != nil {
			return nil, err
		}
		// index question set by ID
		index := make(map[string]repository.TheoryQuestionItem, len(questions))
		for _, q := range questions {
			index[q.ID] = q
		}
		var selected []repository.TheoryQuestionItem
		for _, pq := range paperQuestions {
			q, ok := index[pq.QuestionID]
			if !ok {
				continue
			}
			q.Score = pq.Score
			selected = append(selected, q)
		}
		questions = selected
	}
	return &StartQuestionsResponse{
		AttemptID:        attempt.ID,
		TimeLimitSeconds: 1800,
		Questions:        questions,
	}, nil
}

// SubmitAnswersRequest describes the payload from client on submit.
type SubmitAnswersRequest struct {
	AttemptID string `json:"attempt_id"`
	Answers   []struct {
		QuestionID        string   `json:"question_id"`
		SelectedOptionIDs []string `json:"selected_option_ids"`
	} `json:"answers"`
}

// SubmitAnswersResponse is the grading result.
type SubmitAnswersResponse struct {
	Score        int  `json:"score"`
	MaxScore     int  `json:"max_score"`
	CorrectCount int  `json:"correct_count"`
	TotalCount   int  `json:"total_count"`
	Passed       bool `json:"passed"`
}

// SubmitAnswers grades and stores a completed attempt.
func (uc *TheoryUseCase) SubmitAnswers(ctx context.Context, userID, documentID string, req *SubmitAnswersRequest) (*SubmitAnswersResponse, error) {
	qs, err := uc.questionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// Build index from questionID -> question.
	qIndex := make(map[string]repository.TheoryQuestionItem, len(qs.Questions))
	for _, q := range qs.Questions {
		qIndex[q.ID] = q
	}

	// Grade answers.
	totalCount := len(req.Answers)
	var correctCount, score, maxScore int
	detail := &repository.TheoryAttemptDetailDoc{
		TheoryAttemptID: req.AttemptID,
		AutoGraded:      true,
		CreatedAt:       time.Now().Unix(),
	}
	for _, a := range req.Answers {
		q, ok := qIndex[a.QuestionID]
		if !ok {
			continue
		}
		maxScore += q.Score

		rec := repository.TheoryAnswerRecord{
			QuestionID:        a.QuestionID,
			SelectedOptionIDs: a.SelectedOptionIDs,
		}

		// Compare sets for correctness.
		if isSameStringSet(a.SelectedOptionIDs, q.CorrectOptionIDs) {
			rec.IsCorrect = true
			rec.Score = q.Score
			correctCount++
			score += q.Score
		}
		detail.Answers = append(detail.Answers, rec)
	}

	if err := uc.attemptDetail.Upsert(ctx, detail); err != nil {
		return nil, err
	}

	// Persist summary back to PostgreSQL.
	now := time.Now()
	answersJSONBytes, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, err
	}
	attempt := &domain.TheoryAttempt{
		ID:               req.AttemptID,
		UserID:           userID,
		TheoryDocumentID: documentID,
		AnswersJSON:      string(answersJSONBytes),
		Score:            score,
		SubmittedAt:      now,
		DurationSeconds:  0,
		MaxScore:         maxScore,
		CorrectCount:     correctCount,
		TotalCount:       totalCount,
		Passed:           score*100 >= maxScore*60, // 60% as default passing line
	}
	if err := uc.attemptRepo.UpdateOnSubmit(ctx, attempt); err != nil {
		return nil, err
	}

	return &SubmitAnswersResponse{
		Score:        score,
		MaxScore:     maxScore,
		CorrectCount: correctCount,
		TotalCount:   totalCount,
		Passed:       attempt.Passed,
	}, nil
}

// ListMyAttemptsResponse wraps user's past attempts summary.
type ListMyAttemptsResponse struct {
	Items []*domain.TheoryAttempt `json:"items"`
	Total int                     `json:"total"`
}

func (uc *TheoryUseCase) ListMyAttempts(ctx context.Context, userID, documentID string, page, pageSize int) (*ListMyAttemptsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	items, err := uc.attemptRepo.ListByUserAndDocument(ctx, userID, documentID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &ListMyAttemptsResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// isSameStringSet compares two string slices as sets (order-insensitive, no duplicates).
func isSameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		if m[v] == 0 {
			return false
		}
		m[v]--
		if m[v] == 0 {
			delete(m, v)
		}
	}
	return len(m) == 0
}

// ErrQuestionNotFound is returned when updating a non-existing question.
var ErrQuestionNotFound = errors.New("question not found")

// generateQuestionID returns a simple unique ID for new questions.
func generateQuestionID() string {
	return fmt.Sprintf("q_%d", time.Now().UnixNano())
}

// ----- Admin question management -----

// AdminListQuestionsRequest is the filter for listing questions under a document.
type AdminListQuestionsRequest struct {
	DocumentID string   `json:"document_id"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	Keyword    string   `json:"keyword"`
	Tags       []string `json:"tags"`
	Type       string   `json:"type"`
}

type AdminListQuestionsResponse struct {
	Items []repository.TheoryQuestionItem `json:"items"`
	Total int                             `json:"total"`
}

// AdminListQuestions reads question set from Mongo and returns basic list; filter/分页可以后续细化。
func (uc *TheoryUseCase) AdminListQuestions(ctx context.Context, req *AdminListQuestionsRequest) (*AdminListQuestionsResponse, error) {
	qs, err := uc.questionRepo.GetByDocumentID(ctx, req.DocumentID)
	if err != nil {
		return nil, err
	}
	items := qs.Questions
	// 简单忽略分页与过滤，后续可按需要在内存中过滤或在结构上扩展。
	return &AdminListQuestionsResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// AdminCreateQuestion appends a question into question set for a document.
func (uc *TheoryUseCase) AdminCreateQuestion(ctx context.Context, documentID string, q repository.TheoryQuestionItem) error {
	set, err := uc.questionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		// 如果文档不存在，可初始化一个新的集合
		set = &repository.TheoryQuestionSetDoc{
			DocumentID: documentID,
			Questions:  []repository.TheoryQuestionItem{},
		}
	}
	if q.ID == "" {
		q.ID = generateQuestionID()
	}
	set.Questions = append(set.Questions, q)
	return uc.questionRepo.Upsert(ctx, set)
}

// AdminUpdateQuestion replaces an existing question identified by ID.
func (uc *TheoryUseCase) AdminUpdateQuestion(ctx context.Context, documentID string, q repository.TheoryQuestionItem) error {
	set, err := uc.questionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return err
	}
	for i := range set.Questions {
		if set.Questions[i].ID == q.ID {
			set.Questions[i] = q
			return uc.questionRepo.Upsert(ctx, set)
		}
	}
	// 未找到时视为错误
	return ErrQuestionNotFound
}

// AdminDeleteQuestion removes a question from the set.
func (uc *TheoryUseCase) AdminDeleteQuestion(ctx context.Context, documentID, questionID string) error {
	set, err := uc.questionRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return err
	}
	out := make([]repository.TheoryQuestionItem, 0, len(set.Questions))
	for _, q := range set.Questions {
		if q.ID == questionID {
			continue
		}
		out = append(out, q)
	}
	set.Questions = out
	return uc.questionRepo.Upsert(ctx, set)
}

// ----- Paper management -----

type ListPapersRequest struct {
	DocumentID string `json:"document_id"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

type ListPapersItem struct {
	ID               string    `json:"id"`
	TheoryDocumentID string    `json:"theory_document_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	TotalScore       int       `json:"total_score"`
	Enabled          bool      `json:"enabled"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ListPapersResponse struct {
	Items []ListPapersItem `json:"items"`
	Total int              `json:"total"`
}

func (uc *TheoryUseCase) ListPapers(ctx context.Context, req *ListPapersRequest) (*ListPapersResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize
	papers, total, err := uc.paperRepo.ListByDocumentID(ctx, req.DocumentID, req.PageSize, offset)
	if err != nil {
		return nil, err
	}
	items := make([]ListPapersItem, 0, len(papers))
	for _, p := range papers {
		items = append(items, ListPapersItem{
			ID:               p.ID,
			TheoryDocumentID: p.TheoryDocumentID,
			Name:             p.Name,
			Description:      p.Description,
			TotalScore:       p.TotalScore,
			Enabled:          p.Enabled,
			CreatedBy:        p.CreatedBy,
			CreatedAt:        p.CreatedAt,
			UpdatedAt:        p.UpdatedAt,
		})
	}
	return &ListPapersResponse{
		Items: items,
		Total: total,
	}, nil
}

type SavePaperQuestion struct {
	QuestionID string `json:"question_id"`
	Score      int    `json:"score"`
	SortOrder  int    `json:"sort_order"`
}

type SavePaperRequest struct {
	ID               *string            `json:"id"`
	TheoryDocumentID string             `json:"theory_document_id" binding:"required"`
	Name             string             `json:"name" binding:"required"`
	Description      string             `json:"description"`
	Enabled          bool               `json:"enabled"`
	Questions        []SavePaperQuestion `json:"questions"`
}

func (uc *TheoryUseCase) SavePaper(ctx context.Context, userID string, req *SavePaperRequest) (*domain.TheoryPaper, error) {
	var totalScore int
	pq := make([]*domain.TheoryPaperQuestion, 0, len(req.Questions))
	for _, q := range req.Questions {
		totalScore += q.Score
		pq = append(pq, &domain.TheoryPaperQuestion{
			QuestionID: q.QuestionID,
			Score:      q.Score,
			SortOrder:  q.SortOrder,
		})
	}

	var paper *domain.TheoryPaper
	var err error
	if req.ID == nil || *req.ID == "" {
		// create
		p := &domain.TheoryPaper{
			TheoryDocumentID: req.TheoryDocumentID,
			Name:             req.Name,
			Description:      req.Description,
			TotalScore:       totalScore,
			Enabled:          req.Enabled,
			CreatedBy:        userID,
		}
		if err = uc.paperRepo.Create(ctx, p); err != nil {
			return nil, err
		}
		paper = p
	} else {
		// update
		p, err2 := uc.paperRepo.GetByID(ctx, *req.ID)
		if err2 != nil {
			return nil, err2
		}
		p.Name = req.Name
		p.Description = req.Description
		p.TotalScore = totalScore
		p.Enabled = req.Enabled
		if err = uc.paperRepo.Update(ctx, p); err != nil {
			return nil, err
		}
		paper = p
	}

	// replace questions for this paper
	if err := uc.paperQuestionRepo.ReplaceForPaper(ctx, paper.ID, pq); err != nil {
		return nil, err
	}
	return paper, nil
}

func (uc *TheoryUseCase) DeletePaper(ctx context.Context, paperID string) error {
	return uc.paperRepo.Delete(ctx, paperID)
}

// ----- Document management -----

type CreateDocumentRequest struct {
	Title     string `json:"title" binding:"required"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
	Category  string `json:"category"`
	VideoURL  string `json:"video_url"`
}

type UpdateDocumentRequest struct {
	ID        string `json:"id" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
	Category  string `json:"category"`
	VideoURL  string `json:"video_url"`
}

func (uc *TheoryUseCase) CreateDocument(ctx context.Context, req *CreateDocumentRequest) (*domain.TheoryDocument, error) {
	doc := &domain.TheoryDocument{
		Title:     req.Title,
		Slug:      req.Slug,
		SortOrder: req.SortOrder,
		Category:  req.Category,
		VideoURL:  req.VideoURL,
	}
	if err := uc.docRepo.Create(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (uc *TheoryUseCase) UpdateDocument(ctx context.Context, req *UpdateDocumentRequest) (*domain.TheoryDocument, error) {
	doc, err := uc.docRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	doc.Title = req.Title
	doc.Slug = req.Slug
	doc.SortOrder = req.SortOrder
	doc.Category = req.Category
	doc.VideoURL = req.VideoURL
	if err := uc.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (uc *TheoryUseCase) DeleteDocument(ctx context.Context, id string) error {
	return uc.docRepo.Delete(ctx, id)
}



