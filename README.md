# TASK Application with Bun ORM

A TASK application built with Golang, demonstrating the use of Bun ORM with Postgres.

## Stack

- **Urfave CLI v3**: Command-line interface and configuration management
- **Gin**: HTTP web framework
- **Bun**: SQL-first Golang ORM for PostgreSQL
- **pgx/v5**: High-performance PostgreSQL driver
- **Go Migrate**: Database migration management
- **PostgreSQL**: Database
- **ZeroLog**: Structured logging library

## Architecture

The project follows the following architecture:

```
handlers (HTTP) → usecases (Business Logic) → repositories (Data Access) → database
```

### Project Structure

```
.
├── main.go                          # Application entry point
├── .mockery.yml                     # Mockery configuration for mock generation
├── migrations/                      # Database migrations
├── internal/
│   ├── app/
│   │   ├── app.go                  # Dependency injection container
│   │   ├── db/                     # Repository layer
│   │   │   ├── pg_task.go          # Task repository implementation
│   │   │   ├── pg_task_test.go     # Repository integration tests
│   │   │   ├── suite_pg_test.go    # Test suite setup
│   │   │   ├── errors.go           # Repository errors
│   │   │   └── mocks/              # Generated mocks
│   │   │       └── task_repository.go
│   │   ├── handlers/               # HTTP handlers (Gin)
│   │   │   ├── http_task_handler.go      # HTTP handlers
│   │   │   ├── http_task_handler_test.go # Handler unit tests
│   │   │   ├── http_request.go     # HTTP request DTOs
│   │   │   ├── http_response.go    # HTTP response DTOs
│   │   │   ├── errors.go           # Validation error handling
│   │   │   └── main_test.go        # Test configuration
│   │   ├── models/                 # Domain models
│   │   │   ├── task.go
│   │   │   └── task_item.go
│   │   └── usecases/               # Business logic
│   │       ├── task_usecase.go     # Task business logic
│   │       ├── task_usecase_test.go# Usecase unit tests
│   │       ├── task_params.go      # Input DTOs
│   │       ├── task_result.go      # Output DTOs
│   │       └── mocks/              # Generated mocks
│   │           └── task_usecase.go
│   ├── cmd/                        # CLI commands
│   │   ├── serve.go               # HTTP server command
│   │   └── migrate.go             # Migration commands
│   ├── config/                     # Configuration
│   │   └── config.go              # Config structures & YAML loading
│   └── pkg/
│       ├── logger/                 # Logging utilities
│       │   └── logger.go          # ZeroLog wrapper
│       └── testing/                # Test utilities
│           └── testcontainer.go   # PostgreSQL testcontainer setup
```

## Prerequisites

- Go 1.25+
- Postgres 18+
- Docker (optional, for running Postgres)

## Setup

### 1. Start PostgreSQL

Using Docker:

```bash
docker run --name task-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=task_db \
  -p 5432:5432 \
  -d postgres:18-alpine
```

### 2. Start the Server

```bash
go run main.go serve
```

The server will start on `http://localhost:8080`

**Note:** Database migrations are run automatically on startup, so you don't need to run them manually.

### 3. Manual Migration Management (Optional)

If you need to run migrations manually or rollback:

Run migrations:
```bash
go run main.go migrate up
```

Rollback migrations:
```bash
go run main.go migrate down
```

## Configuration

Configuration can be provided via YAML file, environment variables, or command-line flags (in order of precedence: flags > env vars > YAML file > defaults).

### YAML Configuration File

Create a `config.yaml` file (see `config.yaml.example`):

```yaml
db:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: task_db
  sslMode: disable
  # Connection pool settings (using pgxpool for fine-grained control)
  pool:
    minConns: 3              # Minimum connections to maintain (default: 3)
    maxConns: 5              # Maximum connections allowed (default: 5)
    maxConnLifetime: 5       # Connection max lifetime in minutes (default: 5)
    maxConnIdleTime: 5       # Connection max idle time in minutes (default: 5)

server:
  port: 8080
  mode: debug  # Options: debug, release, test

log:
  level: info   # Options: debug, info, warn, error
  pretty: true  # Enable pretty console output
```

Then run with:
```bash
go run main.go serve --config config.yaml
# or
go run main.go serve -c config.yaml
```

### Environment Variables

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=task_db
export DB_SSLMODE=disable
export DB_POOL_MIN_CONNS=3
export DB_POOL_MAX_CONNS=5
export DB_POOL_MAX_CONN_LIFETIME=5
export DB_POOL_MAX_CONN_IDLE_TIME=5
export SERVER_PORT=8080
export SERVER_MODE=debug  # debug, release, or test
```

### Command-line Flags

```bash
go run main.go serve \
  --db-host=localhost \
  --db-port=5432 \
  --db-user=postgres \
  --db-password=postgres \
  --db-name=task_db \
  --db-pool-min-conns=3 \
  --db-pool-max-conns=5 \
  --db-pool-max-conn-lifetime=5 \
  --db-pool-max-conn-idle-time=5 \
  --server-port=8080 \
  --server-mode=debug
```

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

### Create a TASK with items

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Shopping List",
    "description": "Weekly grocery shopping",
    "items": [
      {
        "title": "Buy milk",
        "completed": false
      },
      {
        "title": "Buy bread",
        "completed": false
      },
      {
        "title": "Buy eggs",
        "completed": true
      }
    ]
  }'
```

### Get all TASKs

```bash
curl http://localhost:8080/api/tasks
```

### Get a specific TASK

```bash
curl http://localhost:8080/api/tasks/1
```

### Delete a TASK

```bash
curl -X DELETE http://localhost:8080/api/tasks/1
```

### Validation Errors

The API returns clean validation error messages:

**Request without required field:**
```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"description": "Missing title"}'
```

**Response:**
```json
{
  "title": "required"
}
```

**Global errors:**
```json
{
  "error": "task not found"
}
```

## Key Features Demonstrated

### 1. Bun ORM Usage

- **Relations**: One-to-Many relationship between `Task` and `TaskItem`
- **Transactions**: Atomic creation of tasks with items
- **Query Builder**: Type-safe query construction
- **Cascade Delete**: Automatic deletion of related items

### 2. Clean Architecture

- **Separation of Concerns**: Clear boundaries between layers
- **Dependency Injection**: Manual DI in `app.go`
- **DTO Pattern**: Separate DTOs for each layer
  - `XXXParams`: Input DTOs (usecases)
  - `XXXResult`: Output DTOs (usecases)
  - `XXXHTTPRequest`: HTTP request DTOs (handlers)
  - `XXXHTTPResponse`: HTTP response DTOs (handlers)

### 3. Database Migrations

- Versioned migrations with Go Migrate
- Up and down migrations support
- Automatic schema management

### 4. YAML Configuration

- Load configuration from `config.yaml` file
- Override with environment variables or CLI flags
- Clear precedence order: CLI flags > env vars > YAML > defaults
- Easy to manage different environments

### 5. Structured Logging with ZeroLog

- High-performance structured logging
- JSON or pretty console output
- Configurable log levels (debug, info, warn, error)
- Request logging middleware with duration tracking
- Context-aware logging throughout the application

### 6. Input Validation

- Field-level validation with clean error messages
- Returns JSON object with field names and validation errors
- Example: `{"title": "required"}` instead of verbose error strings
- Built with go-playground/validator

### 7. Database Connection Pooling with pgxpool

- **pgxpool**: Native connection pool manager for pgx with fine-grained control
- **Connection Pool Configuration**:
  - `minConns` (default: 3): Minimum connections always maintained in the pool for fast request handling
  - `maxConns` (default: 5): Maximum connections allowed in the pool
  - `maxConnLifetime` (default: 5 min): Maximum connection lifetime to prevent stale connections
  - `maxConnIdleTime` (default: 5 min): Maximum idle time before closing unused connections
- **PostgreSQL Resource Awareness**: Conservative defaults respect that each PostgreSQL connection creates a backend process
- **Fine-Grained Control**: Using `pgxpool.Config` allows precise pool management beyond standard `database/sql`
- **Connection Validation**: Pool validates connections on acquisition, ensuring reliability
- **Production Ready**: Industry-standard approach used in high-performance Go applications

## Building for Production

```bash
# Build binary
go build -o task-app main.go

# Run migrations
./task-app migrate up

# Start server
./task-app serve --server-mode=release
```

## Development

### Testing

The project includes comprehensive tests at all layers:
- **Unit Tests**: Handler and usecase tests using Mockery-generated mocks
- **Integration Tests**: Repository tests using testcontainers with real PostgreSQL

#### Prerequisites for Testing

- Go 1.21+
- Docker (required for integration tests with testcontainers)
- Mockery v3 (for regenerating mocks)

Install Mockery:
```bash
go install github.com/vektra/mockery/v3@latest
```

#### Test Structure

Tests follow table-driven testing patterns with:
- Parallel execution using `t.Parallel()` for faster test runs
- Descriptive test case names
- Clear separation of test setup, execution, and assertion
- Transaction isolation for integration tests (with automatic rollback)

#### Running All Tests

```bash
# Run all tests (unit + integration)
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...
```

#### Running Unit Tests Only

Unit tests are fast and don't require Docker:

```bash
# Handler tests
go test ./internal/app/handlers/

# Usecase tests
go test ./internal/app/usecases/
```

#### Running Integration Tests

Integration tests use testcontainers and require Docker to be running:

```bash
# Repository integration tests
go test ./internal/app/db/

# Run with verbose output to see container startup
go test -v ./internal/app/db/
```

**Note**: First run may take longer as Docker pulls the PostgreSQL image.

#### Test Examples

**Unit Test Example (Usecase Layer)**:
```go
func TestTaskUsecase_CreateTask(t *testing.T) {
    t.Parallel()

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
                        return task.Title == "Buy groceries"
                    })).Return(nil)
                    return m
                },
            },
            // ... test case details
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            u := &taskUsecase{
                taskRepo: tt.fields.taskRepo(t),
            }
            got, err := u.CreateTask(tt.args.ctx, tt.args.params)
            // assertions...
        })
    }
}
```

**Integration Test Example (Repository Layer)**:
```go
func (s *PGRepositorySuite) TestPGTask_Create() {
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
                    },
                },
            },
            check: func(t *testing.T, client bun.IDB, err error, task *models.Task) {
                require.NoError(t, err)
                assert.NotZero(t, task.ID)
                assert.Equal(t, "Shopping", task.Title)
                // Verify items were created in database
            },
            wantErr: assert.NoError,
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            t := s.T()

            // Start transaction for isolation
            trx, err := s.pgContainer.TxBegin()
            require.NoError(t, err)
            defer func() {
                require.NoError(t, trx.Rollback())
            }()

            repo := NewTaskRepository(trx)
            err = repo.Create(tt.args.ctx, tt.args.task)

            tt.wantErr(t, err)
            if tt.check != nil {
                tt.check(t, trx, err, tt.args.task)
            }
        })
    }
}
```

#### Mock Generation

Mocks are auto-generated from interfaces using Mockery v3. Configuration is in `.mockery.yml`.

To regenerate mocks after interface changes:

```bash
# Generate all mocks
mockery

# Generate mocks for a specific package
mockery --dir internal/app/db --name TaskRepository
```

Generated mocks are placed in `mocks/` subdirectories next to their interfaces:
- `internal/app/db/mocks/task_repository.go`
- `internal/app/usecases/mocks/task_usecase.go`

#### Test Coverage

View test coverage:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

#### CI/CD Considerations

For continuous integration, ensure:
1. Docker is available for integration tests
2. PostgreSQL testcontainer can pull images
3. Tests run with sufficient timeout for container startup:
   ```bash
   go test -timeout 5m ./...
   ```

### Code Generation

If you modify the models, ensure Bun tags are properly set for:
- Table names: `bun:"table:table_name,alias:alias"`
- Primary keys: `bun:"id,pk,autoincrement"`
- Relations: `bun:"rel:has-many,join:id=foreign_key"`

If you add or modify interfaces:
- Run `mockery` to regenerate mocks
- Ensure `.mockery.yml` includes the new package if needed

## License

MIT
