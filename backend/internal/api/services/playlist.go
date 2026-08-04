package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/chivta/spotiarch/internal/shared/domain"
	"github.com/chivta/spotiarch/internal/shared/repository"
	"github.com/chivta/spotiarch/internal/spotify"
)

type (
	pendingRepo interface {
		Upsert(ctx context.Context, pending *domain.PendingSelection) error
		GetByUserID(ctx context.Context, userID int) (*domain.PendingSelection, error)
		GetByAnonID(ctx context.Context, anonID string) (*domain.PendingSelection, error)
	}
	playlistSpotifyClient interface {
		GetPlaylist(ctx context.Context, playlistID string) (*spotify.PlaylistResponse, error)
	}
)

func NewPlaylistService(pendingRepo pendingRepo, spotifyClient playlistSpotifyClient) *PlaylistService {
	return &PlaylistService{pendingRepo: pendingRepo, spotifyClient: spotifyClient}
}

type PlaylistService struct {
	pendingRepo   pendingRepo
	spotifyClient playlistSpotifyClient
}

func (s *PlaylistService) Resolve(ctx context.Context, ownerAnonID string, ownerUserID *int, playlistURL string) (*domain.PendingResponse, error) {
	playlistID := spotify.PlaylistIDFromURL(playlistURL)
	if playlistID == "" {
		return nil, domain.ErrInvalidPlaylistURL
	}
	playlist, err := s.spotifyClient.GetPlaylist(ctx, playlistID)
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	if !playlist.Public {
		return nil, domain.ErrPlaylistNotPublic
	}

	pending := &domain.PendingSelection{
		AnonID: ownerAnonID, UserID: ownerUserID, SourcePlaylistID: playlistID,
		Step: domain.PendingStepSelected, ExpiresAt: time.Now().Add(domain.PendingTTL),
	}
	if err := s.pendingRepo.Upsert(ctx, pending); err != nil {
		return nil, err
	}
	return pendingResponse(pending, playlist), nil
}

func (s *PlaylistService) GetPending(ctx context.Context, ownerAnonID string, ownerUserID *int) (*domain.PendingResponse, error) {
	var pending *domain.PendingSelection
	var err error
	if ownerUserID != nil {
		pending, err = s.pendingRepo.GetByUserID(ctx, *ownerUserID)
	} else {
		pending, err = s.pendingRepo.GetByAnonID(ctx, ownerAnonID)
	}
	if err != nil {
		return nil, err
	}
	playlist, err := s.spotifyClient.GetPlaylist(ctx, pending.SourcePlaylistID)
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	return pendingResponse(pending, playlist), nil
}

func (s *PlaylistService) IssueVerificationToken(ctx context.Context, userID int) (*domain.PendingResponse, error) {
	pending, err := s.pendingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if pending.Step != domain.PendingStepVerify {
		raw := make([]byte, domain.VerificationTokenBytes)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		pending.VerificationToken = domain.VerificationTokenPrefix + hex.EncodeToString(raw)
		pending.Step = domain.PendingStepVerify
		if err := s.pendingRepo.Upsert(ctx, pending); err != nil {
			return nil, err
		}
	}
	playlist, err := s.spotifyClient.GetPlaylist(ctx, pending.SourcePlaylistID)
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	return pendingResponse(pending, playlist), nil
}

func pendingResponse(pending *domain.PendingSelection, playlist *spotify.PlaylistResponse) *domain.PendingResponse {
	return &domain.PendingResponse{
		Step: pending.Step, VerificationToken: pending.VerificationToken,
		Playlist: playlistPreview(playlist), ExpiresAt: pending.ExpiresAt,
	}
}

func playlistPreview(playlist *spotify.PlaylistResponse) *domain.PlaylistPreview {
	preview := &domain.PlaylistPreview{
		ID: playlist.ID, Name: playlist.Name, OwnerName: playlist.Owner.DisplayName,
		OwnerURL: playlist.Owner.ExternalURLs.Spotify, TrackCount: playlist.Tracks.Total,
		Public: playlist.Public, Description: playlist.Description, SpotifyURL: spotify.PlaylistURL(playlist.ID),
	}
	if len(playlist.Images) > 0 {
		preview.ImageURL = playlist.Images[0].URL
	}
	return preview
}
