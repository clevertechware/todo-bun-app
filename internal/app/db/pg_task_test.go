package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"github.com/clevertechware/todo-bun-app/internal/app/models"
)

func (s *PGRepositorySuite) TestPGTask_Create() {
	type args struct {
		ctx  context.Context
		task *models.Task
	}

	tests := []struct {
		name    string
		args    args
		seed    func(t *testing.T, client bun.IDB)
		check   func(t *testing.T, client bun.IDB, err error, task *models.Task)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should create a new task with items",
			args: args{
				ctx: context.Background(),
				task: &models.Task{
					Title:       "Shopping",
					Description: "Weekly shopping",
					Items: []*models.TaskItem{
						{Title: "Buy milk", Completed: false},
						{Title: "Buy bread", Completed: true},
					},
				},
			},
			seed: func(t *testing.T, client bun.IDB) {
				// No seed needed
			},
			check: func(t *testing.T, client bun.IDB, err error, task *models.Task) {
				require.NoError(t, err)

				// Verify task was created
				assert.NotZero(t, task.ID)
				assert.Equal(t, "Shopping", task.Title)
				assert.Equal(t, "Weekly shopping", task.Description)
				assert.NotZero(t, task.CreatedAt)
				assert.NotZero(t, task.UpdatedAt)

				// Verify items were created
				var items []*models.TaskItem
				err = client.NewSelect().
					Model(&items).
					Where("task_id = ?", task.ID).
					Order("id ASC").
					Scan(context.Background())
				require.NoError(t, err)
				assert.Len(t, items, 2)
				assert.Equal(t, "Buy milk", items[0].Title)
				assert.False(t, items[0].Completed)
				assert.Equal(t, "Buy bread", items[1].Title)
				assert.True(t, items[1].Completed)
			},
			wantErr: assert.NoError,
		},
		{
			name: "should create a new task without items",
			args: args{
				ctx: context.Background(),
				task: &models.Task{
					Title:       "Work",
					Description: "Project tasks",
					Items:       []*models.TaskItem{},
				},
			},
			seed: func(t *testing.T, client bun.IDB) {
				// No seed needed
			},
			check: func(t *testing.T, client bun.IDB, err error, task *models.Task) {
				require.NoError(t, err)

				// Verify task was created
				assert.NotZero(t, task.ID)
				assert.Equal(t, "Work", task.Title)

				// Verify no items were created
				var count int
				count, err = client.NewSelect().
					Model((*models.TaskItem)(nil)).
					Where("task_id = ?", task.ID).
					Count(context.Background())
				require.NoError(t, err)
				assert.Equal(t, 0, count)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			t := s.T()

			trx, err := s.pgContainer.TxBegin()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, trx.Rollback())
			}()

			repo := NewTaskRepository(trx)
			if tt.seed != nil {
				tt.seed(t, trx)
			}

			err = repo.Create(tt.args.ctx, tt.args.task)

			tt.wantErr(t, err)
			if tt.check != nil {
				tt.check(t, trx, err, tt.args.task)
			}
		})
	}
}

func (s *PGRepositorySuite) TestPGTask_GetByID() {
	type args struct {
		ctx    context.Context
		taskID int64
	}

	tests := []struct {
		name    string
		args    args
		seed    func(t *testing.T, client bun.IDB) int64
		check   func(t *testing.T, task *models.Task, err error)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should get task with items",
			args: args{
				ctx: context.Background(),
			},
			seed: func(t *testing.T, client bun.IDB) int64 {
				task := &models.Task{
					Title:       "Shopping",
					Description: "Weekly shopping",
				}
				s.insert(t, client, task)

				item1 := &models.TaskItem{
					TaskID:    task.ID,
					Title:     "Buy milk",
					Completed: false,
				}
				s.insert(t, client, item1)

				item2 := &models.TaskItem{
					TaskID:    task.ID,
					Title:     "Buy bread",
					Completed: true,
				}
				s.insert(t, client, item2)

				return task.ID
			},
			check: func(t *testing.T, task *models.Task, err error) {
				require.NoError(t, err)
				assert.Equal(t, "Shopping", task.Title)
				assert.Equal(t, "Weekly shopping", task.Description)
				assert.Len(t, task.Items, 2)
				assert.Equal(t, "Buy milk", task.Items[0].Title)
				assert.Equal(t, "Buy bread", task.Items[1].Title)
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when task not found",
			args: args{
				ctx:    context.Background(),
				taskID: 999,
			},
			seed: func(t *testing.T, client bun.IDB) int64 {
				return 999
			},
			check: func(t *testing.T, task *models.Task, err error) {
				require.Error(t, err)
				assert.Nil(t, task)
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			t := s.T()

			trx, err := s.pgContainer.TxBegin()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, trx.Rollback())
			}()

			repo := NewTaskRepository(trx)
			var taskID int64
			if tt.seed != nil {
				taskID = tt.seed(t, trx)
			}
			if tt.args.taskID == 0 {
				tt.args.taskID = taskID
			}

			task, err := repo.GetByID(tt.args.ctx, tt.args.taskID)

			tt.wantErr(t, err)
			if tt.check != nil {
				tt.check(t, task, err)
			}
		})
	}
}

func (s *PGRepositorySuite) TestPGTask_List() {
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		args    args
		seed    func(t *testing.T, client bun.IDB)
		check   func(t *testing.T, tasks []*models.Task, err error)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should list all tasks with items",
			args: args{
				ctx: context.Background(),
			},
			seed: func(t *testing.T, client bun.IDB) {
				// Create first task
				task1 := &models.Task{
					Title:       "Shopping",
					Description: "Weekly shopping",
				}
				s.insert(t, client, task1)
				s.insert(t, client, &models.TaskItem{
					TaskID:    task1.ID,
					Title:     "Buy milk",
					Completed: false,
				})

				// Create second task
				task2 := &models.Task{
					Title:       "Work",
					Description: "Project tasks",
				}
				s.insert(t, client, task2)
			},
			check: func(t *testing.T, tasks []*models.Task, err error) {
				require.NoError(t, err)
				assert.Len(t, tasks, 2)

				// Find tasks by title (order may vary)
				var shoppingTodo, workTodo *models.Task
				for _, task := range tasks {
					if task.Title == "Shopping" {
						shoppingTodo = task
					} else if task.Title == "Work" {
						workTodo = task
					}
				}

				require.NotNil(t, shoppingTodo, "Shopping task should exist")
				require.NotNil(t, workTodo, "Work task should exist")
				assert.Len(t, shoppingTodo.Items, 1)
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return empty list when no tasks exist",
			args: args{
				ctx: context.Background(),
			},
			seed: func(t *testing.T, client bun.IDB) {
				// No seed
			},
			check: func(t *testing.T, tasks []*models.Task, err error) {
				require.NoError(t, err)
				assert.Empty(t, tasks)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			t := s.T()

			trx, err := s.pgContainer.TxBegin()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, trx.Rollback())
			}()

			repo := NewTaskRepository(trx)
			if tt.seed != nil {
				tt.seed(t, trx)
			}

			tasks, err := repo.List(tt.args.ctx)

			tt.wantErr(t, err)
			if tt.check != nil {
				tt.check(t, tasks, err)
			}
		})
	}
}

func (s *PGRepositorySuite) TestPGTask_Delete() {
	type args struct {
		ctx    context.Context
		taskID int64
	}

	tests := []struct {
		name    string
		args    args
		seed    func(t *testing.T, client bun.IDB) int64
		check   func(t *testing.T, client bun.IDB, err error)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should delete task and cascade delete items",
			args: args{
				ctx: context.Background(),
			},
			seed: func(t *testing.T, client bun.IDB) int64 {
				task := &models.Task{
					Title:       "Shopping",
					Description: "Weekly shopping",
				}
				s.insert(t, client, task)

				item := &models.TaskItem{
					TaskID:    task.ID,
					Title:     "Buy milk",
					Completed: false,
				}
				s.insert(t, client, item)

				return task.ID
			},
			check: func(t *testing.T, client bun.IDB, err error) {
				require.NoError(t, err)

				// Verify task was deleted
				var count int
				count, err = client.NewSelect().
					Model((*models.Task)(nil)).
					Count(context.Background())
				require.NoError(t, err)
				assert.Equal(t, 0, count)

				// Verify items were cascade deleted
				count, err = client.NewSelect().
					Model((*models.TaskItem)(nil)).
					Count(context.Background())
				require.NoError(t, err)
				assert.Equal(t, 0, count)
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when task not found",
			args: args{
				ctx:    context.Background(),
				taskID: 999,
			},
			seed: func(t *testing.T, client bun.IDB) int64 {
				return 999
			},
			check: func(t *testing.T, client bun.IDB, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrTaskNotFound)
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			t := s.T()

			trx, err := s.pgContainer.TxBegin()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, trx.Rollback())
			}()

			repo := NewTaskRepository(trx)
			var taskID int64
			if tt.seed != nil {
				taskID = tt.seed(t, trx)
			}
			if tt.args.taskID == 0 {
				tt.args.taskID = taskID
			}

			err = repo.Delete(tt.args.ctx, tt.args.taskID)

			tt.wantErr(t, err)
			if tt.check != nil {
				tt.check(t, trx, err)
			}
		})
	}
}
