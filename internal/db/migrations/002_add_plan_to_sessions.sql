-- +goose Up
ALTER TABLE sessions ADD COLUMN plan TEXT;

-- +goose Down
ALTER TABLE sessions DROP COLUMN plan;
