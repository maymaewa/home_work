-- +goose Up

CREATE TABLE events (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    description TEXT,
    user_id TEXT NOT NULL,
    notify_before BIGINT
);

CREATE INDEX idx_events_start_at ON events (start_at);

-- +goose Down

DROP TABLE events;