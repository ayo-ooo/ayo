package db

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"

	// ncruces/go-sqlite3 provides a pure-Go SQLite driver using WebAssembly.
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// Connect opens a SQLite database connection and runs migrations.
func Connect(ctx context.Context, dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path is not set")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// ncruces driver uses "sqlite3" as the driver name and requires file: prefix
	db, err := sql.Open("sqlite3", "file:"+dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set WAL mode for better concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations (silently)
	goose.SetBaseFS(Migrations)
	goose.SetLogger(log.New(io.Discard, "", 0))

	if err := goose.SetDialect("sqlite3"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

// ConnectWithQueries opens a database and returns prepared queries.
func ConnectWithQueries(ctx context.Context, dbPath string) (*sql.DB, *Queries, error) {
	db, err := Connect(ctx, dbPath)
	if err != nil {
		return nil, nil, err
	}

	queries, err := Prepare(ctx, db)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to prepare queries: %w", err)
	}

	return db, queries, nil
}
