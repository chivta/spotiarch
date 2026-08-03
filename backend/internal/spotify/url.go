package spotify

import (
	"net/url"
	"regexp"
	"strings"
)

// playlistIDPattern matches Spotify's base-62 resource ids.
var playlistIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{22}$`)

// PlaylistURL builds the public web link for a playlist.
func PlaylistURL(playlistID string) string {
	return "https://open.spotify.com/playlist/" + playlistID
}

// PlaylistIDFromURL accepts any of the forms a user is likely to paste: an
// open.spotify.com link (with or without query string or locale prefix), a
// "spotify:playlist:<id>" URI, or a bare id. It returns "" when nothing matches.
func PlaylistIDFromURL(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	if playlistIDPattern.MatchString(input) {
		return input
	}

	if id, ok := strings.CutPrefix(input, "spotify:playlist:"); ok {
		if playlistIDPattern.MatchString(id) {
			return id
		}
		return ""
	}

	parsed, err := url.Parse(input)
	if err != nil {
		return ""
	}
	if !strings.HasSuffix(parsed.Host, "spotify.com") {
		return ""
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i, segment := range segments {
		if segment == "playlist" && i+1 < len(segments) {
			if playlistIDPattern.MatchString(segments[i+1]) {
				return segments[i+1]
			}
			return ""
		}
	}
	return ""
}
