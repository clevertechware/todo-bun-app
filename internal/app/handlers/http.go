package handlers

import "github.com/gin-gonic/gin"

type HTTPHandler struct {
	httpTaskHandler *HTTPTaskHandler
}

func NewHTTPHandler(httpTaskHandler *HTTPTaskHandler) *HTTPHandler {
	return &HTTPHandler{httpTaskHandler: httpTaskHandler}
}

func (h *HTTPHandler) RegisterRoutes(router gin.IRouter) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	h.registerTaskRoutes(api)
}

func (h *HTTPHandler) registerTaskRoutes(api gin.IRouter) {
	tasks := api.Group("/tasks")
	{
		tasks.POST("", h.httpTaskHandler.CreateTask)
		tasks.GET("", h.httpTaskHandler.ListTasks)
		tasks.GET("/:id", h.httpTaskHandler.GetTask)
		tasks.DELETE("/:id", h.httpTaskHandler.DeleteTask)
	}
}
