package usecases

// CreateTaskItemParams represents the input for creating a task item
type CreateTaskItemParams struct {
	Title     string
	Completed bool
}

// CreateTaskParams represents the input for creating a task
type CreateTaskParams struct {
	Title       string
	Description string
	Items       []CreateTaskItemParams
}
