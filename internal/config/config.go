package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig `yaml:"db"`
	Server   ServerConfig   `yaml:"server"`
	Log      LogConfig      `yaml:"log"`
}

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	MinConns        int32 `yaml:"minConns"`        // Minimum number of connections to maintain in the pool
	MaxConns        int32 `yaml:"maxConns"`        // Maximum number of connections allowed in the pool
	MaxConnLifetime int   `yaml:"maxConnLifetime"` // Maximum lifetime of a connection in minutes
	MaxConnIdleTime int   `yaml:"maxConnIdleTime"` // Maximum idle time of a connection in minutes
}

// GetMaxConnLifetime returns MaxConnLifetime as time.Duration
func (p *PoolConfig) GetMaxConnLifetime() time.Duration {
	return time.Duration(p.MaxConnLifetime) * time.Minute
}

// GetMaxConnIdleTime returns MaxConnIdleTime as time.Duration
func (p *PoolConfig) GetMaxConnIdleTime() time.Duration {
	return time.Duration(p.MaxConnIdleTime) * time.Minute
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string     `yaml:"host"`
	Port     int        `yaml:"port"`
	User     string     `yaml:"user"`
	Password string     `yaml:"password"`
	Database string     `yaml:"name"`
	SSLMode  string     `yaml:"sslMode"`
	Pool     PoolConfig `yaml:"pool"` // Connection pool settings
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // gin mode: debug, release, test
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Pretty bool   `yaml:"pretty"` // Enable pretty console output
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

// NewDefaultConfig returns a Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Database: "todo_db",
			SSLMode:  "disable",
			Pool: PoolConfig{
				MinConns:        3, // Minimum connections to maintain
				MaxConns:        5, // Maximum connections allowed
				MaxConnLifetime: 5, // 5 minutes
				MaxConnIdleTime: 5, // 5 minutes
			},
		},
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Log: LogConfig{
			Level:  "info",
			Pretty: true,
		},
	}
}

// LoadFromFile loads configuration from a YAML file
// If the file doesn't exist, it returns the default config
func LoadFromFile(path string) (*Config, error) {
	cfg := NewDefaultConfig()

	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// MergeWithFlags merges configuration with values from flags/env vars
// Non-zero/non-empty flag values override config file values
func (c *Config) MergeWithFlags(
	dbHost string, dbPort int, dbUser string, dbPassword string, dbName string, dbSSLMode string,
	poolMinConns int, poolMaxConns int, poolMaxConnLifetime int, poolMaxConnIdleTime int,
	serverPort int, serverMode string,
) {
	if dbHost != "" && dbHost != "localhost" {
		c.Database.Host = dbHost
	}
	if dbPort != 0 && dbPort != 5432 {
		c.Database.Port = dbPort
	}
	if dbUser != "" && dbUser != "postgres" {
		c.Database.User = dbUser
	}
	if dbPassword != "" && dbPassword != "postgres" {
		c.Database.Password = dbPassword
	}
	if dbName != "" && dbName != "todo_db" {
		c.Database.Database = dbName
	}
	if dbSSLMode != "" && dbSSLMode != "disable" {
		c.Database.SSLMode = dbSSLMode
	}
	if poolMinConns != 0 && poolMinConns != 3 {
		c.Database.Pool.MinConns = int32(poolMinConns)
	}
	if poolMaxConns != 0 && poolMaxConns != 5 {
		c.Database.Pool.MaxConns = int32(poolMaxConns)
	}
	if poolMaxConnLifetime != 0 && poolMaxConnLifetime != 5 {
		c.Database.Pool.MaxConnLifetime = poolMaxConnLifetime
	}
	if poolMaxConnIdleTime != 0 && poolMaxConnIdleTime != 5 {
		c.Database.Pool.MaxConnIdleTime = poolMaxConnIdleTime
	}
	if serverPort != 0 && serverPort != 8080 {
		c.Server.Port = serverPort
	}
	if serverMode != "" && serverMode != "debug" {
		c.Server.Mode = serverMode
	}
}
