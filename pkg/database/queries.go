package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/apod.git/pkg/database/queries"
	"github.com/Dyleme/apod.git/pkg/model"
)

type Repository struct {
	db *sql.DB
	q  *queries.Queries
}

func NewRepository(db *sql.DB) (*Repository, error) {
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

func (r *Repository) AddPendingImage(ctx context.Context, date time.Time) error {
	err := r.inTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable}, func(tx *sql.Tx) error {
		_, err := r.q.FetchImagePath(ctx, tx, date)
		if err == nil { // eq nil
			return nil
		}

		if errors.Is(err, sql.ErrNoRows) {
			err = r.q.AddPendingImage(ctx, tx, date)
			if err != nil {
				return fmt.Errorf("set pending image: %w", err)
			}

			return nil
		}

		return fmt.Errorf("fetch image path: %w", err)
	})
	if err != nil {
		return fmt.Errorf("set pending image: %w", err)
	}

	return nil
}

func (r *Repository) SetImagePath(ctx context.Context, date time.Time, path string) error {
	err := r.q.SetImagePath(ctx, r.db, queries.SetImagePathParams{
		Date: date,
		ImagePath: sql.NullString{
			String: path,
			Valid:  true,
		},
	})
	if err != nil {
		return fmt.Errorf("set image path: %w", err)
	}

	return nil
}

func (r *Repository) FetchImagePath(ctx context.Context, date time.Time) (path string, err error) {
	pathSQL, err := r.q.FetchImagePath(ctx, r.db, date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", model.ErrImageNotExists
		}

		return "", fmt.Errorf("fetch image path: %w", err)
	}

	if !pathSQL.Valid {
		return "", model.ErrPendingImage
	}

	return pathSQL.String, nil
}

func (r *Repository) FetchAllImagePaths(ctx context.Context) ([]string, error) {
	sqlPaths, err := r.q.FetchAllImagePaths(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("fetch all images: %w", err)
	}

	paths := make([]string, 0, len(sqlPaths))

	for _, p := range sqlPaths {
		if p.Valid {
			paths = append(paths, p.String)
		}
	}

	return paths, nil
}
