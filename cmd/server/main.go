package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/user/navidrome-playlists/internal/handlers"
	"github.com/user/navidrome-playlists/pkg/navidrome"
	"github.com/user/navidrome-playlists/web"
)

func main() {
	navidromeURL := mustEnv("NAVIDROME_URL")
	navidromeUser := mustEnv("NAVIDROME_USER")
	navidromePass := mustEnv("NAVIDROME_PASS")
	port := envOr("PORT", "8080")

	nd := navidrome.New(navidromeURL, navidromeUser, navidromePass)
	if err := nd.Authenticate(); err != nil {
		log.Fatalf("navidrome auth: %v", err)
	}

	tpl := template.Must(template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFS(web.Files,
		"templates/base.html",
		"templates/playlist_list.html",
		"templates/playlist_detail.html",
		"templates/playlist_form.html",
		"templates/smart_form.html",
		"templates/import_form.html",
		"templates/partials/toast.html",
		"templates/partials/search_results.html",
	))

	h := handlers.New(nd, tpl)

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
	r.Post("/playlists/batch/export", h.BatchExport)
	r.Get("/playlists/{id}", h.Detail)
	r.Get("/playlists/{id}/edit", h.EditPlaylist)
	r.Post("/playlists/{id}", h.UpdatePlaylist)
	r.Get("/playlists/{id}/edit/smart", h.EditSmart)
	r.Post("/playlists/{id}/smart", h.UpdateSmart)
	r.Delete("/playlists/{id}", h.Delete)
	r.Get("/playlists/{id}/export", h.Export)
	r.Get("/search", h.Search)
	r.Get("/import", h.ImportForm)
	r.Post("/import", h.Import)
	r.Post("/import/confirm", h.ImportConfirm)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
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
