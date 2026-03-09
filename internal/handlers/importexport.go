package handlers

import (
	"archive/zip"
	"fmt"
	"log"
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
	log.Printf("import: parsing %q", header.Filename)

	tracks, err := m3u.Parse(file)
	if err != nil {
		log.Printf("import: parse error: %v", err)
		h.renderError(w, r, "Failed to parse M3U: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("import: parsed %d tracks from %q", len(tracks), header.Filename)

	var results []matchResult
	matched := 0
	for _, t := range tracks {
		song, found, _ := h.nd.GetSongByPath(r.Context(), t.Path)
		if !found && t.Title != "" {
			songs, _ := h.nd.SearchSongs(r.Context(), t.Title, 1)
			if len(songs) > 0 {
				results = append(results, matchResult{t.Path, songs[0], true})
				matched++
				continue
			}
		}
		if found {
			matched++
		}
		results = append(results, matchResult{Path: t.Path, Song: song, Matched: found})
	}
	log.Printf("import: matched %d/%d tracks", matched, len(tracks))

	name := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	data := h.baseData("import")
	data["Results"] = results
	data["Name"] = name
	h.tpl.ExecuteTemplate(w, "import_form.html", data)
}

func (h *Handler) ImportConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	req := navidrome.CreatePlaylistRequest{Name: name}
	p, err := h.nd.CreatePlaylist(r.Context(), req)
	if err != nil {
		log.Printf("import confirm: create playlist %q: %v", name, err)
		h.renderError(w, r, err.Error(), http.StatusBadGateway)
		return
	}
	songIDs := r.Form["song_id"]
	if len(songIDs) > 0 {
		h.nd.AddTracks(r.Context(), p.ID, songIDs)
	}
	log.Printf("import confirm: created playlist %q (id=%s) with %d tracks", p.Name, p.ID, len(songIDs))
	http.Redirect(w, r, "/playlists/"+p.ID, http.StatusSeeOther)
}

func (h *Handler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]
	log.Printf("batch delete: deleting %d playlists", len(ids))
	for _, id := range ids {
		if err := h.nd.DeletePlaylist(r.Context(), id); err != nil {
			log.Printf("batch delete: id=%s: %v", id, err)
		}
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
		log.Printf("delete-empty: list playlists: %v", err)
		h.renderError(w, r, "Failed to load playlists: "+err.Error(), http.StatusBadGateway)
		return
	}
	deleted := 0
	for _, p := range playlists {
		if p.SongCount == 0 {
			if err := h.nd.DeletePlaylist(r.Context(), p.ID); err != nil {
				log.Printf("delete-empty: id=%s %q: %v", p.ID, p.Name, err)
			} else {
				deleted++
			}
		}
	}
	log.Printf("delete-empty: deleted %d empty playlists", deleted)
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
	log.Printf("merge: merging %d playlists into %q (delete_sources=%v)", len(ids), name, deleteSources)

	// Collect unique song IDs from all source playlists.
	seen := map[string]bool{}
	var songIDs []string
	for _, id := range ids {
		tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
		if err != nil {
			log.Printf("merge: get tracks id=%s: %v", id, err)
			continue
		}
		for _, t := range tracks {
			if !seen[t.ID] {
				seen[t.ID] = true
				songIDs = append(songIDs, t.ID)
			}
		}
	}
	log.Printf("merge: collected %d unique tracks", len(songIDs))

	newPl, err := h.nd.CreatePlaylist(r.Context(), navidrome.CreatePlaylistRequest{Name: name})
	if err != nil {
		log.Printf("merge: create playlist %q: %v", name, err)
		h.renderError(w, r, "Failed to create playlist: "+err.Error(), http.StatusBadGateway)
		return
	}

	if len(songIDs) > 0 {
		if err := h.nd.AddTracks(r.Context(), newPl.ID, songIDs); err != nil {
			log.Printf("merge: add tracks to %s: %v", newPl.ID, err)
			h.renderError(w, r, "Failed to add tracks: "+err.Error(), http.StatusBadGateway)
			return
		}
	}

	if deleteSources {
		for _, id := range ids {
			if err := h.nd.DeletePlaylist(r.Context(), id); err != nil {
				log.Printf("merge: delete source id=%s: %v", id, err)
			}
		}
	}
	log.Printf("merge: done, new playlist id=%s", newPl.ID)

	http.Redirect(w, r, "/playlists/"+newPl.ID, http.StatusSeeOther)
}

func (h *Handler) DedupForm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]
	if len(ids) < 2 {
		h.renderError(w, r, "Select at least 2 playlists to find duplicates", http.StatusBadRequest)
		return
	}
	log.Printf("dedup: scanning %d playlists", len(ids))

	byFingerprint := map[string][]navidrome.Playlist{}

	for _, id := range ids {
		p, err := h.nd.GetPlaylist(r.Context(), id)
		if err != nil {
			log.Printf("dedup: get playlist id=%s: %v", id, err)
			continue
		}
		tracks, err := h.nd.GetPlaylistTracks(r.Context(), id)
		if err != nil {
			log.Printf("dedup: get tracks id=%s %q: %v", id, p.Name, err)
			continue
		}
		tokens := make([]string, len(tracks))
		for i, t := range tracks {
			// Round duration to nearest 5s to tolerate minor encoding differences.
			dur := int(t.Duration/5+0.5) * 5
			tokens[i] = strings.ToLower(t.Title) + "|" + strings.ToLower(t.Artist) + "|" + fmt.Sprintf("%d", dur)
		}
		sort.Strings(tokens)
		fp := strings.Join(tokens, ",")
		log.Printf("dedup: %q — %d tracks fingerprinted by title+artist+duration", p.Name, len(tracks))
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
		log.Printf("dedup: no duplicate groups found")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("dedup: found %d duplicate group(s)", len(groups))

	data := h.baseData("playlists")
	data["Groups"] = groups
	h.tpl.ExecuteTemplate(w, "dedup_form.html", data)
}

func (h *Handler) DedupConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["delete"]
	log.Printf("dedup confirm: deleting %d duplicate playlists", len(ids))
	for _, id := range ids {
		if err := h.nd.DeletePlaylist(r.Context(), id); err != nil {
			log.Printf("dedup confirm: delete id=%s: %v", id, err)
		} else {
			log.Printf("dedup confirm: deleted id=%s", id)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
