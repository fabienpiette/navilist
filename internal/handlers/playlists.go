package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/user/navidrome-playlists/pkg/navidrome"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	playlists, err := h.nd.ListPlaylists(r.Context())
	if err != nil {
		h.renderError(w, r, "Failed to load playlists: "+err.Error(), http.StatusBadGateway)
		return
	}

	var filtered []navidrome.Playlist
	for _, p := range playlists {
		switch filter {
		case "smart":
			if p.IsSmart() {
				filtered = append(filtered, p)
			}
		case "m3u":
			if !p.IsSmart() {
				filtered = append(filtered, p)
			}
		default:
			filtered = append(filtered, p)
		}
	}

	total := len(playlists)
	smart := 0
	for _, p := range playlists {
		if p.IsSmart() {
			smart++
		}
	}

	data := map[string]any{
		"ActiveTab": "playlists",
		"Filter":    filter,
		"Playlists": filtered,
		"Total":     total,
		"Smart":     smart,
		"Stats":     fmt.Sprintf("%d playlists · %d smart", total, smart),
	}

	if r.Header.Get("HX-Target") == "playlist-table" {
		h.tpl.ExecuteTemplate(w, "playlist_table", data)
		return
	}
	h.tpl.ExecuteTemplate(w, "playlist_list.html", data)
}

func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.nd.GetPlaylist(r.Context(), id)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	tracks, _ := h.nd.GetPlaylistTracks(r.Context(), id)
	h.tpl.ExecuteTemplate(w, "playlist_detail.html", map[string]any{
		"ActiveTab": "playlists",
		"Playlist":  p,
		"Tracks":    tracks,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/playlists/")
		id = strings.TrimSuffix(id, "/delete")
	}
	if err := h.nd.DeletePlaylist(r.Context(), id); err != nil {
		h.renderError(w, r, "Failed to delete: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		h.tpl.ExecuteTemplate(w, "search_results", nil)
		return
	}
	songs, err := h.nd.SearchSongs(r.Context(), q, 20)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	h.tpl.ExecuteTemplate(w, "search_results", songs)
}
