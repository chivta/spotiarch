# HTTP API

Base path `/api`. All responses are JSON. Errors are `{"code": "MACHINE_READABLE_CODE"}` with the
HTTP status from `internal/shared/domain/errors.go`. Human-readable strings live in the frontend
translation map, never in a response body.

Authentication is cookie-based (`jwt` + `refresh_token`, both `HttpOnly`, `SameSite=Lax`, `Secure`
in production). Every `/api` request passes through `ParseAuth`, which issues an anonymous session
JWT when no cookie is present, so an unauthenticated visitor always has a stable `anon` identity.

## Unauthenticated / anonymous

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/me` | `{"userID": "...", "userRole": "user"\|"anon"}` |
| `POST` | `/api/auth/signup` | `SignupDTO`. Claims the anonymous pending selection on success. |
| `POST` | `/api/auth/login` | `LoginDTO`. Claims the anonymous pending selection on success. |
| `POST` | `/api/playlists/resolve` | `ResolvePlaylistDTO`. Resolves the URL, fetches metadata, persists the pending selection against the caller (anon or user), returns `PendingResponse`. Anon callers are capped by `AnonResolveLimit`. |
| `GET` | `/api/pending` | Current `PendingResponse`, or `NO_PENDING_SELECTION` / `PENDING_EXPIRED`. |

## Authenticated (`RequireUserRole`)

| Method | Path | Notes |
| --- | --- | --- |
| `POST` | `/api/auth/logout` | Clears cookies and revokes the refresh token. |
| `POST` | `/api/pending/verification-token` | Issues the ownership token, advances the pending step to `verify`, returns `PendingResponse`. Idempotent — re-issuing returns the existing token. |
| `POST` | `/api/watches` | Re-reads the source playlist, checks the description for the token, creates the watch and the first archive playlist, seeds the archive from the current tracks, clears the pending selection. Returns `WatchResponse`. |
| `GET` | `/api/watches` | `[]WatchResponse` for the caller. |
| `GET` | `/api/watches/:id` | One `WatchResponse`. |
| `DELETE` | `/api/watches/:id` | Deletes the watch and all of its stored rows. |
| `GET` | `/api/watches/:id/tracks` | `ArchiveTracksPage`. Query: `offset` (default 0), `limit` (default and max `TrackPageSize`), `removed=true` to show only tracks the user has since removed. Metadata is re-fetched from Spotify per page and never persisted. |
| `DELETE` | `/api/watches/:id/tracks/:uri` | Removes one track from the archive, in the database and in the Spotify archive playlist. The user cannot do this in a Spotify client because the archive is owned by the service account. |

## Operational

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/health` | 204. Skipped by logging and metrics middleware. |
| `GET` | `/metrics` | Prometheus exposition. Skipped by logging and metrics middleware. |

DTO shapes are defined in `internal/shared/domain/dto.go` and are the single source of truth.
