package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db            *DB
	migrationsDir string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DB, migrationsDir string) *MigrationManager {
	return &MigrationManager{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// Up runs all up migrations
func (m *MigrationManager) Up() error {
	migration, err := m.createMigrationInstance()
	if err != nil {
		return err
	}
	defer migration.Close()

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run up migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Down runs all down migrations
func (m *MigrationManager) Down() error {
	migration, err := m.createMigrationInstance()
	if err != nil {
		return err
	}
	defer migration.Close()

	if err := migration.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run down migrations: %w", err)
	}

	log.Println("Database down migrations completed successfully")
	return nil
}

// Reset drops all tables and re-runs migrations
func (m *MigrationManager) Reset() error {
	log.Println("Resetting database...")

	if err := m.Down(); err != nil {
		return fmt.Errorf("failed to run down migrations during reset: %w", err)
	}

	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to run up migrations during reset: %w", err)
	}

	log.Println("Database reset completed successfully")
	return nil
}

// Version returns current migration version
func (m *MigrationManager) Version() (uint, bool, error) {
	migration, err := m.createMigrationInstance()
	if err != nil {
		return 0, false, err
	}
	defer migration.Close()

	return migration.Version()
}

// Force sets the migration version without running migrations
func (m *MigrationManager) Force(version int) error {
	migration, err := m.createMigrationInstance()
	if err != nil {
		return err
	}
	defer migration.Close()

	return migration.Force(version)
}

// createMigrationInstance creates a new migration instance
func (m *MigrationManager) createMigrationInstance() (*migrate.Migrate, error) {
	// Create a separate connection for migrations to avoid closing the main connection
	cfg := config.AppConfigInstance.DatabaseConfig
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	migrationDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open migration database connection: %w", err)
	}

	driver, err := postgres.WithInstance(migrationDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath, err := filepath.Abs(m.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return migration, nil
}

// EnsureDatabase creates the database if it doesn't exist
func EnsureDatabase(cfg config.DatabaseConfig) error {
	// Connect to postgres database to create the target database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"
	err = db.QueryRow(query, cfg.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		log.Printf("Creating database: %s", cfg.DBName)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database %s created successfully", cfg.DBName)
	} else {
		log.Printf("Database %s already exists", cfg.DBName)
	}

	return nil
}
