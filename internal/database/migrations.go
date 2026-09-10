package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes pending SQL migrations
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Create schema_migrations table if not exists
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read migration files from directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		// Fallback: run embedded default migration if directory not found
		log.Printf("Migrations directory '%s' not found (%v). Running built-in schema fallback...", migrationsDir, err)
		return runDefaultSchema(db)
	}

	var upFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	if len(upFiles) == 0 {
		log.Println("No .up.sql migration files found, running built-in schema fallback...")
		return runDefaultSchema(db)
	}

	for _, filename := range upFiles {
		version := strings.Split(filename, "_")[0]

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration version %s: %w", version, err)
		}

		if exists {
			continue
		}

		fullPath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("Successfully applied migration: %s", filename)
	}

	return nil
}

func runDefaultSchema(db *sql.DB) error {
	defaultSQL := `
	CREATE TABLE IF NOT EXISTS users (
		telegram_id BIGINT PRIMARY KEY,
		username VARCHAR(255),
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS submissions (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
		topic TEXT,
		essay_text TEXT NOT NULL,
		overall_score INT,
		tr_score INT,
		cc_score INT,
		ga_score INT,
		lr_score INT,
		cefr_level VARCHAR(10),
		feedback TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_submissions_user_id ON submissions(user_id);
	CREATE INDEX IF NOT EXISTS idx_submissions_created_at ON submissions(created_at DESC);
	`
	_, err := db.Exec(defaultSQL)
	if err != nil {
		return fmt.Errorf("failed to execute default schema: %w", err)
	}
	log.Println("Default schema ensured successfully.")
	return nil
}
