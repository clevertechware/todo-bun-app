package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type taskItemHTTPResponse struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type taskHTTPResponse struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Items       []taskItemHTTPResponse `json:"items"`
}

type taskListHTTPResponse struct {
	Tasks []taskHTTPResponse `json:"tasks"`
}

type validationErrorResponse map[string]string

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		validationErrorResponse := make(validationErrorResponse)
		for _, fe := range ve {
			field := fe.Field()
			jsonField := toJSONFieldName(field)
			validationErrorResponse[jsonField] = getValidationErrorMessage(fe)
		}
		c.JSON(http.StatusBadRequest, validationErrorResponse)
		return
	}

	// Fallback to generic error
	c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
}

func respondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, errorResponse{Error: message})
}

func toJSONFieldName(field string) string {
	if len(field) == 0 {
		return field
	}
	return strings.ToLower(field[0:1]) + field[1:]
}

func getValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required"
	case "email":
		return "invalid email"
	case "min":
		return "too short"
	case "max":
		return "too long"
	case "len":
		return "invalid length"
	case "numeric":
		return "must be numeric"
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	default:
		return "invalid"
	}
}
