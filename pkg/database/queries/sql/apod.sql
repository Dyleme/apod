-- name: AddPendingImage :exec
INSERT INTO apods 
(date)
VALUES ($1);

-- name: SetImagePath :exec
UPDATE apods
SET image_path = $2
WHERE date = $1;

-- name: FetchImagePath :one
SELECT image_path
FROM apods
WHERE date = $1;

-- name: FetchAllImagePaths :many
SELECT image_path
FROM apods
WHERE image_path IS NOT NULL ;
