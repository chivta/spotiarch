package handlers

import (
	"github.com/chivta/spotiarch/internal/services"
)

func NewSpotifyHandler(spotifyService *services.SpotifyService) *SpotifyHandler {
	return &SpotifyHandler{spotifyService: spotifyService}
}

type SpotifyHandler struct {
	spotifyService *services.SpotifyService
}
