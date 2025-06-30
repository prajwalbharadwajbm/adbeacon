package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
)

// DB holds the database connection
type DB struct {
	*sql.DB
}

// NewConnection creates a new database connection with connection pooling
func NewConnection(cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// HealthCheck performs a health check on the database connection
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetConnectionStats returns database connection statistics
func (db *DB) GetConnectionStats() sql.DBStats {
	return db.Stats()
}

// RunMigrations runs database migrations
func (db *DB) RunMigrations(migrationsPath string) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Initialize sets up the complete database with connection, migrations, and returns cleanup function
func Initialize(cfg config.DatabaseConfig, migrationsPath string) (*DB, func(), error) {
	// Ensure database exists
	if err := EnsureDatabase(cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to ensure database exists: %w", err)
	}

	// Connect to database
	db, err := NewConnection(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	migrationManager := NewMigrationManager(db, migrationsPath)
	if err := migrationManager.Up(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create cleanup function
	cleanup := func() {
		if err := db.Close(); err != nil {
			// Note: Using fmt.Printf instead of log to avoid importing log package
			// The caller can handle logging as needed
			fmt.Printf("Error closing database connection: %v\n", err)
		}
	}

	// Final health check
	if err := db.HealthCheck(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("database health check failed: %w", err)
	}

	return db, cleanup, nil
}
