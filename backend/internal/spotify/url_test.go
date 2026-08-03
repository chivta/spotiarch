package spotify

import "testing"

func TestPlaylistIDFromURL(t *testing.T) {
	const id = "37i9dQZF1DXcBWIGoYBM5M"

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"bare id", id, id},
		{"uri", "spotify:playlist:" + id, id},
		{"web link", "https://open.spotify.com/playlist/" + id, id},
		{"web link with query", "https://open.spotify.com/playlist/" + id + "?si=abc123", id},
		{"locale prefix", "https://open.spotify.com/intl-de/playlist/" + id, id},
		{"surrounding whitespace", "  https://open.spotify.com/playlist/" + id + "  ", id},
		{"album link", "https://open.spotify.com/album/" + id, ""},
		{"foreign host", "https://example.com/playlist/" + id, ""},
		{"truncated id", "https://open.spotify.com/playlist/tooshort", ""},
		{"uri with bad id", "spotify:playlist:tooshort", ""},
		{"empty", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PlaylistIDFromURL(tc.input); got != tc.want {
				t.Errorf("PlaylistIDFromURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestTrackIDFromURI(t *testing.T) {
	cases := map[string]string{
		"spotify:track:4uLU6hMCjMI75M1A2tKUQC": "4uLU6hMCjMI75M1A2tKUQC",
		"spotify:episode:4uLU6hMCjMI75M1A2tKU": "",
		"4uLU6hMCjMI75M1A2tKUQC":               "",
		"":                                     "",
	}

	for input, want := range cases {
		if got := TrackIDFromURI(input); got != want {
			t.Errorf("TrackIDFromURI(%q) = %q, want %q", input, got, want)
		}
	}
}
