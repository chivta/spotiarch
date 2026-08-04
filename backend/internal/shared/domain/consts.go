package domain

import "time"

const (
	RateLimitRequestLimit  = 60
	RateLimitWindowSeconds = 60
	// resolving a playlist hits the Spotify API, so it gets a tighter budget
	ResolveRequestLimit  = 10
	ResolveWindowSeconds = 60
)

const (
	JWTDuration          = 15 * 60           // 15 minutes
	RefreshTokenDuration = 30 * 24 * 60 * 60 // 30 days
	// if the jwt cookie ages out instead of expiring, the user would be considered
	// anonymous despite holding a valid refresh token, so both cookies live equally long
	JWTCookieAge          = 30 * 24 * 60 * 60 // 30 days
	RefreshTokenCookieAge = 30 * 24 * 60 * 60 // 30 days
	CookieJWT             = "jwt"
	CookieRefreshToken    = "refresh_token"
)

const (
	AnonSessionDuration  = 60 * 60 * 24 // 24 hours
	AnonSessionCookieAge = 60 * 60 * 24 // 24 hours
	AnonResolveLimit     = 5
)

const (
	UserRoleKey = "userRole"
	UserIDKey   = "userID"
)

type Role string

const (
	RoleUser Role = "user"
	RoleAnon Role = "anon"
)

// PendingTTL is how long an unclaimed anonymous selection survives.
const PendingTTL = 24 * time.Hour

// PendingStep tracks how far a visitor got before authentication interrupted them.
type PendingStep string

const (
	PendingStepSelected PendingStep = "selected"
	PendingStepVerify   PendingStep = "verify"
)

// VerificationTokenBytes is the entropy of the token the user pastes into the
// source playlist description to prove ownership.
const VerificationTokenBytes = 6

// VerificationTokenPrefix makes the token recognisable inside a description.
const VerificationTokenPrefix = "spotiarch-"

const (
	// SpotifyPlaylistTrackCap is Spotify's hard ceiling on playlist length.
	SpotifyPlaylistTrackCap = 10000
	// ArchivePartCap leaves headroom below the ceiling before rolling over to a
	// continuation playlist.
	ArchivePartCap = 9500
	// SpotifyAddItemsBatch is the maximum number of URIs per add-items request.
	SpotifyAddItemsBatch = 100
)

const (
	// WatchPollInterval is how often a single watch is re-checked.
	WatchPollInterval = 10 * time.Minute
	// WatcherTickInterval is how often the worker looks for due watches.
	WatcherTickInterval = 1 * time.Minute
	// WatcherBatchSize caps how many watches are processed per tick.
	WatcherBatchSize = 25
)

// TrackPageSize is the number of archive tracks returned (and hydrated from
// Spotify) per page.
const TrackPageSize = 50
