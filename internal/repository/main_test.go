package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbName := "testdb"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}

	// Run migrations
	if err := runMigrations(ctx, testPool); err != nil {
		log.Fatalf("failed to run migrations: %s", err)
	}

	code := m.Run()

	testPool.Close()
	if err := postgresContainer.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Find migrations directory. Since we are in internal/repository, 
	// migrations are at ../../migrations
	migrationsDir := "../../migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	for _, f := range migrationFiles {
		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return err
		}

		// Split by -- +goose Down to only run Up part
		parts := strings.Split(string(content), "-- +goose Down")
		upPart := parts[0]
		// Remove -- +goose Up
		upPart = strings.Replace(upPart, "-- +goose Up", "", 1)

		if _, err := pool.Exec(ctx, upPart); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", f, err)
		}
	}

	return nil
}
