package handlers

import (
	"archive/zip"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/user/navidrome-playlists/internal/m3u"
	"github.com/user/navidrome-playlists/pkg/navidrome"
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
	h.tpl.ExecuteTemplate(w, "import_form.html", map[string]any{
		"ActiveTab": "import",
	})
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
	h.tpl.ExecuteTemplate(w, "import_form.html", map[string]any{
		"ActiveTab": "import",
		"Results":   results,
		"Name":      name,
	})
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
