package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/user/navilist/pkg/navidrome"
)

func (h *Handler) NewPlaylist(w http.ResponseWriter, r *http.Request) {
	data := h.baseData("playlists")
	data["Playlist"] = navidrome.Playlist{}
	data["Tracks"] = nil
	data["IsNew"] = true
	h.tpl.ExecuteTemplate(w, "playlist_form.html", data)
}

func (h *Handler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, "Invalid form", http.StatusBadRequest)
		return
	}
	req := navidrome.CreatePlaylistRequest{
		Name:    r.FormValue("name"),
		Comment: r.FormValue("comment"),
	}
	p, err := h.nd.CreatePlaylist(r.Context(), req)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	if songIDs := r.Form["song_id"]; len(songIDs) > 0 {
		h.nd.AddTracks(r.Context(), p.ID, songIDs)
	}
	http.Redirect(w, r, "/playlists/"+p.ID, http.StatusSeeOther)
}

func (h *Handler) EditPlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.nd.GetPlaylist(r.Context(), id)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	tracks, _ := h.nd.GetPlaylistTracks(r.Context(), id)
	data := h.baseData("playlists")
	data["Playlist"] = p
	data["Tracks"] = tracks
	data["IsNew"] = false
	h.tpl.ExecuteTemplate(w, "playlist_form.html", data)
}

func (h *Handler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, "Invalid form", http.StatusBadRequest)
		return
	}
	req := navidrome.UpdatePlaylistRequest{
		Name:    r.FormValue("name"),
		Comment: r.FormValue("comment"),
	}
	if err := h.nd.UpdatePlaylist(r.Context(), id, req); err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	// Replace tracks: remove all then add new set.
	existing, _ := h.nd.GetPlaylistTracks(r.Context(), id)
	if len(existing) > 0 {
		var ids []string
		for _, t := range existing {
			ids = append(ids, t.TrackID)
		}
		h.nd.RemoveTracks(r.Context(), id, ids)
	}
	if songIDs := r.Form["song_id"]; len(songIDs) > 0 {
		h.nd.AddTracks(r.Context(), id, songIDs)
	}
	http.Redirect(w, r, "/playlists/"+id, http.StatusSeeOther)
}
