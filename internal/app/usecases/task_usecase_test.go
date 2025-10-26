package usecases

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/clevertechware/todo-bun-app/internal/app/db"
	"github.com/clevertechware/todo-bun-app/internal/app/db/mocks"
	"github.com/clevertechware/todo-bun-app/internal/app/models"
)

func TestTaskUsecase_CreateTask(t *testing.T) {
	t.Parallel()

	type fields struct {
		taskRepo func(t *testing.T) db.TaskRepository
	}

	type args struct {
		ctx    context.Context
		params CreateTaskParams
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TaskResult
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should create task successfully",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("Create", mock.Anything, mock.MatchedBy(func(task *models.Task) bool {
						return task.Title == "Buy groceries" && task.Description == "Weekly shopping"
					})).Run(func(args mock.Arguments) {
						task := args.Get(1).(*models.Task)
						task.ID = 1
						task.Items = []*models.TaskItem{
							{ID: 1, TaskID: 1, Title: "Buy milk", Completed: false},
							{ID: 2, TaskID: 1, Title: "Buy bread", Completed: false},
						}
					}).Return(nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				params: CreateTaskParams{
					Title:       "Buy groceries",
					Description: "Weekly shopping",
					Items: []CreateTaskItemParams{
						{Title: "Buy milk", Completed: false},
						{Title: "Buy bread", Completed: false},
					},
				},
			},
			want: &TaskResult{
				ID:          1,
				Title:       "Buy groceries",
				Description: "Weekly shopping",
				Items: []TaskItemResult{
					{ID: 1, TaskID: 1, Title: "Buy milk", Completed: false},
					{ID: 2, TaskID: 1, Title: "Buy bread", Completed: false},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when title is empty",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					// Repository should not be called
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				params: CreateTaskParams{
					Title:       "",
					Description: "No title",
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "should return error when item title is empty",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					// Repository should not be called
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				params: CreateTaskParams{
					Title:       "Shopping",
					Description: "Weekly shopping",
					Items: []CreateTaskItemParams{
						{Title: "", Completed: false},
					},
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "should return error when repository fails",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				params: CreateTaskParams{
					Title:       "Shopping",
					Description: "Weekly shopping",
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := &taskUsecase{
				taskRepo: tt.fields.taskRepo(t),
			}

			got, err := u.CreateTask(tt.args.ctx, tt.args.params)

			if !tt.wantErr(t, err) {
				return
			}

			if tt.want != nil {
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Description, got.Description)
				assert.Equal(t, len(tt.want.Items), len(got.Items))
			}
		})
	}
}

func TestTaskUsecase_GetTask(t *testing.T) {
	t.Parallel()

	type fields struct {
		taskRepo func(t *testing.T) db.TaskRepository
	}

	type args struct {
		ctx    context.Context
		todoID int64
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TaskResult
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should get task successfully",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("GetByID", mock.Anything, int64(1)).Return(&models.Task{
						ID:          1,
						Title:       "Shopping",
						Description: "Weekly shopping",
						Items: []*models.TaskItem{
							{ID: 1, TaskID: 1, Title: "Buy milk", Completed: false},
						},
					}, nil)
					return m
				},
			},
			args: args{
				ctx:    context.Background(),
				todoID: 1,
			},
			want: &TaskResult{
				ID:          1,
				Title:       "Shopping",
				Description: "Weekly shopping",
				Items: []TaskItemResult{
					{ID: 1, TaskID: 1, Title: "Buy milk", Completed: false},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when task ID is invalid",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					// Repository should not be called
					return m
				},
			},
			args: args{
				ctx:    context.Background(),
				todoID: 0,
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "should return error when repository returns not found",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("GetByID", mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)
					return m
				},
			},
			args: args{
				ctx:    context.Background(),
				todoID: 999,
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := &taskUsecase{
				taskRepo: tt.fields.taskRepo(t),
			}

			got, err := u.GetTask(tt.args.ctx, tt.args.todoID)

			if !tt.wantErr(t, err) {
				return
			}

			if tt.want != nil {
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Description, got.Description)
				assert.Equal(t, len(tt.want.Items), len(got.Items))
			}
		})
	}
}

func TestTaskUsecase_DeleteTask(t *testing.T) {
	t.Parallel()

	type fields struct {
		taskRepo func(t *testing.T) db.TaskRepository
	}

	type args struct {
		ctx    context.Context
		todoID int64
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should delete task successfully",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("Delete", mock.Anything, int64(1)).Return(nil)
					return m
				},
			},
			args: args{
				ctx:    context.Background(),
				todoID: 1,
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when task ID is invalid",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					// Repository should not be called
					return m
				},
			},
			args: args{
				ctx:    context.Background(),
				todoID: 0,
			},
			wantErr: assert.Error,
		},
		{
			name: "should return error when repository fails",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("Delete", mock.Anything, int64(999)).Return(db.ErrTaskNotFound)
					return m
				},
			},
			args: args{
				ctx:    context.Background(),
				todoID: 999,
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := &taskUsecase{
				taskRepo: tt.fields.taskRepo(t),
			}

			err := u.DeleteTask(tt.args.ctx, tt.args.todoID)
			tt.wantErr(t, err)
		})
	}
}

func TestTaskUsecase_ListTasks(t *testing.T) {
	t.Parallel()

	type fields struct {
		taskRepo func(t *testing.T) db.TaskRepository
	}

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TaskListResult
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should list todos successfully",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("List", mock.Anything).Return([]*models.Task{
						{
							ID:          1,
							Title:       "Shopping",
							Description: "Weekly shopping",
							Items:       []*models.TaskItem{{ID: 1, TaskID: 1, Title: "Buy milk", Completed: false}},
						},
						{
							ID:          2,
							Title:       "Work",
							Description: "Project tasks",
							Items:       []*models.TaskItem{},
						},
					}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: &TaskListResult{
				Tasks: []TaskResult{
					{ID: 1, Title: "Shopping", Description: "Weekly shopping", Items: []TaskItemResult{{ID: 1, TaskID: 1, Title: "Buy milk", Completed: false}}},
					{ID: 2, Title: "Work", Description: "Project tasks", Items: []TaskItemResult{}},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return empty list when no todos exist",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("List", mock.Anything).Return([]*models.Task{}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: &TaskListResult{
				Tasks: []TaskResult{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error when repository fails",
			fields: fields{
				taskRepo: func(t *testing.T) db.TaskRepository {
					m := mocks.NewTaskRepository(t)
					m.On("List", mock.Anything).Return(nil, errors.New("database error"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := &taskUsecase{
				taskRepo: tt.fields.taskRepo(t),
			}

			got, err := u.ListTasks(tt.args.ctx)

			if !tt.wantErr(t, err) {
				return
			}

			if tt.want != nil {
				assert.Equal(t, len(tt.want.Tasks), len(got.Tasks))
			}
		})
	}
}
