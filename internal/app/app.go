package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/clevertechware/todo-bun-app/internal/app/db"
	"github.com/clevertechware/todo-bun-app/internal/app/handlers"
	"github.com/clevertechware/todo-bun-app/internal/app/usecases"
	"github.com/clevertechware/todo-bun-app/internal/config"
	"github.com/clevertechware/todo-bun-app/internal/pkg/logger"
)

// App holds all application dependencies
type App struct {
	DB          *bun.DB
	httpHandler *handlers.HTTPHandler
	logger      *zerolog.Logger
}

// NewApp creates a new App instance with all dependencies wired
func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	globalLogger := logger.GetLogger()
	globalLogger.Info().Msg("Initializing database connection")

	// Parse pgxpool configuration
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.GetDSN())
	if err != nil {
		globalLogger.Error().Err(err).Msg("Failed to parse database config")
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool with fine-grained control
	poolConfig.MinConns = cfg.Database.Pool.MinConns
	poolConfig.MaxConns = cfg.Database.Pool.MaxConns
	poolConfig.MaxConnLifetime = cfg.Database.Pool.GetMaxConnLifetime()
	poolConfig.MaxConnIdleTime = cfg.Database.Pool.GetMaxConnIdleTime()

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		globalLogger.Error().Err(err).Msg("Failed to create database pool")
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Validate the connection by acquiring a connection from the pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		globalLogger.Error().Err(err).Msg("Failed to connect to database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	conn.Release()

	// Convert pgxpool to database/sql for Bun compatibility
	sqldb := stdlib.OpenDBFromPool(pool)
	bunDB := bun.NewDB(sqldb, pgdialect.New())
	bunDB.AddQueryHook(
		bundebug.NewQueryHook(
			// disable the hook
			bundebug.WithEnabled(false),

			// BUNDEBUG=1 logs failed queries
			// BUNDEBUG=2 logs all queries
			bundebug.FromEnv("BUNDEBUG"),
		),
	)

	globalLogger.Info().
		Str("host", cfg.Database.Host).
		Int("port", cfg.Database.Port).
		Str("database", cfg.Database.Database).
		Int32("minConns", cfg.Database.Pool.MinConns).
		Int32("maxConns", cfg.Database.Pool.MaxConns).
		Msg("Database connection established with pgxpool")

	// Wire dependencies: Repository -> Usecase -> Handler
	globalLogger.Debug().Msg("Wiring dependencies")
	taskRepo := db.NewTaskRepository(bunDB)
	taskUsecase := usecases.NewTaskUsecase(taskRepo)
	taskHandler := handlers.NewHTTPTaskHandler(taskUsecase)
	httpHandler := handlers.NewHTTPHandler(taskHandler)

	return &App{
		DB:          bunDB,
		httpHandler: httpHandler,
		logger:      globalLogger,
	}, nil
}

// Close closes all application resources
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}

func (a *App) RegisterRoutes(router gin.IRouter) {
	a.httpHandler.RegisterRoutes(router)
}
