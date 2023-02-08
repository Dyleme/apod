// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: apod.sql

package queries

import (
	"context"
	"time"
)

const addImage = `-- name: AddImage :exec
INSERT INTO apods 
(date, image_path)
VALUES ($1, $2)
`

type AddImageParams struct {
	Date      time.Time
	ImagePath string
}

func (q *Queries) AddImage(ctx context.Context, db DBTX, arg AddImageParams) error {
	_, err := db.ExecContext(ctx, addImage, arg.Date, arg.ImagePath)
	return err
}

const fetchAllImagePaths = `-- name: FetchAllImagePaths :many
SELECT image_path
FROM apods
WHERE image_path IS NOT NULL
`

func (q *Queries) FetchAllImagePaths(ctx context.Context, db DBTX) ([]string, error) {
	rows, err := db.QueryContext(ctx, fetchAllImagePaths)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var image_path string
		if err := rows.Scan(&image_path); err != nil {
			return nil, err
		}
		items = append(items, image_path)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const fetchImagePath = `-- name: FetchImagePath :one
SELECT image_path
FROM apods
WHERE date = $1
`

func (q *Queries) FetchImagePath(ctx context.Context, db DBTX, date time.Time) (string, error) {
	row := db.QueryRowContext(ctx, fetchImagePath, date)
	var image_path string
	err := row.Scan(&image_path)
	return image_path, err
}
