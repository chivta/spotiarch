package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

// uniqueViolation is the Postgres error code for a unique constraint breach.
const uniqueViolation = "23505"

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

type UserRepo struct {
	db *pgxpool.Pool
}

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) (int, error) {
	var userID int
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		user.Email,
		user.PasswordHash,
	).Scan(&userID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == uniqueViolation {
			return 0, domain.ErrEmailExists
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return 0, domain.ErrDatabaseFailure
	}
	return userID, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	user.Role = domain.RoleUser
	return &user, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInvalidCredentials
		}
		log.Error().Msgf("db error: %T: %v", err, err)
		return nil, domain.ErrDatabaseFailure
	}
	user.Role = domain.RoleUser
	return &user, nil
}
