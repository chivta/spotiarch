-- +goose Up
-- +goose StatementBegin
-- A user removing a track through the web UI is an explicit edit and must stick.
-- Deleting the row let the next poll re-insert the track straight back, because
-- it is still present in the source playlist. The row is kept as a tombstone
-- instead: it holds the (watch_id, isrc) key so the upsert cannot resurrect it.
ALTER TABLE archive_tracks ADD COLUMN excluded_at TIMESTAMPTZ;

CREATE INDEX idx_archive_tracks_excluded ON archive_tracks (watch_id) WHERE excluded_at IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_archive_tracks_excluded;
ALTER TABLE archive_tracks DROP COLUMN excluded_at;
-- +goose StatementEnd
