-- +goose Up
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';

-- +goose Down
ALTER TABLE expenses DROP COLUMN IF EXISTS description;
