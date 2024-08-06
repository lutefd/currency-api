-- +goose Up
CREATE TABLE logs (
    id UUID NOT NULL,
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    source VARCHAR(255) NOT NULL,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE logs_y2024m08 PARTITION OF logs
    FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');

CREATE TABLE logs_y2024m09 PARTITION OF logs
    FOR VALUES FROM ('2024-09-01') TO ('2024-10-01');

    CREATE TABLE logs_y2024m10 PARTITION OF logs
    FOR VALUES FROM ('2024-10-01') TO ('2024-11-01');

CREATE INDEX idx_logs_timestamp ON logs (timestamp);
CREATE INDEX idx_logs_level ON logs (level);

-- +goose Down
DROP TABLE logs;
