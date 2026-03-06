package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/user/navidrome-playlists/pkg/navidrome"
)

var ndSongFields = []struct {
	Value string
	Label string
	Type  string
}{
	{"title", "Title", "text"},
	{"album", "Album", "text"},
	{"artist", "Artist", "text"},
	{"albumArtist", "Album Artist", "text"},
	{"genre", "Genre", "text"},
	{"year", "Year", "number"},
	{"rating", "Rating", "number"},
	{"playCount", "Play Count", "number"},
	{"loved", "Loved", "boolean"},
	{"lastPlayed", "Last Played", "date"},
	{"dateAdded", "Date Added", "date"},
}

func (h *Handler) NewSmart(w http.ResponseWriter, r *http.Request) {
	h.tpl.ExecuteTemplate(w, "smart_form.html", map[string]any{
		"ActiveTab": "playlists",
		"Playlist":  navidrome.Playlist{},
		"IsNew":     true,
		"Fields":    ndSongFields,
		"RulesJSON": `{"all":[]}`,
	})
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
	h.tpl.ExecuteTemplate(w, "smart_form.html", map[string]any{
		"ActiveTab": "playlists",
		"Playlist":  p,
		"IsNew":     false,
		"Fields":    ndSongFields,
		"RulesJSON": rulesJSON,
	})
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

func parseRulesFromForm(r *http.Request) (*navidrome.PlaylistRules, error) {
	if r.FormValue("editor_mode") == "json" {
		var rules navidrome.PlaylistRules
		if err := json.Unmarshal([]byte(r.FormValue("rules_json")), &rules); err != nil {
			return nil, err
		}
		return &rules, nil
	}
	fields := r.Form["rule_field"]
	ops := r.Form["rule_op"]
	values := r.Form["rule_value"]
	var all []navidrome.Rule
	for i := range fields {
		if i >= len(ops) || i >= len(values) {
			break
		}
		all = append(all, navidrome.Rule{ops[i]: map[string]any{fields[i]: values[i]}})
	}
	rules := &navidrome.PlaylistRules{All: all}
	if l := r.FormValue("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			rules.Limit = n
		}
	}
	if s := r.FormValue("sort"); s != "" {
		rules.Sort = s
	}
	return rules, nil
}
