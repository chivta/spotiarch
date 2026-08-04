## Feature: Playlist Archiving

### Concept

A user registers a Spotify playlist they own (the **source playlist**). The service
watches it continuously and maintains a second playlist (the **archive playlist**)
containing every track that has ever been in the source, including tracks the user
has since removed. The archive is append-only: nothing is ever removed from it
automatically.

### Spotify access model

There is no per-user Spotify OAuth. Spotify API access is limited to development
mode, so the application authenticates as a single **service account**:

- One long-lived Spotify OAuth token belonging to an account we control.
- All API calls — reading source playlists, creating and mutating archive
  playlists — are made with this token.
- Archive playlists are owned by the service account, not by the user.
- Users never connect a Spotify account to the service.

Consequences that the design must accept:

- The user cannot edit the archive playlist inside a Spotify client. All archive
  edits happen through our web UI, which applies them via the service account.
- The archive playlist is created with `public=false`. It remains reachable by
  link, so the user can open, follow, and play it normally; this only keeps it off
  the service account's public profile.
- The source playlist must be public for us to read its contents.

### Application authentication

The web app has its own account system, independent of Spotify. A user must be
logged in to own archives, but the flow is designed so that authentication is
requested as late as possible.

### Deferred authentication flow

An unauthenticated visitor can begin the process and complete it after signing up
or logging in, without re-entering anything.

1. **Anonymous** — visitor pastes a Spotify playlist URL. We resolve it, fetch
   playlist metadata, and show a preview (name, owner, cover, track count). The
   pending selection is persisted server-side against an anonymous session, keyed
   by a cookie.
2. **Auth gate** — the visitor proceeds to start watching. If not authenticated,
   they are sent to login/registration. The pending state is not discarded.
3. **Claim** — on successful login or registration, the anonymous session's
   pending state is transferred to the user account and the anonymous record is
   discarded.
4. **Resume** — the user lands directly on the step they left, with the playlist
   still selected, and continues to ownership verification.

The same claim mechanism applies to any partially completed step, not just the
initial paste. A pending state expires after a fixed TTL if never claimed.

### Ownership verification

Before watching begins, the user must prove they own the source playlist. We issue
a short random token and ask them to place it in the playlist's **description**;
only the playlist owner can edit that field. We then re-read the playlist and check
for the token. Once verified, the user is told they may remove it. Watching does
not start for unverified playlists.

### Watching and diffing

- Poll `GET /playlists/{id}` and compare `snapshot_id` against the last stored
  value. Fetch the full item list only when it has changed.
- Diff the new item set against the stored set. Tracks that disappeared are
  stamped with `removed_at` and appended to the archive playlist.
- Tracks that reappear later are not duplicated in the archive.

### Storage

The database is the authoritative archive; the Spotify archive playlist is a
projection that can be rebuilt from it.

- Store only track URIs, ISRCs, and timestamps (`first_seen`, `removed_at`).
- Do not persist track titles, artist names, or cover art. Re-fetch metadata from
  Spotify at render time.
- Deleting a watch deletes its stored rows.

### Constraints

- **Dedupe by ISRC**, not track ID: the same recording has different IDs across
  markets. ISRC comes from `external_ids`.
- **Playlist size cap** is approximately 10,000 tracks. When an archive approaches
  it, roll over to a numbered continuation playlist and present the set as one
  logical archive in the UI.
- **Local files** in a source playlist have no addressable Spotify URI and cannot
  be archived. Skip them and surface this in the UI rather than dropping them
  silently.
- All UI displaying Spotify metadata or cover art must link back to the
  corresponding entity on Spotify.

### Out of scope

- Per-user Spotify OAuth and collaborative-playlist invites.
- Archiving Liked Songs.
- Archiving playlists the user has not verified ownership of.