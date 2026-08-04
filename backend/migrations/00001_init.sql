-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash VARCHAR(64)  NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- Server-side half of the deferred authentication flow. A row is keyed by the
-- anonymous session id until a login or signup claims it for a user account.
CREATE TABLE pending_selections (
    id                 SERIAL PRIMARY KEY,
    anon_id            VARCHAR(64),
    user_id            INTEGER REFERENCES users (id) ON DELETE CASCADE,
    source_playlist_id VARCHAR(64)  NOT NULL,
    step               VARCHAR(16)  NOT NULL,
    verification_token VARCHAR(64)  NOT NULL DEFAULT '',
    expires_at         TIMESTAMPTZ  NOT NULL,
    CONSTRAINT pending_selections_owner CHECK (anon_id IS NOT NULL OR user_id IS NOT NULL)
);

CREATE UNIQUE INDEX idx_pending_selections_anon ON pending_selections (anon_id) WHERE anon_id IS NOT NULL;
CREATE UNIQUE INDEX idx_pending_selections_user ON pending_selections (user_id) WHERE user_id IS NOT NULL;

CREATE TABLE watches (
    id                 SERIAL PRIMARY KEY,
    user_id            INTEGER      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    source_playlist_id VARCHAR(64)  NOT NULL,
    verification_token VARCHAR(64)  NOT NULL,
    verified_at        TIMESTAMPTZ,
    last_snapshot_id   VARCHAR(255) NOT NULL DEFAULT '',
    last_polled_at     TIMESTAMPTZ,
    local_file_count   INTEGER      NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (user_id, source_playlist_id)
);

CREATE INDEX idx_watches_due ON watches (last_polled_at) WHERE verified_at IS NOT NULL;

-- An archive spans one or more Spotify playlists because a playlist caps out at
-- roughly 10k tracks; parts are presented as one logical archive in the UI.
CREATE TABLE archive_parts (
    id          SERIAL PRIMARY KEY,
    watch_id    INTEGER     NOT NULL REFERENCES watches (id) ON DELETE CASCADE,
    playlist_id VARCHAR(64) NOT NULL,
    part_number INTEGER     NOT NULL,
    track_count INTEGER     NOT NULL DEFAULT 0,
    UNIQUE (watch_id, part_number)
);

-- The authoritative archive. Only identifiers and timestamps are stored; display
-- metadata is re-fetched from Spotify at render time.
CREATE TABLE archive_tracks (
    id          SERIAL PRIMARY KEY,
    watch_id    INTEGER     NOT NULL REFERENCES watches (id) ON DELETE CASCADE,
    uri         VARCHAR(96) NOT NULL,
    isrc        VARCHAR(24) NOT NULL,
    first_seen  TIMESTAMPTZ NOT NULL DEFAULT now(),
    removed_at  TIMESTAMPTZ,
    in_source   BOOLEAN     NOT NULL DEFAULT TRUE,
    archived_at TIMESTAMPTZ,
    -- dedupe is by ISRC, not track id: the same recording has different ids
    -- across markets
    UNIQUE (watch_id, isrc)
);

CREATE INDEX idx_archive_tracks_watch ON archive_tracks (watch_id, first_seen DESC);
CREATE INDEX idx_archive_tracks_unarchived ON archive_tracks (watch_id) WHERE archived_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS archive_tracks;
DROP TABLE IF EXISTS archive_parts;
DROP TABLE IF EXISTS watches;
DROP TABLE IF EXISTS pending_selections;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
