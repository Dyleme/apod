-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS apods (
    date date PRIMARY KEY,
    image_path varchar(251)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apods;
-- +goose StatementEnd
