.PHONY: help build run migrate-up migrate-down docker-up docker-down clean test

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the application binary
	go build -o todo-app main.go

run: ## Run the application (requires PostgreSQL to be running)
	go run main.go serve

migrate-up: ## Run database migrations up
	go run main.go migrate up

migrate-down: ## Run database migrations down
	go run main.go migrate down

docker-up: ## Start PostgreSQL with Docker Compose
	docker-compose up -d

docker-down: ## Stop PostgreSQL Docker container
	docker-compose down

docker-logs: ## Show PostgreSQL logs
	docker-compose logs -f postgres

setup: docker-up ## Setup the project (start PostgreSQL and run migrations)
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@$(MAKE) migrate-up
	@echo "Setup complete! Run 'make run' to start the server"

clean: ## Clean build artifacts
	rm -f todo-app
	go clean

test: ## Run tests
	go test -v ./...

deps: ## Download dependencies
	go mod download
	go mod tidy

dev: setup run ## Setup and run in development mode
