-- +goose Up
CREATE TABLE currencies (
  code CHAR(3) PRIMARY KEY,
  rate DECIMAL(10, 4) NOT NULL,
  updated_at TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE currencies;
