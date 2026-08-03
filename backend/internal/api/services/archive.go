package services

import (
	"context"
	"strings"
	"time"

	"github.com/chivta/spotiarch/internal/shared/domain"
	"github.com/chivta/spotiarch/internal/shared/repository"
	"github.com/chivta/spotiarch/internal/spotify"
)

type (
	archivePendingRepo interface {
		GetByUserID(ctx context.Context, userID int) (*domain.PendingSelection, error)
		DeleteByUserID(ctx context.Context, userID int) error
	}
	archiveWatchRepo interface {
		Create(ctx context.Context, watch *domain.Watch) (int, error)
		GetByID(ctx context.Context, id int) (*domain.Watch, error)
		ListByUserID(ctx context.Context, userID int) ([]domain.Watch, error)
		Delete(ctx context.Context, id, userID int) error
		UpdateSnapshot(ctx context.Context, id int, snapshotID string, localFileCount int) error
	}
	archiveRepo interface {
		UpsertSeen(ctx context.Context, watchID int, items []domain.SourceItem) error
		MarkArchived(ctx context.Context, ids []int) error
		ListUnarchived(ctx context.Context, watchID int) ([]domain.ArchiveTrack, error)
		Counts(ctx context.Context, watchID int) (repository.ArchiveCounts, error)
		ListTracks(ctx context.Context, watchID int, removedOnly bool, offset, limit int) ([]domain.ArchiveTrack, int, error)
		DeleteTrack(ctx context.Context, watchID int, uri string) error
		CreatePart(ctx context.Context, part *domain.ArchivePart) (int, error)
		ListParts(ctx context.Context, watchID int) ([]domain.ArchivePart, error)
		SetPartTrackCount(ctx context.Context, partID, trackCount int) error
		DecrementPartTrackCount(ctx context.Context, watchID int) error
	}
	archiveSpotifyClient interface {
		GetPlaylist(ctx context.Context, playlistID string) (*spotify.PlaylistResponse, error)
		GetPlaylistItems(ctx context.Context, playlistID string) ([]domain.SourceItem, int, error)
		CreateArchivePlaylist(ctx context.Context, name, description string) (string, error)
		AddItems(ctx context.Context, playlistID string, uris []string) error
		RemoveItems(ctx context.Context, playlistID string, uris []string) error
		GetTracksMetadata(ctx context.Context, uris []string) (map[string]domain.TrackMetadata, error)
	}
)

func NewArchiveService(pendingRepo archivePendingRepo, watchRepo archiveWatchRepo, archiveRepo archiveRepo, spotifyClient archiveSpotifyClient) *ArchiveService {
	return &ArchiveService{pendingRepo: pendingRepo, watchRepo: watchRepo, archiveRepo: archiveRepo, spotifyClient: spotifyClient}
}

type ArchiveService struct {
	pendingRepo   archivePendingRepo
	watchRepo     archiveWatchRepo
	archiveRepo   archiveRepo
	spotifyClient archiveSpotifyClient
}

func (s *ArchiveService) CreateWatch(ctx context.Context, userID int) (*domain.WatchResponse, error) {
	pending, err := s.pendingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if pending.Step != domain.PendingStepVerify || pending.VerificationToken == "" {
		return nil, domain.ErrNotVerified
	}

	playlist, err := s.spotifyClient.GetPlaylist(ctx, pending.SourcePlaylistID)
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	if !playlist.Public {
		return nil, domain.ErrPlaylistNotPublic
	}
	if !strings.Contains(playlist.Description, pending.VerificationToken) {
		return nil, domain.ErrTokenNotInPlaylist
	}

	verifiedAt := time.Now()
	watch := &domain.Watch{
		UserID: userID, SourcePlaylistID: pending.SourcePlaylistID,
		VerificationToken: pending.VerificationToken, VerifiedAt: &verifiedAt,
	}
	watch.ID, err = s.watchRepo.Create(ctx, watch)
	if err != nil {
		return nil, err
	}

	archivePlaylistID, err := s.spotifyClient.CreateArchivePlaylist(ctx, playlist.Name+" — Archive", "Archive of "+spotify.PlaylistURL(playlist.ID))
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	part := &domain.ArchivePart{WatchID: watch.ID, PlaylistID: archivePlaylistID, PartNumber: 1}
	part.ID, err = s.archiveRepo.CreatePart(ctx, part)
	if err != nil {
		return nil, err
	}

	items, localFiles, err := s.spotifyClient.GetPlaylistItems(ctx, playlist.ID)
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	if err := s.archiveRepo.UpsertSeen(ctx, watch.ID, items); err != nil {
		return nil, err
	}
	if err := s.watchRepo.UpdateSnapshot(ctx, watch.ID, playlist.SnapshotID, localFiles); err != nil {
		return nil, err
	}

	tracks, err := s.archiveRepo.ListUnarchived(ctx, watch.ID)
	if err != nil {
		return nil, err
	}
	uris := make([]string, 0, len(tracks))
	ids := make([]int, 0, len(tracks))
	for _, track := range tracks {
		uris = append(uris, track.URI)
		ids = append(ids, track.ID)
	}
	if err := s.spotifyClient.AddItems(ctx, archivePlaylistID, uris); err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	if err := s.archiveRepo.MarkArchived(ctx, ids); err != nil {
		return nil, err
	}
	if err := s.archiveRepo.SetPartTrackCount(ctx, part.ID, len(tracks)); err != nil {
		return nil, err
	}
	if err := s.pendingRepo.DeleteByUserID(ctx, userID); err != nil {
		return nil, err
	}
	return s.GetWatch(ctx, userID, watch.ID)
}

func (s *ArchiveService) ListWatches(ctx context.Context, userID int) ([]domain.WatchResponse, error) {
	watches, err := s.watchRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	responses := make([]domain.WatchResponse, 0, len(watches))
	for i := range watches {
		response, err := s.watchResponse(ctx, &watches[i])
		if err != nil {
			return nil, err
		}
		responses = append(responses, *response)
	}
	return responses, nil
}

func (s *ArchiveService) GetWatch(ctx context.Context, userID, watchID int) (*domain.WatchResponse, error) {
	watch, err := s.ownedWatch(ctx, userID, watchID)
	if err != nil {
		return nil, err
	}
	return s.watchResponse(ctx, watch)
}

func (s *ArchiveService) ListTracks(ctx context.Context, userID, watchID int, removedOnly bool, offset, limit int) (*domain.ArchiveTracksPage, error) {
	if _, err := s.ownedWatch(ctx, userID, watchID); err != nil {
		return nil, err
	}
	tracks, total, err := s.archiveRepo.ListTracks(ctx, watchID, removedOnly, offset, limit)
	if err != nil {
		return nil, err
	}
	uris := make([]string, 0, len(tracks))
	for _, track := range tracks {
		uris = append(uris, track.URI)
	}
	metadata, err := s.spotifyClient.GetTracksMetadata(ctx, uris)
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	responses := make([]domain.ArchiveTrackResponse, 0, len(tracks))
	for _, track := range tracks {
		response := domain.ArchiveTrackResponse{
			URI: track.URI, ISRC: track.ISRC, FirstSeen: track.FirstSeen,
			RemovedAt: track.RemovedAt, InSource: track.InSource,
		}
		if value, ok := metadata[track.URI]; ok {
			response.Metadata = &value
		}
		responses = append(responses, response)
	}
	return &domain.ArchiveTracksPage{Tracks: responses, Total: total, Offset: offset, Limit: limit}, nil
}

func (s *ArchiveService) DeleteTrack(ctx context.Context, userID, watchID int, uri string) error {
	if _, err := s.ownedWatch(ctx, userID, watchID); err != nil {
		return err
	}
	parts, err := s.archiveRepo.ListParts(ctx, watchID)
	if err != nil {
		return err
	}
	if err := s.archiveRepo.DeleteTrack(ctx, watchID, uri); err != nil {
		return err
	}
	for _, part := range parts {
		if err := s.spotifyClient.RemoveItems(ctx, part.PlaylistID, []string{uri}); err != nil {
			return repository.TranslateSpotifyError(err)
		}
	}
	return s.archiveRepo.DecrementPartTrackCount(ctx, watchID)
}

func (s *ArchiveService) DeleteWatch(ctx context.Context, userID, watchID int) error {
	if _, err := s.ownedWatch(ctx, userID, watchID); err != nil {
		return err
	}
	return s.watchRepo.Delete(ctx, watchID, userID)
}

func (s *ArchiveService) ownedWatch(ctx context.Context, userID, watchID int) (*domain.Watch, error) {
	watch, err := s.watchRepo.GetByID(ctx, watchID)
	if err != nil {
		return nil, err
	}
	if watch.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return watch, nil
}

func (s *ArchiveService) watchResponse(ctx context.Context, watch *domain.Watch) (*domain.WatchResponse, error) {
	playlist, err := s.spotifyClient.GetPlaylist(ctx, watch.SourcePlaylistID)
	if err != nil {
		return nil, repository.TranslateSpotifyError(err)
	}
	parts, err := s.archiveRepo.ListParts(ctx, watch.ID)
	if err != nil {
		return nil, err
	}
	counts, err := s.archiveRepo.Counts(ctx, watch.ID)
	if err != nil {
		return nil, err
	}
	partRefs := make([]domain.ArchivePartRef, 0, len(parts))
	for _, part := range parts {
		partRefs = append(partRefs, domain.ArchivePartRef{
			PartNumber: part.PartNumber, PlaylistID: part.PlaylistID,
			SpotifyURL: spotify.PlaylistURL(part.PlaylistID), TrackCount: part.TrackCount,
		})
	}
	return &domain.WatchResponse{
		ID: watch.ID, Source: playlistPreview(playlist), ArchiveParts: partRefs,
		Verified: watch.VerifiedAt != nil, ArchivedTotal: counts.Total, RemovedTotal: counts.Removed,
		LocalFileCount: watch.LocalFileCount, LastPolledAt: watch.LastPolledAt, CreatedAt: watch.CreatedAt,
	}, nil
}
