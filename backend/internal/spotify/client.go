package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

const (
	defaultTokenURL = "https://accounts.spotify.com/api/token"
	apiBase         = "https://api.spotify.com/v1"

	// itemsPageLimit is Spotify's maximum page size for playlist items.
	itemsPageLimit = 100
	// tracksBatchLimit is Spotify's maximum number of ids per /tracks call.
	tracksBatchLimit = 50

	requestsPerWindow = 175
	slidingWindowSize = 30 * time.Second
	maxConcurrent     = 5

	// itemFields keeps playlist item responses small: archiving only needs the
	// URI, the ISRC and whether the entry is a local file.
	itemFields = "total,next,items(is_local,track(uri,external_ids(isrc)))"
)

// APIError carries the Spotify HTTP status so callers can translate it into a
// domain error.
type APIError struct {
	Status int
}

func (e *APIError) Error() string { return "spotify API error: status code " + strconv.Itoa(e.Status) }

func NewClient(clientID, clientSecret, refreshToken string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		tokenURL:     defaultTokenURL,
		sem:          make(chan struct{}, maxConcurrent),
		window:       NewSlidingWindow(slidingWindowSize, requestsPerWindow),
	}
}

// Client talks to Spotify as a single service account. There is no per-user
// OAuth: every call — reading source playlists and mutating archive playlists —
// uses the same long-lived refresh token.
type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	tokenURL     string

	// refreshToken is guarded by tokenMu because Spotify may hand back a rotated
	// one on refresh.
	refreshToken string

	blockedUntil time.Time
	blockMu      sync.RWMutex
	sem          chan struct{}
	window       *SlidingWindow

	accessToken string
	tokenExpiry time.Time
	tokenMu     sync.RWMutex

	serviceUserID string
	userMu        sync.Mutex
}

func (c *Client) getValidToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	if c.accessToken != "" && c.tokenExpiry.After(time.Now()) {
		token := c.accessToken
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	// Recheck after acquiring the write lock in case another goroutine refreshed.
	if c.accessToken != "" && c.tokenExpiry.After(time.Now()) {
		return c.accessToken, nil
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.refreshToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", &APIError{Status: resp.StatusCode}
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	// Spotify may rotate the refresh token. Keeping the stale one would break
	// every refresh after the old token is invalidated. This only lives for the
	// process lifetime — a restart falls back to SPOTIFY_REFRESH_TOKEN, which
	// Spotify keeps honouring, so there is nothing to persist.
	if tokenResp.RefreshToken != "" && tokenResp.RefreshToken != c.refreshToken {
		c.refreshToken = tokenResp.RefreshToken
		log.Info().Msg("spotify rotated the refresh token")
	}

	c.accessToken = tokenResp.AccessToken
	// refresh a minute early so in-flight calls never race the expiry
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - time.Minute)
	return c.accessToken, nil
}

func (c *Client) block(duration time.Duration) {
	c.blockMu.Lock()
	defer c.blockMu.Unlock()
	c.blockedUntil = time.Now().Add(duration)
}

func (c *Client) isBlocked() bool {
	c.blockMu.RLock()
	defer c.blockMu.RUnlock()
	return c.blockedUntil.After(time.Now())
}

// do executes an authenticated request against the Spotify API, honouring the
// sliding window and 429 backoff. On error the body is closed internally;
// callers close the body only on success.
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	if c.isBlocked() {
		return nil, &APIError{Status: http.StatusTooManyRequests}
	}

	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.window.Wait(ctx); err != nil {
		return nil, err
	}

	var reader *bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	APIDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	if err != nil {
		APIRequestsTotal.WithLabelValues(method, "network_error").Inc()
		return nil, err
	}
	APIRequestsTotal.WithLabelValues(method, strconv.Itoa(resp.StatusCode)).Inc()

	if resp.StatusCode == http.StatusTooManyRequests {
		if secs, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
			log.Error().Int("retry_seconds", secs).Msg("spotify rate limit hit, blocking requests")
			c.block(time.Duration(secs) * time.Second)
		}
		resp.Body.Close()
		return nil, &APIError{Status: http.StatusTooManyRequests}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		return nil, &APIError{Status: resp.StatusCode}
	}
	return resp, nil
}

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

// ServiceUserID returns the Spotify id of the account that owns every archive
// playlist. It is fetched once and cached for the process lifetime.
func (c *Client) ServiceUserID(ctx context.Context) (string, error) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	if c.serviceUserID != "" {
		return c.serviceUserID, nil
	}

	var me meResponse
	if err := c.getJSON(ctx, "/me", &me); err != nil {
		return "", err
	}
	c.serviceUserID = me.ID
	return c.serviceUserID, nil
}

// GetPlaylist fetches playlist metadata plus its snapshot id. The item list is
// deliberately limited to a single entry: callers that need the full list use
// GetPlaylistItems, and everyone else only wants the metadata and snapshot.
func (c *Client) GetPlaylist(ctx context.Context, playlistID string) (*PlaylistResponse, error) {
	var playlist PlaylistResponse
	path := fmt.Sprintf("/playlists/%s?fields=%s", playlistID, url.QueryEscape(
		"id,name,description,public,snapshot_id,images,owner(id,display_name,external_urls),external_urls,tracks(total)"))
	if err := c.getJSON(ctx, path, &playlist); err != nil {
		return nil, err
	}
	return &playlist, nil
}

// GetPlaylistItems walks every page of a playlist. Local files have no
// addressable URI, so they are returned separately as a count rather than being
// dropped silently.
func (c *Client) GetPlaylistItems(ctx context.Context, playlistID string) (items []domain.SourceItem, localFiles int, err error) {
	for offset := 0; ; offset += itemsPageLimit {
		path := fmt.Sprintf("/playlists/%s/tracks?offset=%d&limit=%d&fields=%s",
			playlistID, offset, itemsPageLimit, url.QueryEscape(itemFields))

		var page PlaylistItemsPage
		if err := c.getJSON(ctx, path, &page); err != nil {
			return nil, 0, err
		}

		for _, item := range page.Items {
			if item.IsLocal || item.Track == nil || item.Track.URI == "" {
				localFiles++
				continue
			}
			// A track without an ISRC cannot be deduped across markets reliably;
			// fall back to the URI so it is still archived exactly once.
			isrc := item.Track.ExternalIDs.ISRC
			if isrc == "" {
				isrc = item.Track.URI
			}
			items = append(items, domain.SourceItem{URI: item.Track.URI, ISRC: isrc})
		}

		if page.Next == nil || len(page.Items) == 0 {
			return items, localFiles, nil
		}
	}
}

// CreateArchivePlaylist creates a private playlist owned by the service account.
// It stays reachable by link so the user can open, follow and play it; private
// only keeps it off the service account's public profile.
func (c *Client) CreateArchivePlaylist(ctx context.Context, name, description string) (string, error) {
	userID, err := c.ServiceUserID(ctx)
	if err != nil {
		return "", err
	}

	body := map[string]any{
		"name":          name,
		"description":   description,
		"public":        false,
		"collaborative": false,
	}
	resp, err := c.do(ctx, http.MethodPost, "/users/"+userID+"/playlists", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", err
	}
	return created.ID, nil
}

// AddItems appends URIs to a playlist in Spotify-sized batches.
func (c *Client) AddItems(ctx context.Context, playlistID string, uris []string) error {
	for start := 0; start < len(uris); start += domain.SpotifyAddItemsBatch {
		end := min(start+domain.SpotifyAddItemsBatch, len(uris))
		resp, err := c.do(ctx, http.MethodPost, "/playlists/"+playlistID+"/tracks",
			map[string]any{"uris": uris[start:end]})
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}

// RemoveItems deletes URIs from a playlist. Archives are append-only
// automatically; this exists for explicit user edits made through the web UI.
func (c *Client) RemoveItems(ctx context.Context, playlistID string, uris []string) error {
	for start := 0; start < len(uris); start += domain.SpotifyAddItemsBatch {
		end := min(start+domain.SpotifyAddItemsBatch, len(uris))
		tracks := make([]map[string]string, 0, end-start)
		for _, uri := range uris[start:end] {
			tracks = append(tracks, map[string]string{"uri": uri})
		}
		resp, err := c.do(ctx, http.MethodDelete, "/playlists/"+playlistID+"/tracks",
			map[string]any{"tracks": tracks})
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}

// GetTracksMetadata hydrates display data for archive rows. Nothing here is
// persisted — titles, artists and cover art are always re-fetched.
func (c *Client) GetTracksMetadata(ctx context.Context, uris []string) (map[string]domain.TrackMetadata, error) {
	out := make(map[string]domain.TrackMetadata, len(uris))

	ids := make([]string, 0, len(uris))
	for _, uri := range uris {
		if id := TrackIDFromURI(uri); id != "" {
			ids = append(ids, id)
		}
	}

	for start := 0; start < len(ids); start += tracksBatchLimit {
		end := min(start+tracksBatchLimit, len(ids))

		var batch tracksResponse
		if err := c.getJSON(ctx, "/tracks?ids="+strings.Join(ids[start:end], ","), &batch); err != nil {
			return nil, err
		}

		for _, track := range batch.Tracks {
			if track == nil {
				continue
			}
			names := make([]string, 0, len(track.Artists))
			artistURL := ""
			for _, artist := range track.Artists {
				names = append(names, artist.Name)
				if artistURL == "" {
					artistURL = artist.ExternalURLs.Spotify
				}
			}
			metadata := domain.TrackMetadata{
				URI:        track.URI,
				Name:       track.Name,
				Artists:    strings.Join(names, ", "),
				ArtistURL:  artistURL,
				AlbumName:  track.Album.Name,
				SpotifyURL: track.ExternalURL.Spotify,
			}
			if len(track.Album.Images) > 0 {
				metadata.ImageURL = track.Album.Images[0].URL
			}
			out[track.URI] = metadata
		}
	}
	return out, nil
}

// TrackIDFromURI extracts the bare id from a "spotify:track:<id>" URI.
func TrackIDFromURI(uri string) string {
	const prefix = "spotify:track:"
	if !strings.HasPrefix(uri, prefix) {
		return ""
	}
	return strings.TrimPrefix(uri, prefix)
}
