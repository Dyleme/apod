package repository

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func migrateUp(db *sql.DB) error {
	if err := goose.SetDialect("pgx"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("up: %w", err)
	}

	return nil
}
