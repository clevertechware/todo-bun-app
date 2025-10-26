package db

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/clevertechware/todo-bun-app/internal/app/models"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, taskID int64) error
	GetByID(ctx context.Context, taskID int64) (*models.Task, error)
	List(ctx context.Context) ([]*models.Task, error)
}

// taskRepository implements TaskRepository using Bun
type taskRepository struct {
	db bun.IDB
}

// NewTaskRepository creates a new instance of TaskRepository
func NewTaskRepository(db bun.IDB) TaskRepository {
	return &taskRepository{db: db}
}

// Create inserts a new task with its items in a transaction
func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Set timestamps
		now := time.Now()
		task.CreatedAt = now
		task.UpdatedAt = now

		// Insert the task
		if _, err := tx.NewInsert().
			Model(task).
			Exec(ctx); err != nil {
			return err
		}

		// Insert task items if any
		if len(task.Items) > 0 {
			for _, item := range task.Items {
				item.TaskID = task.ID
				item.CreatedAt = now
				item.UpdatedAt = now
			}

			if _, err := tx.NewInsert().
				Model(&task.Items).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete removes a task by ID (cascade deletes items via FK constraint)
func (r *taskRepository) Delete(ctx context.Context, taskID int64) error {
	result, err := r.db.NewDelete().
		Model((*models.Task)(nil)).
		Where("id = ?", taskID).
		Exec(ctx)

	if err != nil {
		return err
	}

	// Check if the task was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

// GetByID retrieves a task by ID with its items
func (r *taskRepository) GetByID(ctx context.Context, taskID int64) (*models.Task, error) {
	task := new(models.Task)

	err := r.db.NewSelect().
		Model(task).
		Where("t.id = ?", taskID).
		Relation("Items").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return task, nil
}

// List retrieves all tasks with their items
func (r *taskRepository) List(ctx context.Context) ([]*models.Task, error) {
	var tasks []*models.Task

	err := r.db.NewSelect().
		Model(&tasks).
		Relation("Items").
		Order("t.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}
