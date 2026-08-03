package domain

import "time"

type User struct {
	ID           int
	Email        string
	PasswordHash string
	Role         Role
}

// PendingSelection is the server-side half of the deferred authentication flow.
// It is keyed by an anonymous session id until a login or signup claims it for a
// user account.
type PendingSelection struct {
	ID                int
	AnonID            string
	UserID            *int
	SourcePlaylistID  string
	Step              PendingStep
	VerificationToken string
	ExpiresAt         time.Time
}

// Watch is a registered source playlist plus the state needed to diff it.
type Watch struct {
	ID                int
	UserID            int
	SourcePlaylistID  string
	VerificationToken string
	VerifiedAt        *time.Time
	LastSnapshotID    string
	LastPolledAt      *time.Time
	LocalFileCount    int
	CreatedAt         time.Time
}

// ArchivePart is one Spotify playlist backing a logical archive. Archives roll
// over into numbered continuation playlists once they approach Spotify's cap.
type ArchivePart struct {
	ID         int
	WatchID    int
	PlaylistID string
	PartNumber int
	TrackCount int
}

// ArchiveTrack is the authoritative archive row. Only identifiers and timestamps
// are stored; display metadata is re-fetched from Spotify at render time.
type ArchiveTrack struct {
	ID        int
	WatchID   int
	URI       string
	ISRC      string
	FirstSeen time.Time
	RemovedAt *time.Time
	// InSource is false once the track disappeared from the source playlist.
	InSource bool
	// ArchivedAt is set when the track was pushed to the Spotify archive playlist.
	ArchivedAt *time.Time
}

// PlaylistPreview is the metadata shown before a visitor commits to watching.
type PlaylistPreview struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	OwnerName   string `json:"owner_name"`
	OwnerURL    string `json:"owner_url"`
	ImageURL    string `json:"image_url"`
	TrackCount  int    `json:"track_count"`
	Public      bool   `json:"public"`
	Description string `json:"description"`
	SpotifyURL  string `json:"spotify_url"`
}

// SourceItem is one entry of a source playlist as far as archiving cares.
type SourceItem struct {
	URI     string
	ISRC    string
	IsLocal bool
}

// TrackMetadata is display data re-fetched from Spotify, never persisted.
type TrackMetadata struct {
	URI        string `json:"uri"`
	Name       string `json:"name"`
	Artists    string `json:"artists"`
	ArtistURL  string `json:"artist_url"`
	AlbumName  string `json:"album_name"`
	ImageURL   string `json:"image_url"`
	SpotifyURL string `json:"spotify_url"`
}
