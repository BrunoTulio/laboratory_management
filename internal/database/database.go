package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"laboratory-management/config"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// Initialize creates a new embedded SQLite database connection.
func Initialize() (*sql.DB, error) {
	cfg := config.Load()

	var err error
	db, err = sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("error opening SQLite database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to SQLite database: %w", err)
	}

	// Single writer connection prevents SQLITE_BUSY under concurrent HTTP requests.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;`); err != nil {
		return nil, fmt.Errorf("error configuring SQLite pragmas: %w", err)
	}

	log.Printf("Embedded SQLite database ready at %s", cfg.DatabasePath)
	return db, nil
}

// GetDB returns the database instance.
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection.
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// RunMigrations executes all SQL migration files in alphabetical order.
func RunMigrations() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("error listing migration files: %w", err)
	}
	sort.Strings(files)

	for _, migrationFile := range files {
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %w", migrationFile, err)
		}

		if _, err = db.Exec(string(content)); err != nil {
			return fmt.Errorf("error executing migration %s: %w", migrationFile, err)
		}
	}

	// Idempotent column additions for existing SQLite databases
	additions := []struct {
		table, column, def string
	}{
		{"users", "classes_csv", "TEXT DEFAULT ''"},
		{"users", "subjects_csv", "TEXT DEFAULT ''"},
		{"bookings", "class_name", "TEXT DEFAULT ''"},
	}
	for _, a := range additions {
		if err := ensureColumn(db, a.table, a.column, a.def); err != nil {
			return err
		}
	}

	log.Printf("Migrations executed successfully (%d files)", len(files))
	return nil
}

func ensureColumn(db *sql.DB, table, column, def string) error {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("error inspecting table %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}

	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, def)); err != nil {
		return fmt.Errorf("error adding column %s.%s: %w", table, column, err)
	}
	return nil
}
