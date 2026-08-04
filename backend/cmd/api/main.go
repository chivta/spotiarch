package main

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/api"
	"github.com/chivta/spotiarch/internal/api/handlers"
	"github.com/chivta/spotiarch/internal/api/middlewares"
	"github.com/chivta/spotiarch/internal/api/services"
	"github.com/chivta/spotiarch/internal/shared/domain"
	"github.com/chivta/spotiarch/internal/shared/metrics"
	"github.com/chivta/spotiarch/internal/shared/repository"
	"github.com/chivta/spotiarch/internal/spotify"
	"github.com/chivta/spotiarch/scripts"
)

const shutdownTimeout = 10 * time.Second

func main() {
	os.Exit(runApp())
}

func runApp() int {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger().Hook(metrics.MetricsHook{
		Component: "api", Counter: metrics.ErrorsTotalCounter,
	})
	stdlog.SetOutput(log.Logger)
	stdlog.SetFlags(0)

	cfg, err := api.LoadConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		return 1
	}
	container, err := initApp(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize app")
		return 1
	}
	defer container.db.Close()
	defer container.redis.Close()

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		ProxyHeader:             fiber.HeaderXForwardedFor,
		ErrorHandler:            handlers.RespondWithError,
	})
	app.Use(middlewares.Logger("/health", "/metrics"))
	app.Use(recover.New())
	app.Use(api.MetricsMiddleware("/health", "/metrics"))
	app.Get("/metrics", adaptor.HTTPHandler(api.MetricsHandler()))
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	apiRoutes := app.Group("/api", container.authMiddleware.ParseAuth(),
		container.rateLimitMiddleware.LimitRequests(domain.RateLimitRequestLimit, domain.RateLimitWindowSeconds))
	apiRoutes.Get("/me", container.authHandler.Me)
	apiRoutes.Post("/auth/signup", container.authHandler.Signup)
	apiRoutes.Post("/auth/login", container.authHandler.Login)
	apiRoutes.Post("/playlists/resolve",
		container.authMiddleware.RequireAnonQuota("/playlists/resolve", domain.AnonResolveLimit),
		container.playlistHandler.ResolvePlaylist)
	apiRoutes.Get("/pending", container.playlistHandler.GetPending)

	userRoutes := apiRoutes.Group("", container.authMiddleware.RequireUserRole())
	userRoutes.Post("/auth/logout", container.authHandler.Logout)
	userRoutes.Post("/pending/verification-token", container.playlistHandler.IssueVerificationToken)
	userRoutes.Post("/watches", container.watchHandler.CreateWatch)
	userRoutes.Get("/watches", container.watchHandler.ListWatches)
	userRoutes.Get("/watches/:id", container.watchHandler.GetWatch)
	userRoutes.Delete("/watches/:id", container.watchHandler.DeleteWatch)
	userRoutes.Get("/watches/:id/tracks", container.watchHandler.ListTracks)
	userRoutes.Delete("/watches/:id/tracks/:uri", container.watchHandler.DeleteTrack)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() { errCh <- app.Listen(":8080") }()

	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down server")
	case err := <-errCh:
		if err != nil {
			log.Error().Err(err).Msg("server error")
			return 1
		}
		return 0
	}
	if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
		return 1
	}
	return 0
}

type appContainer struct {
	ratelimitRepo       *repository.RatelimitRepo
	tokenRepo           *repository.TokenRepo
	userRepo            *repository.UserRepo
	pendingRepo         *repository.PendingRepo
	watchRepo           *repository.WatchRepo
	archiveRepo         *repository.ArchiveRepo
	authService         *services.AuthService
	playlistService     *services.PlaylistService
	archiveService      *services.ArchiveService
	authHandler         *handlers.AuthHandler
	playlistHandler     *handlers.PlaylistHandler
	watchHandler        *handlers.WatchHandler
	authMiddleware      *middlewares.AuthMiddleware
	rateLimitMiddleware *middlewares.RateLimitMiddleware
	db                  *pgxpool.Pool
	redis               *redis.Client
}

func initApp(cfg *api.Config) (*appContainer, error) {
	initCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	db, err := repository.InitializeDatabase(initCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := repository.RunMigrations(initCtx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}
	redisClient, err := repository.InitializeRedis(initCtx, cfg.RedisURL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	ratelimitRepo := repository.NewRatelimitRepo(redisClient)
	if err := ratelimitRepo.LoadRateLimitScript(initCtx, scripts.RateLimitScript); err != nil {
		redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("failed to load rate limit script to redis: %w", err)
	}
	tokenRepo := repository.NewTokenRepo(db, redisClient)
	userRepo := repository.NewUserRepo(db)
	pendingRepo := repository.NewPendingRepo(db)
	watchRepo := repository.NewWatchRepo(db)
	archiveRepo := repository.NewArchiveRepo(db)
	spotifyClient := spotify.NewClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRefreshToken)
	validate := validator.New()

	authService := services.NewAuthService([]byte(cfg.JWTSecret), tokenRepo, userRepo, pendingRepo)
	playlistService := services.NewPlaylistService(pendingRepo, spotifyClient)
	archiveService := services.NewArchiveService(pendingRepo, watchRepo, archiveRepo, spotifyClient)

	return &appContainer{
		ratelimitRepo: ratelimitRepo, tokenRepo: tokenRepo, userRepo: userRepo,
		pendingRepo: pendingRepo, watchRepo: watchRepo, archiveRepo: archiveRepo,
		authService: authService, playlistService: playlistService, archiveService: archiveService,
		authHandler:         handlers.NewAuthHandler(authService, validate, cfg.SecureCookies),
		playlistHandler:     handlers.NewPlaylistHandler(playlistService, validate),
		watchHandler:        handlers.NewWatchHandler(archiveService),
		authMiddleware:      middlewares.NewAuthMiddleware(authService, cfg.SecureCookies),
		rateLimitMiddleware: middlewares.NewRateLimitMiddleware(ratelimitRepo),
		db:                  db, redis: redisClient,
	}, nil
}
