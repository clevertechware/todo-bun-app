package usecases

import "time"

// TaskItemResult represents a task item in the output
type TaskItemResult struct {
	ID        int64
	TaskID    int64
	Title     string
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TaskResult represents a task in the output
type TaskResult struct {
	ID          int64
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items       []TaskItemResult
}

// TaskListResult represents a list of tasks
type TaskListResult struct {
	Tasks []TaskResult
}
