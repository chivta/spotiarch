package domain

import "time"

type SignupDTO struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginDTO struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type ResolvePlaylistDTO struct {
	URL string `json:"url" validate:"required,max=512"`
}

// PendingResponse lets the frontend resume the flow exactly where it stopped.
type PendingResponse struct {
	Step              PendingStep      `json:"step"`
	VerificationToken string           `json:"verification_token,omitempty"`
	Playlist          *PlaylistPreview `json:"playlist"`
	ExpiresAt         time.Time        `json:"expires_at"`
}

type WatchResponse struct {
	ID             int              `json:"id"`
	Source         *PlaylistPreview `json:"source"`
	ArchiveParts   []ArchivePartRef `json:"archive_parts"`
	Verified       bool             `json:"verified"`
	ArchivedTotal  int              `json:"archived_total"`
	RemovedTotal   int              `json:"removed_total"`
	LocalFileCount int              `json:"local_file_count"`
	LastPolledAt   *time.Time       `json:"last_polled_at"`
	CreatedAt      time.Time        `json:"created_at"`
}

type ArchivePartRef struct {
	PartNumber int    `json:"part_number"`
	PlaylistID string `json:"playlist_id"`
	SpotifyURL string `json:"spotify_url"`
	TrackCount int    `json:"track_count"`
}

type ArchiveTrackResponse struct {
	URI       string         `json:"uri"`
	ISRC      string         `json:"isrc"`
	FirstSeen time.Time      `json:"first_seen"`
	RemovedAt *time.Time     `json:"removed_at"`
	InSource  bool           `json:"in_source"`
	Metadata  *TrackMetadata `json:"metadata"`
}

type ArchiveTracksPage struct {
	Tracks []ArchiveTrackResponse `json:"tracks"`
	Total  int                    `json:"total"`
	Offset int                    `json:"offset"`
	Limit  int                    `json:"limit"`
}
