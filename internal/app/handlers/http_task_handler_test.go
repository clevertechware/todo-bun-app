package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/clevertechware/todo-bun-app/internal/app/db"
	"github.com/clevertechware/todo-bun-app/internal/app/usecases"
	"github.com/clevertechware/todo-bun-app/internal/app/usecases/mocks"
)

func TestHTTPTaskHandler_CreateTask(t *testing.T) {
	t.Parallel()

	type args struct {
		method      string
		url         string
		requestBody interface{}
	}

	type setup func(t *testing.T, mockUsecase *mocks.TaskUsecase)

	tests := []struct {
		name             string
		args             args
		setup            setup
		wantStatus       int
		wantResponseBody interface{}
	}{
		{
			name: "should return 201 when task is created successfully",
			args: args{
				method: http.MethodPost,
				url:    "/api/tasks",
				requestBody: map[string]interface{}{
					"title":       "Shopping",
					"description": "Weekly shopping",
					"items": []map[string]interface{}{
						{"title": "Buy milk", "completed": false},
						{"title": "Buy bread", "completed": false},
					},
				},
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("CreateTask", mock.Anything, mock.MatchedBy(func(params usecases.CreateTaskParams) bool {
					return params.Title == "Shopping" &&
						params.Description == "Weekly shopping" &&
						len(params.Items) == 2
				})).Return(&usecases.TaskResult{
					ID:          1,
					Title:       "Shopping",
					Description: "Weekly shopping",
					Items: []usecases.TaskItemResult{
						{ID: 1, TaskID: 1, Title: "Buy milk", Completed: false},
						{ID: 2, TaskID: 1, Title: "Buy bread", Completed: false},
					},
				}, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "should return 400 when title is missing",
			args: args{
				method: http.MethodPost,
				url:    "/api/tasks",
				requestBody: map[string]interface{}{
					"description": "No title",
				},
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				// Usecase should not be called
			},
			wantStatus: http.StatusBadRequest,
			wantResponseBody: map[string]string{
				"title": "required",
			},
		},
		{
			name: "should return 400 when request body is invalid JSON",
			args: args{
				method:      http.MethodPost,
				url:         "/api/tasks",
				requestBody: "invalid json",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				// Usecase should not be called
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 500 when usecase returns error",
			args: args{
				method: http.MethodPost,
				url:    "/api/tasks",
				requestBody: map[string]interface{}{
					"title":       "Shopping",
					"description": "Weekly shopping",
				},
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("CreateTask", mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockUsecase := mocks.NewTaskUsecase(t)
			if tt.setup != nil {
				tt.setup(t, mockUsecase)
			}

			// Setup handler and router
			handler := HTTPHandler{
				httpTaskHandler: NewHTTPTaskHandler(mockUsecase),
			}
			router := gin.Default()
			api := router.Group("/api")
			handler.registerTaskRoutes(api)

			// Create request
			var body bytes.Buffer
			if tt.args.requestBody != nil {
				if strBody, ok := tt.args.requestBody.(string); ok {
					body = *bytes.NewBufferString(strBody)
				} else {
					err := json.NewEncoder(&body).Encode(tt.args.requestBody)
					require.NoError(t, err)
				}
			}

			req, err := http.NewRequestWithContext(context.Background(), tt.args.method, tt.args.url, &body)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantResponseBody != nil {
				var actualBody map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &actualBody)
				require.NoError(t, err)

				expectedBody, ok := tt.wantResponseBody.(map[string]string)
				if ok {
					for key, expectedValue := range expectedBody {
						assert.Equal(t, expectedValue, actualBody[key])
					}
				}
			}
		})
	}
}

func TestHTTPTaskHandler_GetTask(t *testing.T) {
	t.Parallel()

	type args struct {
		method string
		url    string
	}

	type setup func(t *testing.T, mockUsecase *mocks.TaskUsecase)

	tests := []struct {
		name       string
		args       args
		setup      setup
		wantStatus int
	}{
		{
			name: "should return 200 when task is found",
			args: args{
				method: http.MethodGet,
				url:    "/api/tasks/1",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("GetTask", mock.Anything, int64(1)).
					Return(&usecases.TaskResult{
						ID:          1,
						Title:       "Shopping",
						Description: "Weekly shopping",
						Items:       []usecases.TaskItemResult{},
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 400 when task ID is invalid",
			args: args{
				method: http.MethodGet,
				url:    "/api/tasks/invalid",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				// Usecase should not be called
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 404 when task is not found",
			args: args{
				method: http.MethodGet,
				url:    "/api/tasks/999",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("GetTask", mock.Anything, int64(999)).
					Return(nil, sql.ErrNoRows).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "should return 500 when usecase returns error",
			args: args{
				method: http.MethodGet,
				url:    "/api/tasks/1",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("GetTask", mock.Anything, int64(1)).
					Return(nil, errors.New("internal error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockUsecase := mocks.NewTaskUsecase(t)
			if tt.setup != nil {
				tt.setup(t, mockUsecase)
			}

			// Setup handler and router
			handler := HTTPHandler{
				httpTaskHandler: NewHTTPTaskHandler(mockUsecase),
			}
			router := gin.Default()
			api := router.Group("/api")
			handler.registerTaskRoutes(api)

			// Create request
			req, err := http.NewRequestWithContext(context.Background(), tt.args.method, tt.args.url, nil)
			require.NoError(t, err)

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestHTTPTaskHandler_ListTasks(t *testing.T) {
	t.Parallel()

	type args struct {
		method string
		url    string
	}

	type setup func(t *testing.T, mockUsecase *mocks.TaskUsecase)

	tests := []struct {
		name       string
		args       args
		setup      setup
		wantStatus int
	}{
		{
			name: "should return 200 with list of tasks",
			args: args{
				method: http.MethodGet,
				url:    "/api/tasks",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("ListTasks", mock.Anything).
					Return(&usecases.TaskListResult{
						Tasks: []usecases.TaskResult{
							{ID: 1, Title: "Shopping", Description: "Weekly shopping", Items: []usecases.TaskItemResult{}},
							{ID: 2, Title: "Work", Description: "Project tasks", Items: []usecases.TaskItemResult{}},
						},
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 200 with empty list when no tasks exist",
			args: args{
				method: http.MethodGet,
				url:    "/api/tasks",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("ListTasks", mock.Anything).
					Return(&usecases.TaskListResult{
						Tasks: []usecases.TaskResult{},
					}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "should return 500 when usecase returns error",
			args: args{
				method: http.MethodGet,
				url:    "/api/tasks",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("ListTasks", mock.Anything).
					Return(nil, errors.New("internal error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockUsecase := mocks.NewTaskUsecase(t)
			if tt.setup != nil {
				tt.setup(t, mockUsecase)
			}

			// Setup handler and router
			handler := HTTPHandler{
				httpTaskHandler: NewHTTPTaskHandler(mockUsecase),
			}
			router := gin.Default()
			api := router.Group("/api")
			handler.registerTaskRoutes(api)

			// Create request
			req, err := http.NewRequestWithContext(context.Background(), tt.args.method, tt.args.url, nil)
			require.NoError(t, err)

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestHTTPTaskHandler_DeleteTask(t *testing.T) {
	t.Parallel()

	type args struct {
		method string
		url    string
	}

	type setup func(t *testing.T, mockUsecase *mocks.TaskUsecase)

	tests := []struct {
		name       string
		args       args
		setup      setup
		wantStatus int
	}{
		{
			name: "should return 204 when task is deleted successfully",
			args: args{
				method: http.MethodDelete,
				url:    "/api/tasks/1",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("DeleteTask", mock.Anything, int64(1)).
					Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "should return 400 when task ID is invalid",
			args: args{
				method: http.MethodDelete,
				url:    "/api/tasks/invalid",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				// Usecase should not be called
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "should return 404 when task is not found",
			args: args{
				method: http.MethodDelete,
				url:    "/api/tasks/999",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("DeleteTask", mock.Anything, int64(999)).
					Return(db.ErrTaskNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "should return 500 when usecase returns error",
			args: args{
				method: http.MethodDelete,
				url:    "/api/tasks/1",
			},
			setup: func(t *testing.T, mockUsecase *mocks.TaskUsecase) {
				mockUsecase.On("DeleteTask", mock.Anything, int64(1)).
					Return(errors.New("internal error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mocks
			mockUsecase := mocks.NewTaskUsecase(t)
			if tt.setup != nil {
				tt.setup(t, mockUsecase)
			}

			// Setup handler and router
			handler := HTTPHandler{
				httpTaskHandler: NewHTTPTaskHandler(mockUsecase),
			}
			router := gin.Default()
			api := router.Group("/api")
			handler.registerTaskRoutes(api)

			// Create request
			req, err := http.NewRequestWithContext(context.Background(), tt.args.method, tt.args.url, nil)
			require.NoError(t, err)

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
