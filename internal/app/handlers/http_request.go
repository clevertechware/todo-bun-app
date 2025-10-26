package handlers

type createTaskItemHTTPRequest struct {
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

type createTaskHTTPRequest struct {
	Title       string                      `json:"title" binding:"required"`
	Description string                      `json:"description"`
	Items       []createTaskItemHTTPRequest `json:"items"`
}
