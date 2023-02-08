-- name: AddImage :exec
INSERT INTO apods 
(date, image_path)
VALUES ($1, $2);

-- name: FetchImagePath :one
SELECT image_path
FROM apods
WHERE date = $1;

-- name: FetchAllImagePaths :many
SELECT image_path
FROM apods
WHERE image_path IS NOT NULL ;