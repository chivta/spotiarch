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

Install Docker with Compose, export the Spotify service-account credentials, and
start the development stack:

```sh
export SPOTIFY_CLIENT_ID=...
export SPOTIFY_CLIENT_SECRET=...
export SPOTIFY_REFRESH_TOKEN=...
docker compose up --build
```

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
`SECURE_COOKIES=true` from its Deployment. Local Compose supplies development-only
database, Redis, JWT, and cookie settings inline; Spotify values always come from
the developer's environment.

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
