package exam

// SameStringSet compares two string slices as sets (order-insensitive, duplicates counted).
func SameStringSet(a, b []string) bool {
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

// AnswerInput is one submitted answer from the client.
type AnswerInput struct {
	QuestionID        string
	SelectedOptionIDs []string
}

// GradeResult aggregates auto-grading output for a multiple-choice attempt.
type GradeResult struct {
	Answers      []ExamAnswerRecord
	Score        int
	MaxScore     int
	CorrectCount int
	TotalCount   int
}

// GradeMCQ grades answers against a question index built from the official question set.
// Unknown question IDs are skipped (no score, no record).
func GradeMCQ(questionByID map[string]ExamQuestionItem, answers []AnswerInput) GradeResult {
	var out GradeResult
	out.TotalCount = len(answers)
	for _, a := range answers {
		q, ok := questionByID[a.QuestionID]
		if !ok {
			continue
		}
		out.MaxScore += q.Score
		rec := ExamAnswerRecord{
			QuestionID:        a.QuestionID,
			SelectedOptionIDs: a.SelectedOptionIDs,
		}
		if SameStringSet(a.SelectedOptionIDs, q.CorrectOptionIDs) {
			rec.IsCorrect = true
			rec.Score = q.Score
			out.CorrectCount++
			out.Score += q.Score
		}
		out.Answers = append(out.Answers, rec)
	}
	return out
}
