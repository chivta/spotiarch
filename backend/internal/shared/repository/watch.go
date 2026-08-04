package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

func NewWatchRepo(db *pgxpool.Pool) *WatchRepo {
	return &WatchRepo{db: db}
}

type WatchRepo struct {
	db *pgxpool.Pool
}

const watchColumns = `id, user_id, source_playlist_id, verification_token, verified_at,
	last_snapshot_id, last_polled_at, local_file_count, created_at`

func scanWatch(row pgx.Row) (*domain.Watch, error) {
	var watch domain.Watch
	err := row.Scan(&watch.ID, &watch.UserID, &watch.SourcePlaylistID, &watch.VerificationToken,
		&watch.VerifiedAt, &watch.LastSnapshotID, &watch.LastPolledAt, &watch.LocalFileCount, &watch.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &watch, nil
}

func (r *WatchRepo) Create(ctx context.Context, watch *domain.Watch) (int, error) {
	var id int
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO watches (user_id, source_playlist_id, verification_token, verified_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		watch.UserID, watch.SourcePlaylistID, watch.VerificationToken, watch.VerifiedAt,
	).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == uniqueViolation {
			return 0, domain.ErrWatchExists
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return 0, domain.ErrDatabaseFailure
	}
	return id, nil
}

func (r *WatchRepo) GetByID(ctx context.Context, id int) (*domain.Watch, error) {
	watch, err := scanWatch(r.db.QueryRow(ctx, `SELECT `+watchColumns+` FROM watches WHERE id = $1`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	return watch, nil
}

func (r *WatchRepo) ListByUserID(ctx context.Context, userID int) ([]domain.Watch, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+watchColumns+` FROM watches WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	defer rows.Close()

	var watches []domain.Watch
	for rows.Next() {
		watch, err := scanWatch(rows)
		if err != nil {
			log.Error().Msgf("db error: %T: %v", err, err)
			return nil, domain.ErrDatabaseFailure
		}
		watches = append(watches, *watch)
	}
	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("db error while listing watches")
		return nil, domain.ErrDatabaseFailure
	}
	return watches, nil
}

func (r *WatchRepo) Delete(ctx context.Context, id, userID int) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM watches WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ClaimDue returns verified watches whose poll interval has elapsed, stamping
// last_polled_at so concurrent workers never pick up the same watch twice.
func (r *WatchRepo) ClaimDue(ctx context.Context, interval time.Duration, limit int) ([]domain.Watch, error) {
	rows, err := r.db.Query(
		ctx,
		`UPDATE watches SET last_polled_at = now() WHERE id IN (
		   SELECT id FROM watches
		   WHERE verified_at IS NOT NULL
		     AND (last_polled_at IS NULL OR last_polled_at < now() - $1::interval)
		   ORDER BY last_polled_at ASC NULLS FIRST
		   LIMIT $2
		   FOR UPDATE SKIP LOCKED
		 ) RETURNING `+watchColumns,
		interval, limit,
	)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	defer rows.Close()

	var watches []domain.Watch
	for rows.Next() {
		watch, err := scanWatch(rows)
		if err != nil {
			log.Error().Msgf("db error: %T: %v", err, err)
			return nil, domain.ErrDatabaseFailure
		}
		watches = append(watches, *watch)
	}
	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Msg("db error while claiming due watches")
		return nil, domain.ErrDatabaseFailure
	}
	return watches, nil
}

func (r *WatchRepo) UpdateSnapshot(ctx context.Context, id int, snapshotID string, localFileCount int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE watches SET last_snapshot_id = $1, local_file_count = $2 WHERE id = $3`,
		snapshotID, localFileCount, id)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}
