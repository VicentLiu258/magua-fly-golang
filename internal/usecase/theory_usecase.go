package usecase

import (
	"context"
	"encoding/json"
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
}

func NewTheoryUseCase(
	docRepo repository.TheoryDocumentRepository,
	attemptRepo repository.TheoryAttemptRepository,
	contentRepo repository.TheoryContentRepository,
	questionRepo repository.TheoryQuestionSetRepository,
	attemptDetail repository.TheoryAttemptDetailRepository,
) *TheoryUseCase {
	return &TheoryUseCase{
		docRepo:       docRepo,
		attemptRepo:   attemptRepo,
		contentRepo:   contentRepo,
		questionRepo:  questionRepo,
		attemptDetail: attemptDetail,
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
	AttemptID        string                              `json:"attempt_id"`
	TimeLimitSeconds int                                 `json:"time_limit_seconds"`
	Questions        []repository.TheoryQuestionItem     `json:"questions"`
}

func (uc *TheoryUseCase) StartQuestions(ctx context.Context, userID, documentID string) (*StartQuestionsResponse, error) {
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

	return &StartQuestionsResponse{
		AttemptID:        attempt.ID,
		TimeLimitSeconds: 1800,
		Questions:        qs.Questions,
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


