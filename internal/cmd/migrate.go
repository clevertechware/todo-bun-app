package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/urfave/cli/v3"

	"github.com/clevertechware/todo-bun-app/internal/config"
)

// MigrateCommand returns the migrate command for database migrations
func MigrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Database migration commands",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Run all up migrations",
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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg := buildDBConfigFromYAML(cmd)
					return runMigrationUp(&cfg.Database)
				},
			},
			{
				Name:  "down",
				Usage: "Run all down migrations",
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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg := buildDBConfigFromYAML(cmd)
					return runMigrationDown(&cfg.Database)
				},
			},
		},
	}
}

func buildDBConfigFromYAML(cmd *cli.Command) *config.Config {
	// Load configuration from YAML file
	cfg, err := config.LoadFromFile(cmd.String("config"))
	if err != nil {
		log.Printf("Warning: failed to load config file, using defaults: %v", err)
		cfg = config.NewDefaultConfig()
	}

	// Merge with flags/env vars (flags override config file)
	cfg.MergeWithFlags(
		cmd.String("db-host"),
		cmd.Int("db-port"),
		cmd.String("db-user"),
		cmd.String("db-password"),
		cmd.String("db-name"),
		cmd.String("db-sslmode"),
		0,  // pool settings not used in migrate - use defaults
		0,  // pool settings not used in migrate - use defaults
		0,  // pool settings not used in migrate - use defaults
		0,  // pool settings not used in migrate - use defaults
		0,  // server port not used in migrate
		"", // server mode not used in migrate
	)

	return cfg
}

// RunMigrationUp runs all up migrations (exported for use in serve command)
func RunMigrationUp(cfg *config.DatabaseConfig) error {
	m, err := migrate.New(
		"file://migrations",
		cfg.GetDSN(),
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer func(m *migrate.Migrate) {
		closeErr, _ := m.Close()
		if closeErr != nil {
			log.Fatal(closeErr)
		}
	}(m)

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

// runMigrationUp is a wrapper for backward compatibility
func runMigrationUp(cfg *config.DatabaseConfig) error {
	return RunMigrationUp(cfg)
}

func runMigrationDown(cfg *config.DatabaseConfig) error {
	m, err := migrate.New(
		"file://migrations",
		cfg.GetDSN(),
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer func(m *migrate.Migrate) {
		closeErr, _ := m.Close()
		if closeErr != nil {
			log.Fatal(closeErr)
		}
	}(m)

	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	log.Println("Migrations rolled back successfully")
	return nil
}
