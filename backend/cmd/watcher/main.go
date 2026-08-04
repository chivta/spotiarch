package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/metrics"
	"github.com/chivta/spotiarch/internal/shared/repository"
	"github.com/chivta/spotiarch/internal/spotify"
	"github.com/chivta/spotiarch/internal/watcher"
)

const (
	metricsAddress        = ":8081"
	initializationTimeout = time.Minute
	shutdownTimeout       = 10 * time.Second
)

func main() {
	os.Exit(runApp())
}

func runApp() int {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger().Hook(metrics.MetricsHook{
		Component: "watcher",
		Counter:   metrics.ErrorsTotalCounter,
	})
	stdlog.SetOutput(log.Logger)
	stdlog.SetFlags(0)

	cfg, err := watcher.LoadConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := initApp(ctx, cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize app")
		return 1
	}
	defer app.db.Close()

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", watcher.MetricsHandler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	server := &http.Server{Addr: metricsAddress, Handler: mux}

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("address", metricsAddress).Msg("metrics server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	workerDone := make(chan error, 1)
	go func() {
		log.Info().Msg("watcher starting")
		workerDone <- app.worker.Start(ctx)
	}()

	exitCode := 0
	workerFinished := false
	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down watcher")
	case err := <-serverErr:
		log.Error().Err(err).Msg("metrics server exited with error")
		exitCode = 1
		stop()
	case err := <-workerDone:
		workerFinished = true
		if err != nil {
			log.Error().Err(err).Msg("watcher exited with error")
			exitCode = 1
		}
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("metrics server shutdown error")
		exitCode = 1
	}

	if !workerFinished {
		select {
		case err := <-workerDone:
			if err != nil {
				log.Error().Err(err).Msg("watcher shutdown error")
				exitCode = 1
			}
		case <-shutdownCtx.Done():
			log.Error().Err(shutdownCtx.Err()).Msg("watcher shutdown timed out")
			exitCode = 1
		}
	}

	return exitCode
}

type appContainer struct {
	db     *pgxpool.Pool
	worker *watcher.Worker
}

func initApp(ctx context.Context, cfg *watcher.Config) (*appContainer, error) {
	initCtx, cancel := context.WithTimeout(ctx, initializationTimeout)
	defer cancel()

	db, err := repository.InitializeDatabase(initCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	if err := repository.RunMigrations(initCtx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	watchRepo := repository.NewWatchRepo(db)
	archiveRepo := repository.NewArchiveRepo(db)
	pendingRepo := repository.NewPendingRepo(db)
	spotifyClient := spotify.NewClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRefreshToken)
	service := watcher.NewService(spotifyClient, watchRepo, archiveRepo)

	return &appContainer{
		db:     db,
		worker: watcher.NewWorker(watchRepo, pendingRepo, service),
	}, nil
}
