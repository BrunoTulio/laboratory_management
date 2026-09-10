package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"laboratory-management/config"
)

var db *sql.DB

// Initialize creates a new database connection
func Initialize() (*sql.DB, error) {
	cfg := config.Load()
	
	var err error
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * 60) // 5 minutes

	log.Println("Database connection established successfully")
	return db, nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// RunMigrations executes the migration files
func RunMigrations() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Read migration file
	migrationFile := "migrations/001_initial_schema.sql"
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("error reading migration file: %w", err)
	}

	// Execute migration
	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("error executing migration: %w", err)
	}

	log.Println("Migrations executed successfully")
	return nil
}