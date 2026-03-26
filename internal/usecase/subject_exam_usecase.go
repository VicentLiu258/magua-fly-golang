package usecase

import (
	"context"
	"encoding/json"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/exam"
	"magua-fly-golang/internal/repository"
)

// SubjectExamUseCase handles training-subject quizzes (Mongo question sets + PG attempts).
type SubjectExamUseCase struct {
	subjectLearning repository.SubjectLearningRepository
	attemptRepo     repository.SubjectAttemptRepository
	questionRepo    repository.SubjectQuestionSetRepository
	attemptDetail   repository.SubjectAttemptDetailRepository
}

// NewSubjectExamUseCase wires subject exam dependencies.
func NewSubjectExamUseCase(
	subjectLearning repository.SubjectLearningRepository,
	attemptRepo repository.SubjectAttemptRepository,
	questionRepo repository.SubjectQuestionSetRepository,
	attemptDetail repository.SubjectAttemptDetailRepository,
) *SubjectExamUseCase {
	return &SubjectExamUseCase{
		subjectLearning: subjectLearning,
		attemptRepo:     attemptRepo,
		questionRepo:    questionRepo,
		attemptDetail:   attemptDetail,
	}
}

// SubjectStartQuestionsResponse mirrors theory start payload.
type SubjectStartQuestionsResponse struct {
	AttemptID        string                 `json:"attempt_id"`
	TimeLimitSeconds int                    `json:"time_limit_seconds"`
	Questions        []repository.TheoryQuestionItem `json:"questions"`
}

// StartQuestions creates an attempt shell and returns all questions for the subject learning unit.
func (uc *SubjectExamUseCase) StartQuestions(ctx context.Context, userID, subjectLearningID string) (*SubjectStartQuestionsResponse, error) {
	if _, err := uc.subjectLearning.GetByID(ctx, subjectLearningID); err != nil {
		return nil, err
	}
	now := time.Now()
	attempt := &domain.SubjectAttempt{
		UserID:            userID,
		SubjectLearningID: subjectLearningID,
		SubmittedAt:       now,
	}
	if err := uc.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, err
	}
	qs, err := uc.questionRepo.GetBySubjectLearningID(ctx, subjectLearningID)
	if err != nil {
		return nil, err
	}
	return &SubjectStartQuestionsResponse{
		AttemptID:        attempt.ID,
		TimeLimitSeconds: 1800,
		Questions:        qs.Questions,
	}, nil
}

// SubjectSubmitAnswersRequest matches theory submit body (without document_id; use subject_learning_id in handler).
type SubjectSubmitAnswersRequest struct {
	AttemptID string `json:"attempt_id"`
	Answers   []struct {
		QuestionID        string   `json:"question_id"`
		SelectedOptionIDs []string `json:"selected_option_ids"`
	} `json:"answers"`
}

// SubjectSubmitAnswersResponse is the grading result.
type SubjectSubmitAnswersResponse struct {
	Score        int  `json:"score"`
	MaxScore     int  `json:"max_score"`
	CorrectCount int  `json:"correct_count"`
	TotalCount   int  `json:"total_count"`
	Passed       bool `json:"passed"`
}

// SubmitAnswers grades and persists a subject attempt.
func (uc *SubjectExamUseCase) SubmitAnswers(ctx context.Context, userID, subjectLearningID string, req *SubjectSubmitAnswersRequest) (*SubjectSubmitAnswersResponse, error) {
	if _, err := uc.subjectLearning.GetByID(ctx, subjectLearningID); err != nil {
		return nil, err
	}
	qs, err := uc.questionRepo.GetBySubjectLearningID(ctx, subjectLearningID)
	if err != nil {
		return nil, err
	}
	qIndex := make(map[string]exam.ExamQuestionItem, len(qs.Questions))
	for _, q := range qs.Questions {
		qIndex[q.ID] = q
	}
	inputs := make([]exam.AnswerInput, 0, len(req.Answers))
	for _, a := range req.Answers {
		inputs = append(inputs, exam.AnswerInput{
			QuestionID:        a.QuestionID,
			SelectedOptionIDs: a.SelectedOptionIDs,
		})
	}
	graded := exam.GradeMCQ(qIndex, inputs)

	detail := &repository.SubjectAttemptDetailDoc{
		SubjectAttemptID: req.AttemptID,
		Answers:          graded.Answers,
		AutoGraded:       true,
		CreatedAt:        time.Now().Unix(),
	}
	if err := uc.attemptDetail.Upsert(ctx, detail); err != nil {
		return nil, err
	}

	now := time.Now()
	answersJSONBytes, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, err
	}
	attempt := &domain.SubjectAttempt{
		ID:                req.AttemptID,
		UserID:            userID,
		SubjectLearningID: subjectLearningID,
		AnswersJSON:       string(answersJSONBytes),
		Score:             graded.Score,
		SubmittedAt:       now,
		MaxScore:          graded.MaxScore,
		CorrectCount:      graded.CorrectCount,
		TotalCount:        graded.TotalCount,
		Passed:            graded.Score*100 >= graded.MaxScore*60,
	}
	if err := uc.attemptRepo.UpdateOnSubmit(ctx, attempt); err != nil {
		return nil, err
	}
	return &SubjectSubmitAnswersResponse{
		Score:        graded.Score,
		MaxScore:     graded.MaxScore,
		CorrectCount: graded.CorrectCount,
		TotalCount:   graded.TotalCount,
		Passed:       attempt.Passed,
	}, nil
}

// ListMySubjectAttemptsResponse wraps past attempts.
type ListMySubjectAttemptsResponse struct {
	Items []*domain.SubjectAttempt `json:"items"`
	Total int                      `json:"total"`
}

// ListMyAttempts lists the user's attempts for a subject learning unit.
func (uc *SubjectExamUseCase) ListMyAttempts(ctx context.Context, userID, subjectLearningID string, page, pageSize int) (*ListMySubjectAttemptsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	items, err := uc.attemptRepo.ListByUserAndSubjectLearning(ctx, userID, subjectLearningID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &ListMySubjectAttemptsResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// ----- Admin question management (Mongo question set per subject_learning_id) -----

// AdminListSubjectQuestionsRequest filters listing under a subject learning id.
type AdminListSubjectQuestionsRequest struct {
	SubjectLearningID string `json:"subject_learning_id"`
	Page              int    `json:"page"`
	PageSize          int    `json:"page_size"`
}

// AdminListSubjectQuestionsResponse lists questions from Mongo.
type AdminListSubjectQuestionsResponse struct {
	Items []exam.AdminListExamQuestionItem `json:"items"`
	Total int                              `json:"total"`
}

// AdminListQuestions returns all questions in the set (filters deferred).
func (uc *SubjectExamUseCase) AdminListQuestions(ctx context.Context, req *AdminListSubjectQuestionsRequest) (*AdminListSubjectQuestionsResponse, error) {
	qs, err := uc.questionRepo.GetBySubjectLearningID(ctx, req.SubjectLearningID)
	if err != nil {
		return nil, err
	}
	items := make([]exam.AdminListExamQuestionItem, 0, len(qs.Questions))
	for _, q := range qs.Questions {
		items = append(items, exam.ToAdminListItem(q))
	}
	return &AdminListSubjectQuestionsResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// AdminCreateQuestion appends a question to the subject set.
func (uc *SubjectExamUseCase) AdminCreateQuestion(ctx context.Context, subjectLearningID string, q repository.TheoryQuestionItem) error {
	set, err := uc.questionRepo.GetBySubjectLearningID(ctx, subjectLearningID)
	if err != nil {
		set = &repository.SubjectQuestionSetDoc{
			SubjectLearningID: subjectLearningID,
			Questions:         []exam.ExamQuestionItem{},
		}
	}
	if q.ID == "" {
		q.ID = exam.GenerateQuestionID()
	}
	set.Questions = append(set.Questions, q)
	return uc.questionRepo.Upsert(ctx, set)
}

// AdminUpdateQuestion replaces a question by id.
func (uc *SubjectExamUseCase) AdminUpdateQuestion(ctx context.Context, subjectLearningID string, q repository.TheoryQuestionItem) error {
	set, err := uc.questionRepo.GetBySubjectLearningID(ctx, subjectLearningID)
	if err != nil {
		return err
	}
	for i := range set.Questions {
		if set.Questions[i].ID == q.ID {
			set.Questions[i] = q
			return uc.questionRepo.Upsert(ctx, set)
		}
	}
	return ErrQuestionNotFound
}

// AdminDeleteQuestion removes a question by id.
func (uc *SubjectExamUseCase) AdminDeleteQuestion(ctx context.Context, subjectLearningID, questionID string) error {
	set, err := uc.questionRepo.GetBySubjectLearningID(ctx, subjectLearningID)
	if err != nil {
		return err
	}
	out := make([]exam.ExamQuestionItem, 0, len(set.Questions))
	for _, q := range set.Questions {
		if q.ID == questionID {
			continue
		}
		out = append(out, q)
	}
	set.Questions = out
	return uc.questionRepo.Upsert(ctx, set)
}
