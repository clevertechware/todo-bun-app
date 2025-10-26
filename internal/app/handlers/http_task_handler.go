package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/clevertechware/todo-bun-app/internal/app/db"
	"github.com/clevertechware/todo-bun-app/internal/app/usecases"
)

// HTTPTaskHandler handles HTTP requests for tasks
type HTTPTaskHandler struct {
	taskUsecase usecases.TaskUsecase
}

// NewHTTPTaskHandler creates a new HTTPTaskHandler instance
func NewHTTPTaskHandler(taskUsecase usecases.TaskUsecase) *HTTPTaskHandler {
	return &HTTPTaskHandler{
		taskUsecase: taskUsecase,
	}
}

// CreateTask handles POST /api/tasks
func (h *HTTPTaskHandler) CreateTask(c *gin.Context) {
	var req createTaskHTTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithValidationError(c, err)
		return
	}

	// Map HTTP request to usecase params
	params := h.requestToParams(req)

	// Call usecase
	result, err := h.taskUsecase.CreateTask(c.Request.Context(), params)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Map usecase result to HTTP response
	response := h.resultToResponse(result)

	c.JSON(http.StatusCreated, response)
}

// GetTask handles GET /api/tasks/:id
func (h *HTTPTaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "invalid task ID")
		return
	}

	// Call usecase
	result, err := h.taskUsecase.GetTask(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(c, http.StatusNotFound, "task not found")
			return
		}
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Map usecase result to HTTP response
	response := h.resultToResponse(result)

	c.JSON(http.StatusOK, response)
}

// ListTasks handles GET /api/tasks
func (h *HTTPTaskHandler) ListTasks(c *gin.Context) {
	// Call usecase
	result, err := h.taskUsecase.ListTasks(c.Request.Context())
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Map usecase result to HTTP response
	response := h.listResultToResponse(result)

	c.JSON(http.StatusOK, response)
}

// DeleteTask handles DELETE /api/tasks/:id
func (h *HTTPTaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "invalid task ID")
		return
	}

	// Call usecase
	err = h.taskUsecase.DeleteTask(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			respondWithError(c, http.StatusNotFound, "task not found")
			return
		}
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// requestToParams maps HTTP request to usecase params
func (h *HTTPTaskHandler) requestToParams(req createTaskHTTPRequest) usecases.CreateTaskParams {
	items := make([]usecases.CreateTaskItemParams, 0, len(req.Items))
	for _, itemReq := range req.Items {
		items = append(items, usecases.CreateTaskItemParams{
			Title:     itemReq.Title,
			Completed: itemReq.Completed,
		})
	}

	return usecases.CreateTaskParams{
		Title:       req.Title,
		Description: req.Description,
		Items:       items,
	}
}

// resultToResponse maps usecase result to HTTP response
func (h *HTTPTaskHandler) resultToResponse(result *usecases.TaskResult) *taskHTTPResponse {
	items := make([]taskItemHTTPResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, taskItemHTTPResponse{
			ID:        item.ID,
			TaskID:    item.TaskID,
			Title:     item.Title,
			Completed: item.Completed,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return &taskHTTPResponse{
		ID:          result.ID,
		Title:       result.Title,
		Description: result.Description,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
		Items:       items,
	}
}

// listResultToResponse maps usecase list result to HTTP response
func (h *HTTPTaskHandler) listResultToResponse(result *usecases.TaskListResult) *taskListHTTPResponse {
	tasks := make([]taskHTTPResponse, 0, len(result.Tasks))
	for _, task := range result.Tasks {
		tasks = append(tasks, *h.resultToResponse(&task))
	}

	return &taskListHTTPResponse{
		Tasks: tasks,
	}
}
