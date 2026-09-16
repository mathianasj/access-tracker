package db

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations, err := getMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	for _, migration := range migrations {
		applied, err := isMigrationApplied(ctx, pool, migration.name)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", migration.name, err)
		}

		if applied {
			continue
		}

		fmt.Printf("Applying migration: %s\n", migration.name)
		_, err = pool.Exec(ctx, migration.sql)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.name, err)
		}

		err = recordMigration(ctx, pool, migration.name)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.name, err)
		}
		fmt.Printf("Applied migration: %s\n", migration.name)
	}

	return nil
}

type migration struct {
	name string
	sql  string
}

func getMigrations() ([]migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		sql, err := migrationsFS.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration{
			name: strings.TrimSuffix(entry.Name(), ".sql"),
			sql:  string(sql),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].name < migrations[j].name
	})

	return migrations, nil
}

func isMigrationApplied(ctx context.Context, pool *pgxpool.Pool, name string) (bool, error) {
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE name = $1", name).Scan(&count)
	if err != nil {
		if strings.Contains(err.Error(), `relation "schema_migrations" does not exist`) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func recordMigration(ctx context.Context, pool *pgxpool.Pool, name string) error {
	_, err := pool.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", name)
	if err != nil {
		if strings.Contains(err.Error(), `relation "schema_migrations" does not exist`) {
			_, err = pool.Exec(ctx, `CREATE TABLE schema_migrations (name VARCHAR(255) PRIMARY KEY)`)
			if err != nil {
				return err
			}
			_, err = pool.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", name)
		}
	}
	return err
}

func EnsureSchemaMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name VARCHAR(255) PRIMARY KEY)`)
	return err
}

func RunMigrationsOnStartup(ctx context.Context) error {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/access_tracker?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	if err := RunMigrations(ctx, pool); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
