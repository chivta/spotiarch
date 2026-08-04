package spotify

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

type Owner struct {
	ID           string       `json:"id"`
	DisplayName  string       `json:"display_name"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Artist struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Album struct {
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

type Track struct {
	ID          string       `json:"id"`
	URI         string       `json:"uri"`
	Name        string       `json:"name"`
	Album       Album        `json:"album"`
	Artists     []Artist     `json:"artists"`
	ExternalIDs ExternalIDs  `json:"external_ids"`
	ExternalURL ExternalURLs `json:"external_urls"`
}

type ExternalIDs struct {
	ISRC string `json:"isrc"`
}

type PlaylistItem struct {
	IsLocal bool   `json:"is_local"`
	Track   *Track `json:"track"`
}

type PlaylistItemsPage struct {
	Items []PlaylistItem `json:"items"`
	Total int            `json:"total"`
	Next  *string        `json:"next"`
}

type PlaylistResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Public       bool              `json:"public"`
	SnapshotID   string            `json:"snapshot_id"`
	Images       []Image           `json:"images"`
	Owner        Owner             `json:"owner"`
	ExternalURLs ExternalURLs      `json:"external_urls"`
	Tracks       PlaylistItemsPage `json:"tracks"`
}

type tracksResponse struct {
	Tracks []*Track `json:"tracks"`
}

type meResponse struct {
	ID string `json:"id"`
}
