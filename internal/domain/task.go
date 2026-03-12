package domain

import "time"

// TaskType is the type of student task.
type TaskType string

const (
	TaskTypeTheory  TaskType = "theory"
	TaskTypeSubject TaskType = "subject"
	TaskTypeOther   TaskType = "other"
)

// TaskStatus for task completion.
type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusDone    TaskStatus = "done"
	TaskStatusOverdue TaskStatus = "overdue"
)

// Task is a student task.
type Task struct {
	ID          int64      `json:"id"`
	StudentID   int64      `json:"student_id"`
	AssignerID  int64      `json:"assigner_id"`
	Type        TaskType   `json:"type"`
	RefID       *int64     `json:"ref_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
