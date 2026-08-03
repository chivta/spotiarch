import type { TranslationKey } from "../i18n";

const ERROR_TRANSLATION_KEYS: Record<string, TranslationKey> = {
  INTERNAL_ERROR: "errorInternal",
  DATABASE_ERROR: "errorDatabase",
  SPOTIFY_API_ERROR: "errorSpotifyApi",
  TOO_MANY_REQUESTS: "errorTooManyRequests",
  NOT_FOUND: "errorNotFound",
  SPOTIFY_NOT_FOUND: "errorSpotifyNotFound",
  BAD_REQUEST: "errorBadRequest",
  UNAUTHORIZED: "errorUnauthorized",
  FORBIDDEN: "errorForbidden",
  ANON_QUOTA_EXCEEDED: "errorAnonQuotaExceeded",
  EMAIL_EXISTS: "errorEmailExists",
  INVALID_CREDENTIALS: "errorInvalidCredentials",
  INVALID_PLAYLIST_URL: "errorInvalidPlaylistUrl",
  PLAYLIST_NOT_PUBLIC: "errorPlaylistNotPublic",
  NO_PENDING_SELECTION: "errorNoPendingSelection",
  PENDING_EXPIRED: "errorPendingExpired",
  VERIFICATION_TOKEN_NOT_FOUND: "errorVerificationTokenNotFound",
  ALREADY_VERIFIED: "errorAlreadyVerified",
  PLAYLIST_NOT_VERIFIED: "errorPlaylistNotVerified",
  WATCH_EXISTS: "errorWatchExists",
  ARCHIVE_TRACK_NOT_FOUND: "errorArchiveTrackNotFound",
  NETWORK_ERROR: "errorNetwork",
};

export function errorTranslationKey(error: unknown): TranslationKey {
  const code = typeof error === "object" && error && "code" in error
    ? String(error.code)
    : "INTERNAL_ERROR";
  return ERROR_TRANSLATION_KEYS[code] || "errorInternal";
}
