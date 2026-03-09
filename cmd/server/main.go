package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/user/navilist/internal/handlers"
	"github.com/user/navilist/pkg/navidrome"
	"github.com/user/navilist/web"
)

// version is set at build time via ldflags: -X main.version=v1.2.3
var version = "dev"

func main() {
	navidromeURL := mustEnv("NAVIDROME_URL")
	navidromeUser := mustEnv("NAVIDROME_USER")
	navidromePass := mustEnv("NAVIDROME_PASS")
	port := envOr("PORT", "8080")

	nd := navidrome.New(navidromeURL, navidromeUser, navidromePass)
	if err := nd.Authenticate(); err != nil {
		log.Fatalf("navidrome auth: %v", err)
	}

	tpl := buildTemplates()
	h := handlers.New(nd, tpl, version)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	staticSub, _ := fs.Sub(web.Files, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	r.Get("/", h.List)
	r.Get("/playlists/new", h.NewPlaylist)
	r.Post("/playlists", h.CreatePlaylist)
	r.Get("/playlists/new/smart", h.NewSmart)
	r.Post("/playlists/smart", h.CreateSmart)
	r.Post("/playlists/batch/delete", h.BatchDelete)
	r.Post("/playlists/batch/delete-empty", h.DeleteEmpty)
	r.Post("/playlists/batch/export", h.BatchExport)
	r.Post("/playlists/merge", h.MergeForm)
	r.Post("/playlists/merge/confirm", h.MergeConfirm)
	r.Post("/playlists/dedup", h.DedupForm)
	r.Post("/playlists/dedup/confirm", h.DedupConfirm)
	r.Post("/playlists/smart-rename", h.SmartRenameForm)
	r.Post("/playlists/smart-rename/confirm", h.SmartRenameConfirm)
	r.Get("/playlists/{id}/suggest-name", h.SuggestName)
	r.Get("/playlists/{id}", h.Detail)
	r.Get("/playlists/{id}/edit", h.EditPlaylist)
	r.Post("/playlists/{id}", h.UpdatePlaylist)
	r.Get("/playlists/{id}/edit/smart", h.EditSmart)
	r.Post("/playlists/{id}/smart", h.UpdateSmart)
	r.Delete("/playlists/{id}", h.Delete)
	r.Get("/playlists/{id}/export", h.Export)
	r.Get("/smart/suggest", h.SuggestField)
	r.Get("/search", h.Search)
	r.Get("/import", h.ImportForm)
	r.Post("/import", h.Import)
	r.Post("/import/confirm", h.ImportConfirm)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// buildTemplates parses each page template into its own cloned set so that
// each page's {{define "content"}} block is isolated and does not overwrite
// the others (Go templates use a flat namespace within a single set).
func buildTemplates() *handlers.Templates {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	// Base set: skeleton + shared partials only. No "content" defined here.
	base := template.Must(template.New("").Funcs(funcMap).ParseFS(web.Files,
		"templates/base.html",
		"templates/partials/toast.html",
		"templates/partials/search_results.html",
	))

	// newPage clones the base and adds a single page template into the clone.
	newPage := func(page string) *template.Template {
		c := template.Must(base.Clone())
		return template.Must(c.ParseFS(web.Files, "templates/"+page))
	}

	listSet := newPage("playlist_list.html")

	return handlers.NewTemplates(map[string]*template.Template{
		// Full-page templates — each in its own isolated clone.
		"playlist_list.html":     listSet,
		"playlist_detail.html":   newPage("playlist_detail.html"),
		"playlist_form.html":     newPage("playlist_form.html"),
		"smart_form.html":        newPage("smart_form.html"),
		"import_form.html":       newPage("import_form.html"),
		"merge_form.html":        newPage("merge_form.html"),
		"dedup_form.html":        newPage("dedup_form.html"),
		"smart_rename_form.html": newPage("smart_rename_form.html"),
		// Partials — "playlist_table" is defined inside playlist_list.html;
		// "search_results" is already in the base set.
		"playlist_table": listSet,
		"search_results": base,
	})
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
