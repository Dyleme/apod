package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/apod.git/pkg/model"
	"github.com/Dyleme/apod.git/pkg/repository/queries"
)

type Repository struct {
	db *sql.DB
	q  *queries.Queries
}

func New(db *sql.DB) (*Repository, error) {
	if err := migrateUp(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Repository{
		db: db,
		q:  &queries.Queries{},
	}, nil
}

// inTx is method which allows you to make queries in transaction.
func (r *Repository) inTx(ctx context.Context, options *sql.TxOptions, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			return fmt.Errorf("rolling back transaction %w, (original error %w)", err1, err)
		}

		return fmt.Errorf("in tx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (r *Repository) AddImage(ctx context.Context, date time.Time, path string) error {
	err := r.q.AddImage(ctx, r.db, queries.AddImageParams{
		Date:      date,
		ImagePath: path,
	})
	if err != nil {
		return fmt.Errorf("set image path: %w", err)
	}

	return nil
}

func (r *Repository) FetchImagePath(ctx context.Context, date time.Time) (string, error) {
	path, err := r.q.FetchImagePath(ctx, r.db, date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", model.ErrImageNotExists
		}

		return "", fmt.Errorf("fetch image path: %w", err)
	}

	return path, nil
}

func (r *Repository) FetchAllImagePaths(ctx context.Context) ([]string, error) {
	paths, err := r.q.FetchAllImagePaths(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("fetch all images: %w", err)
	}

	return paths, nil
}
