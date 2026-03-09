package handlers

import (
	"archive/zip"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/user/navilist/internal/m3u"
	"github.com/user/navilist/pkg/navidrome"
)

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.nd.GetPlaylist(r.Context(), id)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
	if err != nil {
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	filename := sanitizeFilename(p.Name) + ".m3u"
	w.Header().Set("Content-Type", "audio/x-mpegurl")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	var wt []m3u.WriteTrack
	for _, t := range tracks {
		wt = append(wt, m3u.WriteTrack{Path: t.Path, Title: t.Title, Artist: t.Artist, Duration: int(t.Duration)})
	}
	m3u.Write(w, wt)
}

func (h *Handler) ImportForm(w http.ResponseWriter, r *http.Request) {
	h.tpl.ExecuteTemplate(w, "import_form.html", h.baseData("import"))
}

type matchResult struct {
	Path    string
	Song    navidrome.Song
	Matched bool
}

func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("m3u_file")
	if err != nil {
		h.renderError(w, r, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tracks, err := m3u.Parse(file)
	if err != nil {
		h.renderError(w, r, "Failed to parse M3U: "+err.Error(), http.StatusBadRequest)
		return
	}

	var results []matchResult
	for _, t := range tracks {
		song, found, _ := h.nd.GetSongByPath(r.Context(), t.Path)
		if !found && t.Title != "" {
			songs, _ := h.nd.SearchSongs(r.Context(), t.Title, 1)
			if len(songs) > 0 {
				results = append(results, matchResult{t.Path, songs[0], true})
				continue
			}
		}
		results = append(results, matchResult{Path: t.Path, Song: song, Matched: found})
	}

	name := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	data := h.baseData("import")
	data["Results"] = results
	data["Name"] = name
	h.tpl.ExecuteTemplate(w, "import_form.html", data)
}

func (h *Handler) ImportConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	req := navidrome.CreatePlaylistRequest{Name: r.FormValue("name")}
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

func (h *Handler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for _, id := range r.Form["ids"] {
		h.nd.DeletePlaylist(r.Context(), id)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) BatchExport(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="playlists.zip"`)
	zw := zip.NewWriter(w)
	defer zw.Close()
	for _, id := range r.Form["ids"] {
		p, err := h.nd.GetPlaylist(r.Context(), id)
		if err != nil {
			continue
		}
		tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
		if err != nil {
			continue
		}
		fw, err := zw.Create(sanitizeFilename(p.Name) + ".m3u")
		if err != nil {
			continue
		}
		var wt []m3u.WriteTrack
		for _, t := range tracks {
			wt = append(wt, m3u.WriteTrack{Path: t.Path, Title: t.Title, Artist: t.Artist, Duration: int(t.Duration)})
		}
		m3u.Write(fw, wt)
	}
}

func sanitizeFilename(name string) string {
	r := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", `"`, "-", "<", "-", ">", "-", "|", "-")
	return r.Replace(name)
}

func (h *Handler) DeleteEmpty(w http.ResponseWriter, r *http.Request) {
	playlists, err := h.nd.ListPlaylists(r.Context())
	if err != nil {
		h.renderError(w, r, "Failed to load playlists: "+err.Error(), http.StatusBadGateway)
		return
	}
	for _, p := range playlists {
		if p.SongCount == 0 {
			h.nd.DeletePlaylist(r.Context(), p.ID)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type playlistSummary struct {
	ID        string
	Name      string
	SongCount int
}

type dedupGroup struct {
	SongCount int
	Playlists []navidrome.Playlist
}

func (h *Handler) MergeForm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]
	if len(ids) < 2 {
		h.renderError(w, r, "Select at least 2 playlists to merge", http.StatusBadRequest)
		return
	}

	var playlists []playlistSummary
	total := 0
	for _, id := range ids {
		p, err := h.nd.GetPlaylist(r.Context(), id)
		if err != nil {
			continue
		}
		playlists = append(playlists, playlistSummary{ID: p.ID, Name: p.Name, SongCount: p.SongCount})
		total += p.SongCount
	}

	name := ""
	if len(playlists) > 0 {
		name = playlists[0].Name + " (merged)"
	}

	data := h.baseData("playlists")
	data["Playlists"] = playlists
	data["Total"] = total
	data["Name"] = name
	h.tpl.ExecuteTemplate(w, "merge_form.html", data)
}

func (h *Handler) MergeConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]
	name := strings.TrimSpace(r.FormValue("name"))
	deleteSources := r.FormValue("delete_sources") == "on"

	if len(ids) < 2 || name == "" {
		h.renderError(w, r, "Invalid merge request", http.StatusBadRequest)
		return
	}

	// Collect unique song IDs from all source playlists.
	seen := map[string]bool{}
	var songIDs []string
	for _, id := range ids {
		tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
		if err != nil {
			continue
		}
		for _, t := range tracks {
			if !seen[t.ID] {
				seen[t.ID] = true
				songIDs = append(songIDs, t.ID)
			}
		}
	}

	newPl, err := h.nd.CreatePlaylist(r.Context(), navidrome.CreatePlaylistRequest{Name: name})
	if err != nil {
		h.renderError(w, r, "Failed to create playlist: "+err.Error(), http.StatusBadGateway)
		return
	}

	if len(songIDs) > 0 {
		if err := h.nd.AddTracks(r.Context(), newPl.ID, songIDs); err != nil {
			h.renderError(w, r, "Failed to add tracks: "+err.Error(), http.StatusBadGateway)
			return
		}
	}

	if deleteSources {
		for _, id := range ids {
			h.nd.DeletePlaylist(r.Context(), id)
		}
	}

	http.Redirect(w, r, "/playlists/"+newPl.ID, http.StatusSeeOther)
}

func (h *Handler) DedupForm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]
	if len(ids) < 2 {
		h.renderError(w, r, "Select at least 2 playlists to find duplicates", http.StatusBadRequest)
		return
	}

	byFingerprint := map[string][]navidrome.Playlist{}

	for _, id := range ids {
		p, err := h.nd.GetPlaylist(r.Context(), id)
		if err != nil {
			continue
		}
		tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
		if err != nil {
			continue
		}
		songIDs := make([]string, len(tracks))
		for i, t := range tracks {
			songIDs[i] = t.ID
		}
		sort.Strings(songIDs)
		fp := strings.Join(songIDs, ",")
		byFingerprint[fp] = append(byFingerprint[fp], p)
	}

	var groups []dedupGroup
	for _, playlists := range byFingerprint {
		if len(playlists) > 1 {
			groups = append(groups, dedupGroup{
				SongCount: playlists[0].SongCount,
				Playlists: playlists,
			})
		}
	}

	if len(groups) == 0 {
		h.renderError(w, r, "No duplicate playlists found", http.StatusOK)
		return
	}

	data := h.baseData("playlists")
	data["Groups"] = groups
	h.tpl.ExecuteTemplate(w, "dedup_form.html", data)
}

func (h *Handler) DedupConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for _, id := range r.Form["delete"] {
		h.nd.DeletePlaylist(r.Context(), id)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
