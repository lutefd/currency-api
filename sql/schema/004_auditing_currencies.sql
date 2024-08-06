-- +goose Up
ALTER TABLE currencies
ADD COLUMN created_by UUID,
ADD COLUMN updated_by UUID,
ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE currencies
DROP COLUMN created_by,
DROP COLUMN updated_by,
DROP COLUMN created_at;
