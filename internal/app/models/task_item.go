package models

import (
	"time"

	"github.com/uptrace/bun"
)

// TaskItem represents an item belonging to a task
type TaskItem struct {
	bun.BaseModel `bun:"table:task_items,alias:ti"`

	ID        int64     `bun:"id,pk,autoincrement"`
	TaskID    int64     `bun:"task_id,notnull"`
	Title     string    `bun:"title,notnull"`
	Completed bool      `bun:"completed,notnull,default:false"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
	Task      *Task     `bun:"rel:belongs-to,join:task_id=id"`
}
