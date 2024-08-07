-- +goose Up
ALTER TABLE currencies
ALTER COLUMN code TYPE CHAR(5);

-- +goose Down
ALTER TABLE currencies
ALTER COLUMN code TYPE CHAR(3);
