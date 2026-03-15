package domain

import "time"

// TheoryDocument is metadata for a theory document (content in MongoDB).
type TheoryDocument struct {
	ID         string    `json:"id"`
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
	ID               string    `json:"id"`
	TheoryDocumentID string    `json:"theory_document_id"`
	QuestionText     string    `json:"question_text"`
	Options          []string  `json:"options"`
	CorrectIndex     int       `json:"correct_index"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TheoryAttempt is a user's quiz attempt.
type TheoryAttempt struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	TheoryDocumentID string    `json:"theory_document_id"`
	Answers          []int     `json:"answers"`
	Score            int       `json:"score"`
	SubmittedAt      time.Time `json:"submitted_at"`
}

// SubjectLearning is metadata for subject learning (1:1 with syllabus subject).
type SubjectLearning struct {
	ID                 string    `json:"id"`
	SyllabusSubjectID  string    `json:"syllabus_subject_id"`
	Title              string    `json:"title"`
	ContentRef         string    `json:"content_ref,omitempty"`
	VideoURL           string    `json:"video_url,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
