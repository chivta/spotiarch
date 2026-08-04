package api

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string `validate:"required,uri"`
	RedisURL            string `validate:"required,uri"`
	JWTSecret           string `validate:"required,min=32"`
	SpotifyClientID     string `validate:"required,min=1"`
	SpotifyClientSecret string `validate:"required,min=1"`
	SpotifyRefreshToken string `validate:"required,min=1"`
	SecureCookies       bool
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load("./.env")

	secureCookies, err := strconv.ParseBool(os.Getenv("SECURE_COOKIES"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse bool from env var SECURE_COOKIES: %w", err)
	}

	cfg := &Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN"),
		SecureCookies:       secureCookies,
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
		msgs := make([]string, 0, len(validationErrors))
		for _, e := range validationErrors {
			msgs = append(msgs, fmt.Sprintf("field '%s' failed validation: %s", e.Field(), e.Tag()))
		}
		return nil, fmt.Errorf("config validation failed: %v", msgs)
	}

	return cfg, nil
}
