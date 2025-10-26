package usecases

import (
	"context"
	"errors"

	"github.com/clevertechware/todo-bun-app/internal/app/db"
	"github.com/clevertechware/todo-bun-app/internal/app/models"
)

// TaskUsecase defines the interface for task business logic
type TaskUsecase interface {
	CreateTask(ctx context.Context, params CreateTaskParams) (*TaskResult, error)
	DeleteTask(ctx context.Context, taskID int64) error
	GetTask(ctx context.Context, taskID int64) (*TaskResult, error)
	ListTasks(ctx context.Context) (*TaskListResult, error)
}

// taskUsecase implements TaskUsecase
type taskUsecase struct {
	taskRepo db.TaskRepository
}

// NewTaskUsecase creates a new instance of TaskUsecase
func NewTaskUsecase(taskRepo db.TaskRepository) TaskUsecase {
	return &taskUsecase{
		taskRepo: taskRepo,
	}
}

// CreateTask creates a new task with items
func (u *taskUsecase) CreateTask(ctx context.Context, params CreateTaskParams) (*TaskResult, error) {
	// Validate input
	if params.Title == "" {
		return nil, errors.New("task title is required")
	}

	// Convert params to model
	task := &models.Task{
		Title:       params.Title,
		Description: params.Description,
		Items:       make([]*models.TaskItem, 0, len(params.Items)),
	}

	for _, itemParam := range params.Items {
		if itemParam.Title == "" {
			return nil, errors.New("task item title is required")
		}

		item := &models.TaskItem{
			Title:     itemParam.Title,
			Completed: itemParam.Completed,
		}
		task.Items = append(task.Items, item)
	}

	// Create in repository
	if err := u.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// Convert model to result
	return u.modelToResult(task), nil
}

// DeleteTask deletes a task by ID
func (u *taskUsecase) DeleteTask(ctx context.Context, taskID int64) error {
	if taskID <= 0 {
		return errors.New("invalid task ID")
	}

	return u.taskRepo.Delete(ctx, taskID)
}

// GetTask retrieves a task by ID
func (u *taskUsecase) GetTask(ctx context.Context, taskID int64) (*TaskResult, error) {
	if taskID <= 0 {
		return nil, errors.New("invalid task ID")
	}

	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return u.modelToResult(task), nil
}

// ListTasks retrieves all tasks
func (u *taskUsecase) ListTasks(ctx context.Context) (*TaskListResult, error) {
	tasks, err := u.taskRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]TaskResult, 0, len(tasks))
	for _, task := range tasks {
		results = append(results, *u.modelToResult(task))
	}

	return &TaskListResult{
		Tasks: results,
	}, nil
}

// modelToResult converts a Task model to TaskResult
func (u *taskUsecase) modelToResult(task *models.Task) *TaskResult {
	items := make([]TaskItemResult, 0, len(task.Items))
	for _, item := range task.Items {
		items = append(items, TaskItemResult{
			ID:        item.ID,
			TaskID:    item.TaskID,
			Title:     item.Title,
			Completed: item.Completed,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return &TaskResult{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Items:       items,
	}
}
