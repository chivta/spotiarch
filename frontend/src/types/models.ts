export type UserRole = "user" | "anon";
export type PendingStep = "selected" | "verify";

export interface User {
  userID: string;
  userRole: UserRole;
}

export interface AuthCredentials {
  email: string;
  password: string;
}

export interface ResolvePlaylistDTO {
  url: string;
}

export interface PlaylistPreview {
  id: string;
  name: string;
  owner_name: string;
  owner_url: string;
  image_url: string;
  track_count: number;
  public: boolean;
  description: string;
  spotify_url: string;
}

export interface PendingResponse {
  step: PendingStep;
  verification_token?: string;
  playlist: PlaylistPreview;
  expires_at: string;
}

export interface ArchivePartRef {
  part_number: number;
  playlist_id: string;
  spotify_url: string;
  track_count: number;
}

export interface WatchResponse {
  id: number;
  source: PlaylistPreview;
  archive_parts: ArchivePartRef[];
  verified: boolean;
  archived_total: number;
  removed_total: number;
  local_file_count: number;
  last_polled_at: string | null;
  created_at: string;
}

export interface TrackMetadata {
  uri: string;
  name: string;
  artists: string;
  artist_url: string;
  album_name: string;
  image_url: string;
  spotify_url: string;
}

export interface ArchiveTrackResponse {
  uri: string;
  isrc: string;
  first_seen: string;
  removed_at: string | null;
  in_source: boolean;
  metadata: TrackMetadata | null;
}

export interface ArchiveTracksPage {
  tracks: ArchiveTrackResponse[];
  total: number;
  offset: number;
  limit: number;
}
