# spotiarch

Spotiarch watches public Spotify playlists that a user owns and preserves every
track that disappears. The database is the source of truth; one or more private
Spotify archive playlists are append-only projections that users can follow and
play. Playlist ownership is verified through a temporary token in the source
playlist description.

## Architecture

- `api`: Go/Fiber HTTP API on port 8080, backed by PostgreSQL and Redis.
- `watcher`: Go polling worker with health and metrics endpoints on port 8081.
- `frontend`: React SPA built with Vite and served by nginx in production.
- PostgreSQL stores accounts, watches, pending sessions, and archived track data.
- Redis stores API cache and other ephemeral data.

Both Go processes embed and run database migrations during startup. There is no
separate migration container or deployment job.

## Local development

Install Docker with Compose, put the Spotify service-account credentials in
`backend/.env` (see `backend/.example.env`), and start the development stack:

```sh
cp backend/.example.env backend/.env   # then fill in the Spotify values
docker compose up --build
```

### Minting the service-account refresh token

There is no per-user Spotify OAuth: the app acts as a single service account, and
Spotify has no notion of an app-owned playlist, so creating archive playlists
needs a token that acts as that user. Client credentials are not enough.

Run this once, with the app's Client ID and secret already in `backend/.env`:

```sh
cd backend && go run ./cmd/spotify-auth
```

It prints an authorisation URL, captures the callback on a loopback listener, and
writes `SPOTIFY_REFRESH_TOKEN` back into `backend/.env`. The token is never
printed. Approve as the **service account**, not a personal account — the tool
forces Spotify's account chooser and afterwards reports which account it
authorised, so a mistake is obvious.

The default redirect URI is `http://127.0.0.1:8888/callback` and it must be
registered on the Spotify app exactly. Pass `-redirect` if the app has a
different one registered.

The frontend is available at <http://localhost:5173>, the API at
<http://localhost:8080>, and watcher diagnostics at <http://localhost:8081>.
Backend source is bind-mounted and reloaded by Air; frontend source is reloaded by
Vite. PostgreSQL and Redis data are kept in named volumes.

## Configuration

Production secrets belong in the SOPS-encrypted Kubernetes secret manifest:

- `DATABASE_URL`
- `DATABASE_PASSWORD`
- `JWT_SECRET`
- `SPOTIFY_CLIENT_ID`
- `SPOTIFY_CLIENT_SECRET`
- `SPOTIFY_REFRESH_TOKEN`

The API also receives `REDIS_URL=redis://redis:6379` and
`SECURE_COOKIES=true` from its Deployment. Local Compose supplies the
development-only database, Redis and cookie settings inline, and loads secrets
from `backend/.env` via `env_file`. Secrets must not be listed under Compose's
`environment:` block: an unset shell variable expands to the empty string, which
counts as "already set" and stops the `.env` file from being applied.

## Deployment

CI builds, tests, and vets the Go module and builds the frontend. After CI passes
on `main`, CD compares the current commit with the full image SHAs recorded in
`k8s/kustomization.yaml`, builds only changed components, and publishes SHA and
`latest` tags to `ghcr.io/chivta/spotiarch`.

The deploy job updates image tags with Kustomize and commits the manifest back to
`main` using the reserved `deploy:` prefix. It never applies resources directly;
the GitOps controller reconciles the `spotiarch` namespace from Git. Before the
first deployment, replace `k8s/secrets.enc.yaml` with real SOPS-encrypted
`app-secrets` and `ghcr-secret` resources.
