package exam

// ExamQuestionOption is one selectable option in a multiple-choice question.
type ExamQuestionOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

// ExamQuestionItem is one question in a question set (theory or subject learning).
type ExamQuestionItem struct {
	ID               string               `bson:"id" json:"id"`
	Type             string               `bson:"type" json:"type"` // single | multi
	Stem             string               `bson:"stem" json:"stem"`
	Options          []ExamQuestionOption `bson:"options" json:"options"`
	CorrectOptionIDs []string             `bson:"correct_option_ids" json:"-"`
	Explanation      string               `bson:"explanation,omitempty" json:"-"`
	Tags             []string             `bson:"tags,omitempty" json:"tags,omitempty"`
	Score            int                  `bson:"score" json:"score"`
}

// AdminListExamQuestionItem is the JSON shape for admin question list/create-edit UIs.
// It exposes correct_option_ids and explanation; ExamQuestionItem omits them (json:"-")
// so student-facing start/submit flows never leak answers.
type AdminListExamQuestionItem struct {
	ID               string               `json:"id"`
	Type             string               `json:"type"`
	Stem             string               `json:"stem"`
	Options          []ExamQuestionOption `json:"options"`
	CorrectOptionIDs []string             `json:"correct_option_ids"`
	Explanation      string               `json:"explanation"`
	Tags             []string             `json:"tags,omitempty"`
	Score            int                  `json:"score"`
}

// ToAdminListItem copies a stored question into the admin-list JSON DTO.
func ToAdminListItem(q ExamQuestionItem) AdminListExamQuestionItem {
	return AdminListExamQuestionItem{
		ID:               q.ID,
		Type:             q.Type,
		Stem:             q.Stem,
		Options:          q.Options,
		CorrectOptionIDs: q.CorrectOptionIDs,
		Explanation:      q.Explanation,
		Tags:             q.Tags,
		Score:            q.Score,
	}
}

// ExamAnswerRecord is the graded result for one question in an attempt.
type ExamAnswerRecord struct {
	QuestionID        string   `bson:"question_id" json:"question_id"`
	SelectedOptionIDs []string `bson:"selected_option_ids" json:"selected_option_ids"`
	IsCorrect         bool     `bson:"is_correct" json:"is_correct"`
	Score             int      `bson:"score" json:"score"`
}
