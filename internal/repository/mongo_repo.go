package repository

import "context"

// TheoryContentBlock is one rich-content block for a theory document.
// It is designed to be flexible enough for text, images and embedded videos.
type TheoryContentBlock struct {
	Type     string `bson:"type" json:"type"`                         // text | image | video_embed | ...
	Content  string `bson:"content,omitempty" json:"content,omitempty"` // text / HTML / Markdown
	URL      string `bson:"url,omitempty" json:"url,omitempty"`         // asset URL for image, etc.
	Caption  string `bson:"caption,omitempty" json:"caption,omitempty"`
	VideoID  string `bson:"video_id,omitempty" json:"video_id,omitempty"`
	Position int    `bson:"position,omitempty" json:"position,omitempty"`
}

// TheoryContentDoc is stored in MongoDB (rich content for theory document).
type TheoryContentDoc struct {
	DocumentID string               `bson:"document_id" json:"document_id"`
	Blocks     []TheoryContentBlock `bson:"blocks" json:"blocks"`
	UpdatedAt  int64                `bson:"updated_at" json:"updated_at"`
}

// TheoryContentRepository defines theory content (MongoDB) access.
type TheoryContentRepository interface {
	GetByDocumentID(ctx context.Context, documentID string) (*TheoryContentDoc, error)
	Upsert(ctx context.Context, doc *TheoryContentDoc) error
}

// TheoryQuestionOption represents a single option in a theory question.
type TheoryQuestionOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

// TheoryQuestionItem is a single question within a theory question set.
type TheoryQuestionItem struct {
	ID               string               `bson:"id" json:"id"`
	Type             string               `bson:"type" json:"type"` // single | multi
	Stem             string               `bson:"stem" json:"stem"`
	Options          []TheoryQuestionOption `bson:"options" json:"options"`
	CorrectOptionIDs []string             `bson:"correct_option_ids" json:"-"`          // not sent to client
	Explanation      string               `bson:"explanation,omitempty" json:"-"`       // not sent to client
	Tags             []string             `bson:"tags,omitempty" json:"tags,omitempty"` // for filtering / stats
	Score            int                  `bson:"score" json:"score"`
}

// TheoryQuestionSetDoc stores all questions for one theory document.
type TheoryQuestionSetDoc struct {
	DocumentID string              `bson:"document_id" json:"document_id"`
	Questions  []TheoryQuestionItem `bson:"questions" json:"questions"`
	UpdatedAt  int64               `bson:"updated_at" json:"updated_at"`
}

// TheoryQuestionSetRepository defines access to theory question sets.
type TheoryQuestionSetRepository interface {
	GetByDocumentID(ctx context.Context, documentID string) (*TheoryQuestionSetDoc, error)
	Upsert(ctx context.Context, doc *TheoryQuestionSetDoc) error
}

// TheoryAnswerRecord is the per-question answer result in an attempt.
type TheoryAnswerRecord struct {
	QuestionID        string   `bson:"question_id" json:"question_id"`
	SelectedOptionIDs []string `bson:"selected_option_ids" json:"selected_option_ids"`
	IsCorrect         bool     `bson:"is_correct" json:"is_correct"`
	Score             int      `bson:"score" json:"score"`
}

// TheoryAttemptDetailDoc stores detailed answers for one theory attempt.
type TheoryAttemptDetailDoc struct {
	TheoryAttemptID string              `bson:"theory_attempt_id" json:"theory_attempt_id"`
	Answers         []TheoryAnswerRecord `bson:"answers" json:"answers"`
	AutoGraded      bool                `bson:"auto_graded" json:"auto_graded"`
	CreatedAt       int64               `bson:"created_at" json:"created_at"`
}

// TheoryAttemptDetailRepository defines access to detailed attempt docs.
type TheoryAttemptDetailRepository interface {
	GetByAttemptID(ctx context.Context, attemptID string) (*TheoryAttemptDetailDoc, error)
	Upsert(ctx context.Context, doc *TheoryAttemptDetailDoc) error
}

// ContractFileMeta 记录存储在 MongoDB / GridFS 中的合同文档元数据。
type ContractFileMeta struct {
	ID         string   `bson:"_id,omitempty" json:"id"`
	StudentID  string   `bson:"student_id" json:"student_id"`
	ContractID string   `bson:"contract_id" json:"contract_id"`
	Filename   string   `bson:"filename" json:"filename"`
	ContentType string  `bson:"content_type" json:"content_type"`
	Size       int64    `bson:"size" json:"size"`
	GridFSID   string   `bson:"gridfs_id" json:"gridfs_id"`
	Tags       []string `bson:"tags,omitempty" json:"tags,omitempty"`
	UploadedAt int64    `bson:"uploaded_at" json:"uploaded_at"`
	UploadedBy string   `bson:"uploaded_by" json:"uploaded_by"`
}

// ContractFileMetaRepository 定义合同文档元数据访问接口。
type ContractFileMetaRepository interface {
	Insert(ctx context.Context, doc *ContractFileMeta) error
	ListByContract(ctx context.Context, contractID string) ([]*ContractFileMeta, error)
	GetByID(ctx context.Context, id string) (*ContractFileMeta, error)
}

