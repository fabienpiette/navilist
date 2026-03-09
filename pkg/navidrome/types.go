package navidrome

import "time"

// Playlist represents a Navidrome playlist (M3U or Smart).
// Rules == nil means M3U; Rules != nil means Smart Playlist.
type Playlist struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Comment   string         `json:"comment"`
	SongCount int            `json:"songCount"`
	Duration  float64        `json:"duration"`
	Public    bool           `json:"public"`
	Rules     *PlaylistRules `json:"rules,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func (p Playlist) IsSmart() bool { return p.Rules != nil }

// PlaylistRules is the NSP smart playlist JSON format.
type PlaylistRules struct {
	All   []Rule `json:"all,omitempty"`
	Any   []Rule `json:"any,omitempty"`
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// Rule is a single criterion: {"operator": {"field": "value"}}
type Rule map[string]any

// Song represents a Navidrome track.
type Song struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Album    string  `json:"album"`
	Duration float64 `json:"duration"`
	Path     string  `json:"path"`
	TrackID  string  `json:"playlistItemId,omitempty"`
}

// CreatePlaylistRequest is sent to POST /api/playlist.
type CreatePlaylistRequest struct {
	Name    string         `json:"name"`
	Comment string         `json:"comment"`
	Public  bool           `json:"public"`
	Rules   *PlaylistRules `json:"rules,omitempty"`
}

// UpdatePlaylistRequest is sent to PUT /api/playlist/:id.
type UpdatePlaylistRequest struct {
	Name    string         `json:"name"`
	Comment string         `json:"comment"`
	Public  bool           `json:"public"`
	Rules   *PlaylistRules `json:"rules,omitempty"`
}

// AddTracksRequest is sent to POST /api/playlist/:id/tracks.
type AddTracksRequest struct {
	IDs []string `json:"ids"`
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}
