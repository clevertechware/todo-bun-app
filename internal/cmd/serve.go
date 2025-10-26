package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/clevertechware/todo-bun-app/internal/app"
	"github.com/clevertechware/todo-bun-app/internal/config"
	"github.com/clevertechware/todo-bun-app/internal/pkg/logger"
)

// ServeCommand returns the serve command for running the HTTP server
func ServeCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Start the HTTP server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to configuration file (YAML)",
				Sources: cli.EnvVars("CONFIG_FILE"),
				Value:   "config.yaml",
			},
			&cli.StringFlag{
				Name:    "db-host",
				Usage:   "Database host",
				Sources: cli.EnvVars("DB_HOST"),
				Value:   "localhost",
			},
			&cli.IntFlag{
				Name:    "db-port",
				Usage:   "Database port",
				Sources: cli.EnvVars("DB_PORT"),
				Value:   5432,
			},
			&cli.StringFlag{
				Name:    "db-user",
				Usage:   "Database user",
				Sources: cli.EnvVars("DB_USER"),
				Value:   "postgres",
			},
			&cli.StringFlag{
				Name:    "db-password",
				Usage:   "Database password",
				Sources: cli.EnvVars("DB_PASSWORD"),
				Value:   "postgres",
			},
			&cli.StringFlag{
				Name:    "db-name",
				Usage:   "Database name",
				Sources: cli.EnvVars("DB_NAME"),
				Value:   "todo_db",
			},
			&cli.StringFlag{
				Name:    "db-sslmode",
				Usage:   "Database SSL mode",
				Sources: cli.EnvVars("DB_SSLMODE"),
				Value:   "disable",
			},
			&cli.IntFlag{
				Name:    "db-pool-min-conns",
				Usage:   "Minimum number of connections to maintain in the pool",
				Sources: cli.EnvVars("DB_POOL_MIN_CONNS"),
				Value:   3,
			},
			&cli.IntFlag{
				Name:    "db-pool-max-conns",
				Usage:   "Maximum number of connections allowed in the pool",
				Sources: cli.EnvVars("DB_POOL_MAX_CONNS"),
				Value:   5,
			},
			&cli.IntFlag{
				Name:    "db-pool-max-conn-lifetime",
				Usage:   "Maximum lifetime of a connection in minutes",
				Sources: cli.EnvVars("DB_POOL_MAX_CONN_LIFETIME"),
				Value:   5,
			},
			&cli.IntFlag{
				Name:    "db-pool-max-conn-idle-time",
				Usage:   "Maximum idle time of a connection in minutes",
				Sources: cli.EnvVars("DB_POOL_MAX_CONN_IDLE_TIME"),
				Value:   5,
			},
			&cli.IntFlag{
				Name:    "server-port",
				Usage:   "HTTP server port",
				Sources: cli.EnvVars("SERVER_PORT"),
				Value:   8080,
			},
			&cli.StringFlag{
				Name:    "server-mode",
				Usage:   "Server mode (debug, release, test)",
				Sources: cli.EnvVars("SERVER_MODE"),
				Value:   "debug",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Load configuration from YAML file
			cfg, err := config.LoadFromFile(cmd.String("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Merge with flags/env vars (flags override config file)
			cfg.MergeWithFlags(
				cmd.String("db-host"),
				cmd.Int("db-port"),
				cmd.String("db-user"),
				cmd.String("db-password"),
				cmd.String("db-name"),
				cmd.String("db-sslmode"),
				cmd.Int("db-pool-min-conns"),
				cmd.Int("db-pool-max-conns"),
				cmd.Int("db-pool-max-conn-lifetime"),
				cmd.Int("db-pool-max-conn-idle-time"),
				cmd.Int("server-port"),
				cmd.String("server-mode"),
			)

			// Initialize logger
			logger.Init(logger.Config{
				Level:  cfg.Log.Level,
				Pretty: cfg.Log.Pretty,
			})

			log.Info().Msg("Starting application")

			// Run database migrations
			log.Info().Msg("Running database migrations")
			if err = RunMigrationUp(&cfg.Database); err != nil {
				log.Error().Err(err).Msg("Failed to run migrations")
				return fmt.Errorf("failed to run migrations: %w", err)
			}
			log.Info().Msg("Database migrations completed")

			// Set Gin mode
			gin.SetMode(cfg.Server.Mode)

			// Initialize application
			application, err := app.NewApp(ctx, cfg)
			if err != nil {
				log.Error().Err(err).Msg("Failed to initialize application")
				return fmt.Errorf("failed to initialize application: %w", err)
			}
			defer func(application *app.App) {
				closeAppErr := application.Close()
				if closeAppErr != nil {
					log.Fatal().Err(closeAppErr).Msg("Failed to close application")
				}
			}(application)

			log.Info().Msg("Application initialized successfully")

			// Setup routes
			router := setupRoutes(application)

			// Start server
			addr := fmt.Sprintf(":%d", cfg.Server.Port)
			log.Info().Str("address", addr).Msg("Starting HTTP server")
			return router.Run(addr)
		},
	}
}

// setupRoutes configures all HTTP routes
func setupRoutes(app *app.App) *gin.Engine {
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add zerolog middleware
	router.Use(zerologMiddleware())
	app.RegisterRoutes(router)

	return router
}

// zerologMiddleware logs HTTP requests using zerolog
func zerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Str("ip", c.ClientIP()).
			Str("user-agent", c.Request.UserAgent()).
			Msg("HTTP request")
	}
}
