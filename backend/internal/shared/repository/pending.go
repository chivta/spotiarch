package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

func NewPendingRepo(db *pgxpool.Pool) *PendingRepo {
	return &PendingRepo{db: db}
}

// PendingRepo stores the partially completed archiving flow so an anonymous
// visitor can authenticate mid-way without losing their selection.
type PendingRepo struct {
	db *pgxpool.Pool
}

// anon_id is null for claimed rows, so it is folded to the empty string the
// domain model uses.
const pendingColumns = `id, coalesce(anon_id, ''), user_id, source_playlist_id, step, verification_token, expires_at`

func scanPending(row pgx.Row) (*domain.PendingSelection, error) {
	var pending domain.PendingSelection
	err := row.Scan(&pending.ID, &pending.AnonID, &pending.UserID, &pending.SourcePlaylistID,
		&pending.Step, &pending.VerificationToken, &pending.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &pending, nil
}

// Upsert replaces whatever the owner had pending: a visitor only ever works on
// one selection at a time.
func (r *PendingRepo) Upsert(ctx context.Context, pending *domain.PendingSelection) error {
	var conflict string
	if pending.UserID != nil {
		conflict = `(user_id) WHERE user_id IS NOT NULL`
	} else {
		conflict = `(anon_id) WHERE anon_id IS NOT NULL`
	}

	_, err := r.db.Exec(
		ctx,
		// nullif keeps user-owned rows out of the partial unique index on anon_id,
		// which would otherwise collide on the empty string across users
		`INSERT INTO pending_selections (anon_id, user_id, source_playlist_id, step, verification_token, expires_at)
		 VALUES (nullif($1, ''), $2, $3, $4, $5, $6)
		 ON CONFLICT `+conflict+` DO UPDATE SET
		   source_playlist_id = EXCLUDED.source_playlist_id,
		   step = EXCLUDED.step,
		   verification_token = EXCLUDED.verification_token,
		   expires_at = EXCLUDED.expires_at`,
		pending.AnonID, pending.UserID, pending.SourcePlaylistID,
		pending.Step, pending.VerificationToken, pending.ExpiresAt.UTC(),
	)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

func (r *PendingRepo) GetByUserID(ctx context.Context, userID int) (*domain.PendingSelection, error) {
	pending, err := scanPending(r.db.QueryRow(ctx,
		`SELECT `+pendingColumns+` FROM pending_selections WHERE user_id = $1`, userID))
	return r.translate(pending, err)
}

func (r *PendingRepo) GetByAnonID(ctx context.Context, anonID string) (*domain.PendingSelection, error) {
	pending, err := scanPending(r.db.QueryRow(ctx,
		`SELECT `+pendingColumns+` FROM pending_selections WHERE anon_id = $1`, anonID))
	return r.translate(pending, err)
}

func (r *PendingRepo) translate(pending *domain.PendingSelection, err error) (*domain.PendingSelection, error) {
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNoPendingSelection
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	if time.Now().After(pending.ExpiresAt) {
		return nil, domain.ErrPendingExpired
	}
	return pending, nil
}

// Claim transfers an anonymous session's pending state to a user account and
// discards the anonymous record. It is a no-op when nothing is pending.
func (r *PendingRepo) Claim(ctx context.Context, anonID string, userID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	defer tx.Rollback(ctx)

	// The user may already have a pending selection from another device; the
	// freshly claimed one wins.
	if _, err := tx.Exec(ctx,
		`DELETE FROM pending_selections WHERE user_id = $1
		   AND EXISTS (SELECT 1 FROM pending_selections WHERE anon_id = $2)`,
		userID, anonID,
	); err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}

	if _, err := tx.Exec(ctx,
		`UPDATE pending_selections SET user_id = $1, anon_id = NULL WHERE anon_id = $2`,
		userID, anonID,
	); err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

func (r *PendingRepo) DeleteByUserID(ctx context.Context, userID int) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM pending_selections WHERE user_id = $1`, userID); err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return domain.ErrDatabaseFailure
	}
	return nil
}

// DeleteExpired drops unclaimed selections that outlived their TTL.
func (r *PendingRepo) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM pending_selections WHERE expires_at < now()`)
	if err != nil {
		log.Error().Msgf("db error: %T: %v", err, err)
		return 0, domain.ErrDatabaseFailure
	}
	return tag.RowsAffected(), nil
}
