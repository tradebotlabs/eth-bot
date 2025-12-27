package storage

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog/log"
)

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string // disable, require, verify-ca, verify-full
	MaxConns int
	MaxIdle  int
	ConnMaxLifetime time.Duration
}

// DefaultPostgresConfig returns default PostgreSQL configuration
func DefaultPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "eth_trading",
		SSLMode:         "disable",
		MaxConns:        25,
		MaxIdle:         5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *PostgresConfig) (*sqlx.DB, error) {
	if cfg == nil {
		cfg = DefaultPostgresConfig()
	}

	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Open connection
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.DBName).
		Msg("PostgreSQL connection established")

	return db, nil
}

// MigratePostgres runs database migrations
func MigratePostgres(db *sqlx.DB, schemaPath string) error {
	// Read schema file
	// In production, use a proper migration tool like golang-migrate
	log.Info().Str("schema", schemaPath).Msg("Running database migrations")

	// For now, we'll assume the schema is applied manually
	// TODO: Implement proper migrations with golang-migrate

	return nil
}
