package repository

import (
	"context"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
	"github.com/chivta/spotiarch/internal/shared/metrics"
	"github.com/chivta/spotiarch/internal/spotify"
	"github.com/chivta/spotiarch/migrations"
)

func InitializeDatabase(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.Tracer = metrics.PgxTracer()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// RunMigrations applies the embedded migrations. There is no separate migration
// step in deployment; every process runs this on startup.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, ".")
}

func InitializeRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	rds := redis.NewClient(options)
	if _, err := rds.Ping(ctx).Result(); err != nil {
		rds.Close()
		return nil, err
	}

	rds.AddHook(metrics.RedisHook())
	return rds, nil
}

// TranslateSpotifyError logs unexpected Spotify API and network errors before
// translating them into application errors.
func TranslateSpotifyError(err error) error {
	if apiErr, ok := err.(*spotify.APIError); ok {
		switch apiErr.Status {
		case 404:
			return domain.ErrSpotifyNotFound
		case 400:
			return domain.ErrBadRequest
		case 429:
			return domain.ErrTooManyRequests
		default:
			log.Error().Err(apiErr).Int("status", apiErr.Status).Msg("spotify API error")
			return domain.ErrSpotifyAPIError
		}
	}
	if _, ok := err.(*url.Error); ok {
		log.Error().Err(err).Msg("network error when calling spotify API")
		return domain.ErrSpotifyAPIError
	}
	log.Error().Err(err).Msg("unexpected error when calling spotify API")
	return domain.ErrSpotifyAPIError
}
