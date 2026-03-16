package usecase

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// TaskUseCase 封装学员任务的查询和状态更新逻辑。
type TaskUseCase struct {
	taskRepo    repository.TaskRepository
	studentRepo repository.StudentRepository
}

func NewTaskUseCase(taskRepo repository.TaskRepository, studentRepo repository.StudentRepository) *TaskUseCase {
	return &TaskUseCase{
		taskRepo:    taskRepo,
		studentRepo: studentRepo,
	}
}

// ListMyTasksRequest 学员任务列表请求。
type ListMyTasksRequest struct {
	Status   string `json:"status"`    // pending | done | overdue | all
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// ListMyTasksResponse 学员任务列表响应。
type ListMyTasksResponse struct {
	Items []*domain.Task `json:"items"`
	Total int            `json:"total"`
}

// ListMyTasks 按学员（当前用户）返回任务列表。
func (uc *TaskUseCase) ListMyTasks(ctx context.Context, userID string, req *ListMyTasksRequest) (*ListMyTasksResponse, error) {
	// 通过 students.user_id 精确映射出学员 ID。
	student, err := uc.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var statusFilter *domain.TaskStatus
	if req.Status != "" && req.Status != "all" {
		s := domain.TaskStatus(req.Status)
		statusFilter = &s
	}
	// 现有仓储一次性返回该学员的所有任务；如果后续任务量增大，可扩展为带 limit/offset 的查询。
	items, err := uc.taskRepo.ListByStudent(ctx, student.ID, statusFilter)
	if err != nil {
		return nil, err
	}
	return &ListMyTasksResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// UpdateTaskStatusRequest 更新任务状态请求。
type UpdateTaskStatusRequest struct {
	TaskID string `json:"task_id" binding:"required"`
	Status string `json:"status" binding:"required"` // in_progress | done | overdue
}

// UpdateTaskStatus 更新任务状态。
func (uc *TaskUseCase) UpdateTaskStatus(ctx context.Context, userID string, req *UpdateTaskStatusRequest) error {
	// TODO: 如有需要，这里可以先根据 userID 找到 studentID，再校验 task.StudentID 是否匹配。
	newStatus := domain.TaskStatus(req.Status)
	return uc.taskRepo.UpdateStatus(ctx, req.TaskID, newStatus)
}

