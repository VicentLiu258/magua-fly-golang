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

// ExamAnswerRecord is the graded result for one question in an attempt.
type ExamAnswerRecord struct {
	QuestionID        string   `bson:"question_id" json:"question_id"`
	SelectedOptionIDs []string `bson:"selected_option_ids" json:"selected_option_ids"`
	IsCorrect         bool     `bson:"is_correct" json:"is_correct"`
	Score             int      `bson:"score" json:"score"`
}
