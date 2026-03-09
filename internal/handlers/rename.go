package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/user/navilist/pkg/navidrome"
)

const purityThreshold = 0.8

type renameRow struct {
	ID          string
	CurrentName string
	Suggested   string
}

func suggestName(tracks []navidrome.Song) string {
	if len(tracks) == 0 {
		return ""
	}
	type key struct{ artist, album string }
	counts := map[key]int{}
	for _, t := range tracks {
		counts[key{t.Artist, t.Album}]++
	}
	var best key
	var bestCount int
	for k, n := range counts {
		if n > bestCount {
			best, bestCount = k, n
		}
	}
	if float64(bestCount)/float64(len(tracks)) >= purityThreshold {
		return best.artist + " - " + best.album
	}
	return ""
}

func (h *Handler) SmartRenameForm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]
	if len(ids) == 0 {
		h.renderError(w, r, "Select at least 1 playlist to rename", http.StatusBadRequest)
		return
	}
	log.Printf("smart-rename: computing suggestions for %d playlists", len(ids))

	var rows []renameRow
	for _, id := range ids {
		p, err := h.nd.GetPlaylist(r.Context(), id)
		if err != nil {
			log.Printf("smart-rename: get playlist id=%s: %v", id, err)
			continue
		}
		tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
		if err != nil {
			log.Printf("smart-rename: get tracks id=%s: %v", id, err)
			continue
		}
		rows = append(rows, renameRow{
			ID:          p.ID,
			CurrentName: p.Name,
			Suggested:   suggestName(tracks),
		})
	}

	data := h.baseData("playlists")
	data["Rows"] = rows
	h.tpl.ExecuteTemplate(w, "smart_rename_form.html", data)
}

func (h *Handler) SmartRenameConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["id"]
	names := r.Form["name"]
	currents := r.Form["current"]

	for i, id := range ids {
		if i >= len(names) || i >= len(currents) {
			break
		}
		newName := strings.TrimSpace(names[i])
		if newName == "" || newName == currents[i] {
			continue
		}
		if err := h.nd.UpdatePlaylist(r.Context(), id, navidrome.UpdatePlaylistRequest{Name: newName}); err != nil {
			log.Printf("smart-rename: update id=%s: %v", id, err)
		} else {
			log.Printf("smart-rename: renamed %q → %q", currents[i], newName)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) SuggestName(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
	if err != nil {
		log.Printf("suggest-name: get tracks id=%s: %v", id, err)
		w.WriteHeader(http.StatusOK)
		return
	}
	suggestion := suggestName(tracks)
	log.Printf("suggest-name: id=%s → %q", id, suggestion)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(suggestion))
}
