package domain

import "time"

// TheoryDocument is metadata for a theory document (content in MongoDB).
type TheoryDocument struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug,omitempty"`
	SortOrder  int       `json:"sort_order"`
	ContentRef string    `json:"content_ref,omitempty"`
	VideoURL   string    `json:"video_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TheoryQuestion is a single choice question for a theory document.
type TheoryQuestion struct {
	ID               int64     `json:"id"`
	TheoryDocumentID int64     `json:"theory_document_id"`
	QuestionText     string    `json:"question_text"`
	Options          []string  `json:"options"`
	CorrectIndex     int       `json:"correct_index"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TheoryAttempt is a user's quiz attempt.
type TheoryAttempt struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	TheoryDocumentID int64     `json:"theory_document_id"`
	Answers          []int     `json:"answers"`
	Score            int       `json:"score"`
	SubmittedAt      time.Time `json:"submitted_at"`
}

// SubjectLearning is metadata for subject learning (1:1 with syllabus subject).
type SubjectLearning struct {
	ID                 int64     `json:"id"`
	SyllabusSubjectID  int64     `json:"syllabus_subject_id"`
	Title              string    `json:"title"`
	ContentRef         string    `json:"content_ref,omitempty"`
	VideoURL           string    `json:"video_url,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
