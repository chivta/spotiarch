package domain

type AppError struct {
	HTTPCode int    `json:"-"`
	Code     string `json:"code"`
}

func (e *AppError) Error() string { return e.Code }

// sentinel errors returned by the services layer to translate into HTTP responses
var (
	ErrInternal        = &AppError{500, "INTERNAL_ERROR"}
	ErrDatabaseFailure = &AppError{500, "DATABASE_ERROR"}
	ErrSpotifyAPIError = &AppError{500, "SPOTIFY_API_ERROR"}
	ErrTooManyRequests = &AppError{429, "TOO_MANY_REQUESTS"}
	ErrNotFound        = &AppError{404, "NOT_FOUND"}
	ErrSpotifyNotFound = &AppError{404, "SPOTIFY_NOT_FOUND"}
	ErrBadRequest      = &AppError{400, "BAD_REQUEST"}
	ErrUnauthorized    = &AppError{401, "UNAUTHORIZED"}
	ErrForbidden       = &AppError{403, "FORBIDDEN"}
	ErrQuotaExceeded   = &AppError{429, "ANON_QUOTA_EXCEEDED"}
	ErrEmailExists     = &AppError{409, "EMAIL_EXISTS"}

	ErrInvalidCredentials = &AppError{401, "INVALID_CREDENTIALS"}

	ErrInvalidPlaylistURL  = &AppError{400, "INVALID_PLAYLIST_URL"}
	ErrPlaylistNotPublic   = &AppError{400, "PLAYLIST_NOT_PUBLIC"}
	ErrNoPendingSelection  = &AppError{404, "NO_PENDING_SELECTION"}
	ErrPendingExpired      = &AppError{410, "PENDING_EXPIRED"}
	ErrTokenNotInPlaylist  = &AppError{400, "VERIFICATION_TOKEN_NOT_FOUND"}
	ErrAlreadyVerified     = &AppError{409, "ALREADY_VERIFIED"}
	ErrNotVerified         = &AppError{403, "PLAYLIST_NOT_VERIFIED"}
	ErrWatchExists         = &AppError{409, "WATCH_EXISTS"}
	ErrArchiveTrackMissing = &AppError{404, "ARCHIVE_TRACK_NOT_FOUND"}
)
