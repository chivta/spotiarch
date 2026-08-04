package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/chivta/spotiarch/internal/shared/domain"
	"github.com/chivta/spotiarch/internal/shared/repository"
	"github.com/chivta/spotiarch/internal/spotify"
)

type watchRepo interface {
	UpdateSnapshot(ctx context.Context, id int, snapshotID string, localFileCount int) error
}

type archiveRepo interface {
	UpsertSeen(ctx context.Context, watchID int, items []domain.SourceItem) error
	MarkRemoved(ctx context.Context, watchID int, presentISRCs []string) (int64, error)
	ListUnarchived(ctx context.Context, watchID int) ([]domain.ArchiveTrack, error)
	MarkArchived(ctx context.Context, ids []int) error
	CreatePart(ctx context.Context, part *domain.ArchivePart) (int, error)
	ListParts(ctx context.Context, watchID int) ([]domain.ArchivePart, error)
	SetPartTrackCount(ctx context.Context, partID, trackCount int) error
}

type spotifyClient interface {
	GetPlaylist(ctx context.Context, playlistID string) (*spotify.PlaylistResponse, error)
	GetPlaylistItems(ctx context.Context, playlistID string) ([]domain.SourceItem, int, error)
	CreateArchivePlaylist(ctx context.Context, name, description string) (string, error)
	AddItems(ctx context.Context, playlistID string, uris []string) error
}

func NewService(spotifyClient spotifyClient, watchRepo watchRepo, archiveRepo archiveRepo) *Service {
	return &Service{spotifyClient: spotifyClient, watchRepo: watchRepo, archiveRepo: archiveRepo}
}

type Service struct {
	spotifyClient spotifyClient
	watchRepo     watchRepo
	archiveRepo   archiveRepo
}

func (s *Service) ProcessWatch(ctx context.Context, watch domain.Watch) (err error) {
	start := time.Now()
	result := "error"
	defer func() {
		watcherPollDuration.Observe(time.Since(start).Seconds())
		watcherPollsTotal.WithLabelValues(result).Inc()
	}()

	playlist, err := s.spotifyClient.GetPlaylist(ctx, watch.SourcePlaylistID)
	if err != nil {
		return translateSpotifyError(ctx, err)
	}
	if playlist.SnapshotID == watch.LastSnapshotID {
		result = "unchanged"
		return nil
	}

	items, localFiles, err := s.spotifyClient.GetPlaylistItems(ctx, watch.SourcePlaylistID)
	if err != nil {
		return translateSpotifyError(ctx, err)
	}
	if err := s.archiveRepo.UpsertSeen(ctx, watch.ID, items); err != nil {
		return err
	}

	presentISRCs := make([]string, 0, len(items))
	for _, item := range items {
		presentISRCs = append(presentISRCs, item.ISRC)
	}
	removed, err := s.archiveRepo.MarkRemoved(ctx, watch.ID, presentISRCs)
	if err != nil {
		return err
	}
	watcherTracksRemovedTotal.Add(float64(removed))

	tracks, err := s.archiveRepo.ListUnarchived(ctx, watch.ID)
	if err != nil {
		return err
	}
	if err := s.projectArchive(ctx, watch.ID, playlist.Name, tracks); err != nil {
		return err
	}
	if err := s.watchRepo.UpdateSnapshot(ctx, watch.ID, playlist.SnapshotID, localFiles); err != nil {
		return err
	}

	result = "changed"
	return nil
}

func (s *Service) projectArchive(ctx context.Context, watchID int, sourceName string, tracks []domain.ArchiveTrack) error {
	if len(tracks) == 0 {
		return nil
	}

	parts, err := s.archiveRepo.ListParts(ctx, watchID)
	if err != nil {
		return err
	}

	var part domain.ArchivePart
	if len(parts) == 0 {
		part, err = s.createPart(ctx, watchID, sourceName, 1)
		if err != nil {
			return err
		}
	} else {
		part = parts[len(parts)-1]
	}

	for len(tracks) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		if part.TrackCount >= domain.ArchivePartCap {
			part, err = s.createPart(ctx, watchID, sourceName, part.PartNumber+1)
			if err != nil {
				return err
			}
		}

		batchSize := min(len(tracks), domain.ArchivePartCap-part.TrackCount, domain.SpotifyAddItemsBatch)
		batch := tracks[:batchSize]
		uris := make([]string, batchSize)
		ids := make([]int, batchSize)
		for i, track := range batch {
			uris[i] = track.URI
			ids[i] = track.ID
		}

		if err := s.spotifyClient.AddItems(ctx, part.PlaylistID, uris); err != nil {
			return translateSpotifyError(ctx, err)
		}
		if err := s.archiveRepo.MarkArchived(ctx, ids); err != nil {
			return err
		}
		part.TrackCount += batchSize
		if err := s.archiveRepo.SetPartTrackCount(ctx, part.ID, part.TrackCount); err != nil {
			return err
		}
		watcherTracksArchivedTotal.Add(float64(batchSize))
		tracks = tracks[batchSize:]
	}

	return nil
}

func (s *Service) createPart(ctx context.Context, watchID int, sourceName string, partNumber int) (domain.ArchivePart, error) {
	name := sourceName + " — Archive"
	if partNumber > 1 {
		name += fmt.Sprintf(" (Part %d)", partNumber)
	}
	description := "Append-only archive maintained by spotiarch."
	playlistID, err := s.spotifyClient.CreateArchivePlaylist(ctx, name, description)
	if err != nil {
		return domain.ArchivePart{}, translateSpotifyError(ctx, err)
	}

	part := domain.ArchivePart{
		WatchID:    watchID,
		PlaylistID: playlistID,
		PartNumber: partNumber,
	}
	part.ID, err = s.archiveRepo.CreatePart(ctx, &part)
	return part, err
}

func translateSpotifyError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return repository.TranslateSpotifyError(err)
}
