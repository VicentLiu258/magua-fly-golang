package exam

import (
	"fmt"
	"time"
)

// GenerateQuestionID returns a client-agnostic unique id for a new question row.
func GenerateQuestionID() string {
	return fmt.Sprintf("q_%d", time.Now().UnixNano())
}
