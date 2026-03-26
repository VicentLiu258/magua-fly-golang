package domain

import "time"

// 理论课视频存储方式（PostgreSQL theory_documents.video_storage）
const (
	TheoryVideoStorageExternal = "external"
	TheoryVideoStorageGridFS   = "gridfs"
	TheoryVideoStorageCOS      = "cos"
)

// TheoryDocument is metadata for a theory document (content in MongoDB).
type TheoryDocument struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug,omitempty"`
	SortOrder  int       `json:"sort_order"`
	ContentRef string    `json:"content_ref,omitempty"`
	VideoURL   string    `json:"video_url,omitempty"`
	// VideoStorage：外链 external；托管 gridfs / cos（与 VideoObjectID 配合）
	VideoStorage         string `json:"video_storage,omitempty"`
	VideoObjectID        string `json:"video_object_id,omitempty"`
	VideoContentType     string `json:"video_content_type,omitempty"`
	VideoSizeBytes       int64  `json:"video_size_bytes,omitempty"`
	VideoDurationSeconds int    `json:"video_duration_seconds,omitempty"`
	Category             string `json:"category,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// HasHostedTheoryVideo 是否为 GridFS/COS 托管视频
func (d *TheoryDocument) HasHostedTheoryVideo() bool {
	if d == nil {
		return false
	}
	return d.VideoStorage == TheoryVideoStorageGridFS || d.VideoStorage == TheoryVideoStorageCOS
}

// HasExternalTheoryVideo 是否为外链视频（显式 external 或仅填 video_url 的兼容行）
func (d *TheoryDocument) HasExternalTheoryVideo() bool {
	if d == nil {
		return false
	}
	if d.VideoStorage == TheoryVideoStorageExternal {
		return true
	}
	return d.VideoStorage == "" && d.VideoURL != ""
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

// TheoryAttempt is a user's quiz attempt summary (answers are stored in MongoDB).
type TheoryAttempt struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	TheoryDocumentID string    `json:"theory_document_id"`
	// AnswersJSON stores the raw JSONB payload persisted in PostgreSQL; detailed
	// per-question grading is stored separately in MongoDB.
	AnswersJSON string    `json:"-"`
	Score       int       `json:"score"`
	SubmittedAt time.Time `json:"submitted_at"`

	DurationSeconds int  `json:"duration_seconds,omitempty"`
	MaxScore        int  `json:"max_score,omitempty"`
	CorrectCount    int  `json:"correct_count,omitempty"`
	TotalCount      int  `json:"total_count,omitempty"`
	Passed          bool `json:"passed,omitempty"`
}

// SubjectLearning is metadata for subject learning (1:1 with syllabus subject).
type SubjectLearning struct {
	ID                 string    `json:"id"`
	SyllabusSubjectID  string    `json:"syllabus_subject_id"`
	Title              string    `json:"title"`
	ContentRef         string    `json:"content_ref,omitempty"`
	VideoURL           string    `json:"video_url,omitempty"`
	SortOrder          int       `json:"sort_order,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SubjectAttempt is a user's quiz attempt for a subject learning unit (details in MongoDB).
type SubjectAttempt struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	SubjectLearningID string    `json:"subject_learning_id"`
	AnswersJSON       string    `json:"-"`
	Score             int       `json:"score"`
	SubmittedAt       time.Time `json:"submitted_at"`

	DurationSeconds int  `json:"duration_seconds,omitempty"`
	MaxScore        int  `json:"max_score,omitempty"`
	CorrectCount    int  `json:"correct_count,omitempty"`
	TotalCount      int  `json:"total_count,omitempty"`
	Passed          bool `json:"passed,omitempty"`
}

// TheoryPaper represents an exam paper under a theory document.
type TheoryPaper struct {
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

// TheoryPaperQuestion links a paper to one question in the question set.
type TheoryPaperQuestion struct {
	ID        string `json:"id"`
	PaperID   string `json:"paper_id"`
	QuestionID string `json:"question_id"`
	Score     int    `json:"score"`
	SortOrder int    `json:"sort_order"`
}
