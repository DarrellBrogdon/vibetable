package migrate

import (
	"context"
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var migrations embed.FS

// Run executes all pending migrations in order
func Run(ctx context.Context, db *pgxpool.Pool) error {
	// Create migrations tracking table if it doesn't exist
	if err := createMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files
	entries, err := migrations.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// Execute pending migrations
	for _, file := range files {
		name := strings.TrimSuffix(file, filepath.Ext(file))

		if applied[name] {
			log.Printf("Migration already applied: %s", name)
			continue
		}

		log.Printf("Applying migration: %s", name)

		content, err := migrations.ReadFile("sql/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		// Execute migration in a transaction
		tx, err := db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}

		// Record migration as applied
		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (name) VALUES ($1)",
			name,
		); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", name, err)
		}

		log.Printf("Migration applied successfully: %s", name)
	}

	return nil
}

func createMigrationsTable(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	return err
}

func getAppliedMigrations(ctx context.Context, db *pgxpool.Pool) (map[string]bool, error) {
	rows, err := db.Query(ctx, "SELECT name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, rows.Err()
}
