package repository

import (
	"context"

	"magua-fly-golang/internal/exam"
)

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

// TheoryQuestionOption is an alias for shared exam option types.
type TheoryQuestionOption = exam.ExamQuestionOption

// TheoryQuestionItem is an alias for shared exam question types.
type TheoryQuestionItem = exam.ExamQuestionItem

// TheoryQuestionSetDoc stores all questions for one theory document.
type TheoryQuestionSetDoc struct {
	DocumentID string               `bson:"document_id" json:"document_id"`
	Questions  []exam.ExamQuestionItem `bson:"questions" json:"questions"`
	UpdatedAt  int64                `bson:"updated_at" json:"updated_at"`
}

// TheoryQuestionSetRepository defines access to theory question sets.
type TheoryQuestionSetRepository interface {
	GetByDocumentID(ctx context.Context, documentID string) (*TheoryQuestionSetDoc, error)
	Upsert(ctx context.Context, doc *TheoryQuestionSetDoc) error
}

// TheoryAnswerRecord is an alias for per-question graded rows.
type TheoryAnswerRecord = exam.ExamAnswerRecord

// TheoryAttemptDetailDoc stores detailed answers for one theory attempt.
type TheoryAttemptDetailDoc struct {
	TheoryAttemptID string                 `bson:"theory_attempt_id" json:"theory_attempt_id"`
	Answers         []exam.ExamAnswerRecord `bson:"answers" json:"answers"`
	AutoGraded      bool                   `bson:"auto_graded" json:"auto_graded"`
	CreatedAt       int64                  `bson:"created_at" json:"created_at"`
}

// SubjectQuestionSetDoc stores all questions for one subject_learning unit (same shape as theory).
type SubjectQuestionSetDoc struct {
	SubjectLearningID string                 `bson:"subject_learning_id" json:"subject_learning_id"`
	Questions         []exam.ExamQuestionItem `bson:"questions" json:"questions"`
	UpdatedAt         int64                  `bson:"updated_at" json:"updated_at"`
}

// SubjectQuestionSetRepository defines access to subject question sets in MongoDB.
type SubjectQuestionSetRepository interface {
	GetBySubjectLearningID(ctx context.Context, subjectLearningID string) (*SubjectQuestionSetDoc, error)
	Upsert(ctx context.Context, doc *SubjectQuestionSetDoc) error
}

// SubjectAttemptDetailDoc stores detailed answers for one subject attempt.
type SubjectAttemptDetailDoc struct {
	SubjectAttemptID string                 `bson:"subject_attempt_id" json:"subject_attempt_id"`
	Answers          []exam.ExamAnswerRecord `bson:"answers" json:"answers"`
	AutoGraded       bool                    `bson:"auto_graded" json:"auto_graded"`
	CreatedAt        int64                   `bson:"created_at" json:"created_at"`
}

// SubjectAttemptDetailRepository defines access to subject attempt detail docs.
type SubjectAttemptDetailRepository interface {
	GetByAttemptID(ctx context.Context, attemptID string) (*SubjectAttemptDetailDoc, error)
	Upsert(ctx context.Context, doc *SubjectAttemptDetailDoc) error
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

