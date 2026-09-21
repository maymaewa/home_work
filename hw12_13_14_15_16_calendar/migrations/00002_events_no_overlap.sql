-- +goose Up

CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE events
ADD CONSTRAINT events_no_overlap
EXCLUDE USING gist (
    tstzrange(start_at, end_at, '[)') WITH &&
);

-- +goose Down

ALTER TABLE events
DROP CONSTRAINT events_no_overlap;