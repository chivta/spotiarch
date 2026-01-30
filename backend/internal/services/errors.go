package services

import (
	"errors"
)

var (
	ErrInternal         = errors.New("internal server error")
	ErrTokenInvalid     = errors.New("invalid jwt token")
	ErrTokenExpired     = errors.New("jwt token expired")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrEmailExists      = errors.New("email already exists")
	ErrSpotifyAPI       = errors.New("spotify api error")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrPlaylistNotFound = errors.New("playlist not found")
)
