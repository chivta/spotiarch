package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

func NewArchiveRepo(db *pgxpool.Pool) *ArchiveRepo {
	return &ArchiveRepo{db: db}
}

// ArchiveRepo owns the authoritative archive. The Spotify archive playlists are
// a projection of these rows and can be rebuilt from them.
type ArchiveRepo struct {
	db *pgxpool.Pool
}

// UpsertSeen records the tracks currently present in the source playlist.
// Re-appearing tracks clear their removal stamp rather than being duplicated.
// Tombstoned rows are left alone: the conflicting insert is swallowed, so a
// track the user deleted through the UI is never resurrected by a later poll.
func (r *ArchiveRepo) UpsertSeen(ctx context.Context, watchID int, items []domain.SourceItem) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(
			`INSERT INTO archive_tracks (watch_id, uri, isrc, in_source)
			 VALUES ($1, $2, $3, TRUE)
			 ON CONFLICT (watch_id, isrc) DO UPDATE SET in_source = TRUE, removed_at = NULL
			 WHERE archive_tracks.excluded_at IS NULL`,
			watchID, item.URI, item.ISRC,
		)
	}

	if err := r.db.SendBatch(ctx, batch).Close(); err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

// MarkRemoved stamps removed_at on every archived track whose ISRC disappeared
// from the source. Nothing is ever deleted: the archive is append-only.
func (r *ArchiveRepo) MarkRemoved(ctx context.Context, watchID int, presentISRCs []string) (int64, error) {
	tag, err := r.db.Exec(
		ctx,
		`UPDATE archive_tracks SET in_source = FALSE, removed_at = now()
		 WHERE watch_id = $1 AND in_source = TRUE AND excluded_at IS NULL AND NOT (isrc = ANY($2))`,
		watchID, presentISRCs,
	)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return 0, domain.ErrDatabaseFailure
	}
	return tag.RowsAffected(), nil
}

// ListUnarchived returns tracks that exist in the database archive but have not
// yet been pushed to a Spotify archive playlist.
func (r *ArchiveRepo) ListUnarchived(ctx context.Context, watchID int) ([]domain.ArchiveTrack, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, uri, isrc FROM archive_tracks
		 WHERE watch_id = $1 AND archived_at IS NULL AND excluded_at IS NULL ORDER BY id ASC`, watchID)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	defer rows.Close()

	var tracks []domain.ArchiveTrack
	for rows.Next() {
		track := domain.ArchiveTrack{WatchID: watchID}
		if err := rows.Scan(&track.ID, &track.URI, &track.ISRC); err != nil {
			log.Error().Msgf("db error: %T: %v", err, err)
			return nil, domain.ErrDatabaseFailure
		}
		tracks = append(tracks, track)
	}
	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("db error while listing unarchived tracks")
		return nil, domain.ErrDatabaseFailure
	}
	return tracks, nil
}

func (r *ArchiveRepo) MarkArchived(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Exec(ctx,
		`UPDATE archive_tracks SET archived_at = now() WHERE id = ANY($1)`, ids)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

// ExcludeTrack tombstones a single track. This only happens through an explicit
// user edit in the web UI; watching never removes anything. The row survives so
// its (watch_id, isrc) key keeps blocking re-insertion while the track is still
// in the source playlist.
func (r *ArchiveRepo) ExcludeTrack(ctx context.Context, watchID int, uri string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE archive_tracks SET excluded_at = now()
		 WHERE watch_id = $1 AND uri = $2 AND excluded_at IS NULL`, watchID, uri)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrArchiveTrackMissing
	}
	return nil
}

type ArchiveCounts struct {
	Total   int
	Removed int
}

func (r *ArchiveRepo) Counts(ctx context.Context, watchID int) (ArchiveCounts, error) {
	var counts ArchiveCounts
	err := r.db.QueryRow(ctx,
		`SELECT count(*), count(*) FILTER (WHERE removed_at IS NOT NULL)
		 FROM archive_tracks WHERE watch_id = $1 AND excluded_at IS NULL`, watchID,
	).Scan(&counts.Total, &counts.Removed)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return counts, domain.ErrDatabaseFailure
	}
	return counts, nil
}

// ListTracks returns one page of the archive, newest first, optionally narrowed
// to tracks the user has since removed from the source.
func (r *ArchiveRepo) ListTracks(ctx context.Context, watchID int, removedOnly bool, offset, limit int) ([]domain.ArchiveTrack, int, error) {
	// tombstoned tracks are gone as far as the user is concerned
	filter := ` AND excluded_at IS NULL`
	if removedOnly {
		filter += ` AND removed_at IS NOT NULL`
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT count(*) FROM archive_tracks WHERE watch_id = $1`+filter, watchID,
	).Scan(&total); err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, 0, domain.ErrDatabaseFailure
	}

	rows, err := r.db.Query(ctx,
		`SELECT uri, isrc, first_seen, removed_at, in_source FROM archive_tracks
		 WHERE watch_id = $1`+filter+`
		 ORDER BY first_seen DESC, id DESC OFFSET $2 LIMIT $3`,
		watchID, offset, limit)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, 0, domain.ErrDatabaseFailure
	}
	defer rows.Close()

	var tracks []domain.ArchiveTrack
	for rows.Next() {
		track := domain.ArchiveTrack{WatchID: watchID}
		if err := rows.Scan(&track.URI, &track.ISRC, &track.FirstSeen, &track.RemovedAt, &track.InSource); err != nil {
			log.Error().Msgf("db error: %T: %v", err, err)
			return nil, 0, domain.ErrDatabaseFailure
		}
		tracks = append(tracks, track)
	}
	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("db error while listing archive tracks")
		return nil, 0, domain.ErrDatabaseFailure
	}
	return tracks, total, nil
}

func (r *ArchiveRepo) CreatePart(ctx context.Context, part *domain.ArchivePart) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO archive_parts (watch_id, playlist_id, part_number, track_count)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		part.WatchID, part.PlaylistID, part.PartNumber, part.TrackCount,
	).Scan(&id)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return 0, domain.ErrDatabaseFailure
	}
	return id, nil
}

func (r *ArchiveRepo) ListParts(ctx context.Context, watchID int) ([]domain.ArchivePart, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, watch_id, playlist_id, part_number, track_count
		 FROM archive_parts WHERE watch_id = $1 ORDER BY part_number ASC`, watchID)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	defer rows.Close()

	var parts []domain.ArchivePart
	for rows.Next() {
		var part domain.ArchivePart
		if err := rows.Scan(&part.ID, &part.WatchID, &part.PlaylistID, &part.PartNumber, &part.TrackCount); err != nil {
			log.Error().Msgf("db error: %T: %v", err, err)
			return nil, domain.ErrDatabaseFailure
		}
		parts = append(parts, part)
	}
	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("db error while listing archive parts")
		return nil, domain.ErrDatabaseFailure
	}
	return parts, nil
}

func (r *ArchiveRepo) SetPartTrackCount(ctx context.Context, partID, trackCount int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE archive_parts SET track_count = $1 WHERE id = $2`, trackCount, partID)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

func (r *ArchiveRepo) DecrementPartTrackCount(ctx context.Context, watchID int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE archive_parts SET track_count = greatest(track_count - 1, 0)
		 WHERE id = (SELECT id FROM archive_parts WHERE watch_id = $1 ORDER BY part_number DESC LIMIT 1)`,
		watchID)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}
