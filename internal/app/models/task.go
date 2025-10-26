package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Task represents a task with multiple items
type Task struct {
	bun.BaseModel `bun:"table:tasks,alias:t"`

	ID          int64       `bun:"id,pk,autoincrement"`
	Title       string      `bun:"title,notnull"`
	Description string      `bun:"description"`
	CreatedAt   time.Time   `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time   `bun:"updated_at,notnull,default:current_timestamp"`
	Items       []*TaskItem `bun:"rel:has-many,join:id=task_id"`
}
