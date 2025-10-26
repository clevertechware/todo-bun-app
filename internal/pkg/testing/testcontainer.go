package testing

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// PostgresTestDatabase provides a PostgreSQL test database using testcontainers
type PostgresTestDatabase struct {
	container *postgres.PostgresContainer
	DB        *bun.DB
}

// NewPostgresDatabase creates a new PostgreSQL test database container
func NewPostgresDatabase() *PostgresTestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	container, err := createPostgresTestContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup postgres container:", err)
	}

	bunDB, err := createBunDB(ctx, container)
	if err != nil {
		log.Fatal("failed to setup bun db:", err)
	}

	return &PostgresTestDatabase{
		container: container,
		DB:        bunDB,
	}
}

func createPostgresTestContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"docker.io/postgres:18-alpine",
		postgres.WithDatabase("todo_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	return pgContainer, nil
}

func createBunDB(ctx context.Context, container *postgres.PostgresContainer) (*bun.DB, error) {
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connStr)))
	bunDB := bun.NewDB(sqldb, pgdialect.New())

	// Test connection
	if err = bunDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Obtenir la racine du projet et construire le chemin des migrations
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get project root: %w", err)
	}

	migrationPath := fmt.Sprintf("file://%s/migrations", projectRoot)

	// Run migrations
	m, err := migrate.New(migrationPath, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer func(m *migrate.Migrate) {
		closeErr, _ := m.Close()
		if closeErr != nil {
			log.Fatal(closeErr)
		}
	}(m)
	if err = m.Up(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return bunDB, nil
}

func getProjectRoot() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Remonte jusqu'à trouver go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("impossible de trouver la racine du projet (go.mod introuvable)")
}

// TxBegin starts a new transaction
func (testContainer *PostgresTestDatabase) TxBegin() (bun.Tx, error) {
	return testContainer.DB.Begin()
}

// TearDown terminates the container
func (testContainer *PostgresTestDatabase) TearDown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := testContainer.DB.Close(); err != nil {
		log.Printf("failed to close database: %v", err)
	}

	return testContainer.container.Terminate(ctx)
}
