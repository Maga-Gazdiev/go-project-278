package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const (
	DefaultMigrationsDir = "db/migrations"
	DefaultDialect       = "postgres"
)

type MigrationOptions struct {
	DatabaseURL   string
	MigrationsDir string
	Dialect       string
}

func DefaultMigrationOptions() *MigrationOptions {
	_ = godotenv.Load()

	return &MigrationOptions{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		MigrationsDir: DefaultMigrationsDir,
		Dialect:       DefaultDialect,
	}
}

func openDB(opts *MigrationOptions) (*sql.DB, error) {
	if opts.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if err := goose.SetDialect(opts.Dialect); err != nil {
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	db, err := sql.Open(opts.Dialect, opts.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	if _, err := goose.EnsureDBVersion(db); err != nil {
		fmt.Printf("Warning: failed to ensure goose_db_version table (will be created by first migration): %v\n", err)
	}

	return db, nil
}

func MigrateUp(opts *MigrationOptions) error {
	if opts == nil {
		opts = DefaultMigrationOptions()
	}

	db, err := openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.Up(db, opts.MigrationsDir)
}

func MigrateDown(opts *MigrationOptions) error {
	if opts == nil {
		opts = DefaultMigrationOptions()
	}

	db, err := openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.Down(db, opts.MigrationsDir)
}

func MigrateStatus(opts *MigrationOptions) error {
	if opts == nil {
		opts = DefaultMigrationOptions()
	}

	db, err := openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.Status(db, opts.MigrationsDir)
}

func MigrateReset(opts *MigrationOptions) error {
	if opts == nil {
		opts = DefaultMigrationOptions()
	}

	db, err := openDB(opts)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.Reset(db, opts.MigrationsDir)
}
