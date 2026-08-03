# Spotiarch frontend

React 19 + TypeScript SPA for creating and managing append-only Spotify playlist archives.

## Development

```bash
npm install
cp .env.example .env
npm run dev
```

`VITE_API_BASE_URL` is the backend API base path and defaults to `/api`. Every request includes the cookie-based session.

## Verification

```bash
npx tsc --noEmit
npm run build
```
