package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/user/navilist/pkg/navidrome"
)

var ndSongFields = []struct {
	Value string
	Label string
	Type  string
	Group string
}{
	// Song
	{"title", "Title", "text", "Song"},
	{"artist", "Artist", "text", "Song"},
	{"albumartist", "Album Artist", "text", "Song"},
	{"album", "Album", "text", "Song"},
	{"genre", "Genre", "text", "Song"},
	{"year", "Year", "number", "Song"},
	{"tracknumber", "Track Number", "number", "Song"},
	{"discnumber", "Disc Number", "number", "Song"},
	{"discsubtitle", "Disc Subtitle", "text", "Song"},
	{"compilation", "Is Compilation", "boolean", "Song"},
	{"hascoverart", "Has Cover Art", "boolean", "Song"},
	// Dates
	{"date", "Recording Date", "date", "Dates"},
	{"originalyear", "Original Year", "number", "Dates"},
	{"originaldate", "Original Date", "date", "Dates"},
	{"releaseyear", "Release Year", "number", "Dates"},
	{"releasedate", "Release Date", "date", "Dates"},
	// Metadata
	{"comment", "Comment", "text", "Metadata"},
	{"lyrics", "Lyrics", "text", "Metadata"},
	{"grouping", "Grouping", "text", "Metadata"},
	{"sorttitle", "Sort Title", "text", "Metadata"},
	{"sortalbum", "Sort Album", "text", "Metadata"},
	{"sortartist", "Sort Artist", "text", "Metadata"},
	{"sortalbumartist", "Sort Album Artist", "text", "Metadata"},
	{"albumtype", "Album Type", "text", "Metadata"},
	{"albumcomment", "Album Comment", "text", "Metadata"},
	{"catalognumber", "Catalog Number", "text", "Metadata"},
	// File
	{"filepath", "File Path", "text", "File"},
	{"filetype", "File Type", "text", "File"},
	{"duration", "Duration (s)", "number", "File"},
	{"bitrate", "Bitrate", "number", "File"},
	{"bitdepth", "Bit Depth", "number", "File"},
	{"bpm", "BPM", "number", "File"},
	{"channels", "Channels", "number", "File"},
	{"size", "File Size", "number", "File"},
	// Library
	{"dateadded", "Date Added", "date", "Library"},
	{"datemodified", "Date Modified", "date", "Library"},
	{"lastplayed", "Last Played", "date", "Library"},
	{"playcount", "Play Count", "number", "Library"},
	{"rating", "Rating (0-5)", "number", "Library"},
	{"loved", "Is Favorite", "boolean", "Library"},
	{"dateloved", "Date Favorited", "date", "Library"},
	{"daterated", "Date Rated", "date", "Library"},
	// MusicBrainz
	{"mbz_album_id", "Album ID", "text", "MusicBrainz"},
	{"mbz_album_artist_id", "Album Artist ID", "text", "MusicBrainz"},
	{"mbz_artist_id", "Artist ID", "text", "MusicBrainz"},
	{"mbz_recording_id", "Recording ID", "text", "MusicBrainz"},
	{"mbz_release_track_id", "Release Track ID", "text", "MusicBrainz"},
	{"mbz_release_group_id", "Release Group ID", "text", "MusicBrainz"},
	// Other
	{"library_id", "Library ID", "number", "Other"},
}

func (h *Handler) NewSmart(w http.ResponseWriter, r *http.Request) {
	data := h.baseData("playlists")
	data["Playlist"] = navidrome.Playlist{}
	data["IsNew"] = true
	data["Fields"] = ndSongFields
	data["RulesJSON"] = `{"all":[]}`
	h.tpl.ExecuteTemplate(w, "smart_form.html", data)
}

func (h *Handler) CreateSmart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, "Invalid form", http.StatusBadRequest)
		return
	}
	rules, err := parseRulesFromForm(r)
	if err != nil {
		h.renderError(w, r, "Invalid rules: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := navidrome.CreatePlaylistRequest{
		Name:    r.FormValue("name"),
		Comment: r.FormValue("comment"),
		Rules:   rules,
	}
	p, err := h.nd.CreatePlaylist(r.Context(), req)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	http.Redirect(w, r, "/playlists/"+p.ID, http.StatusSeeOther)
}

func (h *Handler) EditSmart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.nd.GetPlaylist(r.Context(), id)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	rulesJSON := "{}"
	if p.Rules != nil {
		b, _ := json.MarshalIndent(p.Rules, "", "  ")
		rulesJSON = string(b)
	}
	data := h.baseData("playlists")
	data["Playlist"] = p
	data["IsNew"] = false
	data["Fields"] = ndSongFields
	data["RulesJSON"] = rulesJSON
	h.tpl.ExecuteTemplate(w, "smart_form.html", data)
}

func (h *Handler) UpdateSmart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, "Invalid form", http.StatusBadRequest)
		return
	}
	rules, err := parseRulesFromForm(r)
	if err != nil {
		h.renderError(w, r, "Invalid rules: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := navidrome.UpdatePlaylistRequest{
		Name:    r.FormValue("name"),
		Comment: r.FormValue("comment"),
		Rules:   rules,
	}
	if err := h.nd.UpdatePlaylist(r.Context(), id, req); err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	http.Redirect(w, r, "/playlists/"+id, http.StatusSeeOther)
}

func (h *Handler) SuggestField(w http.ResponseWriter, r *http.Request) {
	field := r.URL.Query().Get("field")
	q := r.URL.Query().Get("q")
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	var (
		results []string
		err     error
	)
	switch field {
	case "genre":
		results, err = h.nd.SearchGenres(r.Context(), q, 20)
	case "artist", "albumartist":
		results, err = h.nd.SearchArtists(r.Context(), q, 20)
	case "album":
		results, err = h.nd.SearchAlbums(r.Context(), q, 20)
	default:
		http.Error(w, "unknown field", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if results == nil {
		results = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func parseRulesFromForm(r *http.Request) (*navidrome.PlaylistRules, error) {
	var rules navidrome.PlaylistRules
	if err := json.Unmarshal([]byte(r.FormValue("rules_json")), &rules); err != nil {
		return nil, err
	}
	return &rules, nil
}
