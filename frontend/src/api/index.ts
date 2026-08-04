import { apiRequest } from "./client";
import type {
  ArchiveTracksPage,
  AuthCredentials,
  PendingResponse,
  ResolvePlaylistDTO,
  User,
  WatchResponse,
} from "../types/models";

export const getMe = () => apiRequest<User>("/me");
export const login = (credentials: AuthCredentials) =>
  apiRequest<void>("/auth/login", { method: "POST", body: JSON.stringify(credentials) });
export const signup = (credentials: AuthCredentials) =>
  apiRequest<void>("/auth/signup", { method: "POST", body: JSON.stringify(credentials) });
export const logout = () => apiRequest<void>("/auth/logout", { method: "POST" });

export const resolvePlaylist = (payload: ResolvePlaylistDTO) =>
  apiRequest<PendingResponse>("/playlists/resolve", { method: "POST", body: JSON.stringify(payload) });
export const getPending = () => apiRequest<PendingResponse>("/pending");
export const issueVerificationToken = () =>
  apiRequest<PendingResponse>("/pending/verification-token", { method: "POST" });

export const createWatch = () => apiRequest<WatchResponse>("/watches", { method: "POST" });
export const getWatches = () => apiRequest<WatchResponse[]>("/watches");
export const getWatch = (id: number) => apiRequest<WatchResponse>(`/watches/${id}`);
export const deleteWatch = (id: number) => apiRequest<void>(`/watches/${id}`, { method: "DELETE" });
export const getArchiveTracks = (id: number, offset: number, limit: number, removedOnly: boolean) => {
  const query = new URLSearchParams({ offset: String(offset), limit: String(limit) });
  if (removedOnly) query.set("removed", "true");
  return apiRequest<ArchiveTracksPage>(`/watches/${id}/tracks?${query}`);
};
export const deleteArchiveTrack = (id: number, uri: string) =>
  apiRequest<void>(`/watches/${id}/tracks/${encodeURIComponent(uri)}`, { method: "DELETE" });
