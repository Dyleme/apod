package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/apod.git/pkg/models"
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
			return "", models.ErrImageNotExists
		}

		return "", fmt.Errorf("fetch image path: %w", err)
	}

	return path, nil
}

func (r *Repository) FetchAlbum(ctx context.Context) ([]models.AlbumRecord, error) {
	paths, err := r.q.FetchAlbum(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("fetch all images: %w", err)
	}

	album := make([]models.AlbumRecord, 0, len(paths))
	for _, p := range paths {
		album = append(album, models.AlbumRecord{
			URL:  p.ImagePath,
			Date: p.Date,
		})
	}

	return album, nil
}
