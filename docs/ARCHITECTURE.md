# Architecture

This document describes the high-level architecture of navilist.
If you want to familiarize yourself with the codebase, you are in the right place.

## Bird's Eye View

navilist is a thin HTTP proxy that translates browser requests into Navidrome REST API calls and renders the results as server-side HTML. There is no local database; all playlist state lives in Navidrome.

A request arrives at the chi router in `cmd/server/main.go`, is dispatched to a handler in `internal/handlers/`, and the handler calls `pkg/navidrome/` to talk to Navidrome's `/api/` endpoints. The handler then renders a `html/template` template from `web/templates/` and writes HTML back to the browser. HTMX handles partial page updates by targeting specific DOM elements; the server detects HTMX requests via the `HX-Request` header and responds with template fragments rather than full pages.

The only stateful concern is the JWT token used to authenticate against Navidrome. It is held in memory on the `navidrome.Client` struct and refreshed automatically before expiry.

```
Browser ──HTMX──▶ chi router ──▶ handlers ──▶ navidrome.Client ──▶ Navidrome /api/
                                     │
                                     └──▶ m3u.Parse / m3u.Write   (import/export only)
                                     │
                                     └──▶ html/template (web/templates/)
```

## Code Map

### `cmd/server/`

Entrypoint. Reads env vars (`NAVIDROME_URL`, `NAVIDROME_USER`, `NAVIDROME_PASS`, `PORT`), authenticates, builds the per-page template sets via `buildTemplates()`, wires the chi router, and starts the HTTP server.

Key file: `main.go`.

**Architecture Invariant:** `cmd/server/` is the only package that imports both `internal/handlers/` and `pkg/navidrome/`. It owns wiring; neither package knows about the other.

---

### `web/`

Owns the embedded filesystem. A single `//go:embed templates static` directive in `web.go` bundles all HTML templates and static assets into the binary at compile time.

Key file: `web.go` — exports `Files embed.FS`.

**Architecture Invariant:** `web/` contains no Go logic beyond the embed declaration. Template rendering lives in `internal/handlers/`; asset serving is wired in `cmd/server/`.

**Why a separate package:** Go's `//go:embed` cannot use `..` path segments. Since the entry point is in `cmd/server/`, a dedicated package at the repo root is required to embed `templates/` and `static/` as siblings.

---

### `internal/handlers/`

HTTP handler layer. Each file groups related routes:

| File | Routes |
|------|--------|
| `handler.go` | `Handler` struct, `Templates` dispatcher, `New()`, `renderError()` |
| `playlists.go` | `List`, `Detail`, `Delete`, `Search` |
| `edit.go` | `NewPlaylist`, `CreatePlaylist`, `EditPlaylist`, `UpdatePlaylist` |
| `smart.go` | `NewSmart`, `CreateSmart`, `EditSmart`, `UpdateSmart`, `parseRulesFromForm` |
| `importexport.go` | `Import`, `ImportConfirm`, `ImportForm`, `Export`, `BatchDelete`, `BatchExport`, `DeleteEmpty`, `MergeForm`, `MergeConfirm` |

`renderError` is the only shared helper: it emits an `HX-Trigger: {"showToast":"..."}` header for HTMX partial requests and falls back to `http.Error` for full-page requests.

`Templates` is a dispatcher type that routes `ExecuteTemplate(w, name, data)` calls to the correct per-page cloned template set (see Template Isolation below).

**Architecture Invariant:** handlers never call `html/template` directly — they always go through `h.tpl.ExecuteTemplate`. This keeps template name coupling in one place per handler, not scattered across the file.

**Architecture Invariant:** handlers never construct raw JSON or serialize to non-HTML formats except via `internal/m3u` (for `.m3u` export) and `archive/zip` (for batch export). All Navidrome data marshalling lives in `pkg/navidrome/`.

---

### `pkg/navidrome/`

The Navidrome REST API client. Handles JWT authentication, token refresh, and all HTTP communication with Navidrome's `/api/` endpoints.

| File | Responsibility |
|------|---------------|
| `types.go` | All request/response types. `Playlist.IsSmart()` returns `true` when `Rules != nil`. |
| `client.go` | `Client` struct, `Authenticate()`, `ensureToken()`, `Do()` (the one authenticated HTTP method) |
| `playlists.go` | `ListPlaylists`, `GetPlaylist`, `GetPlaylistTracks`, `CreatePlaylist`, `UpdatePlaylist`, `DeletePlaylist`, `AddTracks`, `RemoveTracks` |
| `songs.go` | `SearchSongs`, `GetSongByPath` |

Key types: `Client`, `Playlist`, `PlaylistRules`, `Rule`, `Song`.

**Authentication note:** Navidrome's REST API uses `X-ND-Authorization: Bearer <token>` (not the standard `Authorization` header). The auth endpoint is `POST /auth/login` (outside the `/api/` prefix). All other endpoints are under `/api/`.

**Architecture Invariant:** `pkg/navidrome/` has no knowledge of HTTP request context beyond `context.Context`. It never reads `http.Request` fields, headers, or form values.

---

### `internal/m3u/`

Stateless M3U/EXTM3U parser and writer. No dependencies on any other internal package.

`Parse(io.Reader) ([]Track, error)` reads an M3U file line by line. It handles `#EXTM3U` headers, `#EXTINF:duration,Artist - Title` metadata lines, bare path lines, and Windows `\r\n` line endings. The scanner buffer is 1 MiB to accommodate URL-based playlists with long lines.

`Write(io.Writer, []WriteTrack) error` renders a valid EXTM3U file.

Key types: `Track`, `WriteTrack`.

**Architecture Invariant:** `internal/m3u/` is pure I/O transformation. It never calls the Navidrome client and never touches HTTP.

## Invariants

1. **No local state.** The application stores nothing on disk and holds no in-memory cache of playlist data. Every page load fetches fresh data from Navidrome.

2. **JWT is the only shared mutable state.** `Client.token` and `Client.tokenExp` are guarded by `Client.mu` (`sync.Mutex`). No other shared state exists.

3. **`pkg/` does not import `internal/`.** Standard Go module visibility rules enforce this, but it is worth stating: `pkg/navidrome/` is a reusable library; `internal/` is application code.

4. **All user-supplied strings reach the browser only through `html/template`.** Go's `html/template` applies context-aware escaping automatically. Handlers never write raw user data to `http.ResponseWriter` with `fmt.Fprintf` or similar.

5. **Smart playlist rules round-trip through `navidrome.PlaylistRules`.** The visual form builder and the JSON textarea both serialize to `PlaylistRules` before any Navidrome call. Neither the handler nor the template touches raw rule JSON from user input directly.

## Cross-Cutting Concerns

**Error handling.** All handler errors go through `renderError(w, r, msg, code)`. For HTMX requests (`HX-Request` header present) it sends a toast trigger header with no body. For full-page requests it falls back to `http.Error`. Errors from `pkg/navidrome/` are passed through as-is; no wrapping is added at the handler layer.

**Configuration.** All configuration is via environment variables read at startup in `cmd/server/main.go`. There is no config file. Missing required vars (`NAVIDROME_URL`, `NAVIDROME_USER`, `NAVIDROME_PASS`) cause a fatal exit with a descriptive message.

**Testing.** Tests live in `pkg/navidrome/client_test.go` (4 tests, `package navidrome_test`) and `internal/m3u/parser_test.go` (6 tests, `package m3u_test`). Handler tests are not present — the handlers are thin wrappers and their correctness is verified by integration against a live Navidrome instance. The `navidrome` tests use `httptest.NewServer` as a mock.

**Template isolation.** Go's `html/template` uses a flat namespace: all `{{define "name"}}` blocks across all files parsed into the same set share the same name registry, so the last file parsed wins. To prevent every page from rendering the same `content` block, `buildTemplates()` in `cmd/server/main.go` clones the base template set once per page using `template.Clone()` and parses each page into its own clone. The `Templates` dispatcher in `handler.go` routes `ExecuteTemplate(w, name, data)` calls to the correct clone by template name.

## A Typical Change

**Adding a new playlist field** (e.g., exposing `public: bool` in the M3U create form):

1. `pkg/navidrome/types.go` — confirm `CreatePlaylistRequest.Public bool` exists (it does).
2. `internal/handlers/edit.go` — read `r.FormValue("public") == "on"` and set `req.Public` in `CreatePlaylist` and `UpdatePlaylist`.
3. `web/templates/playlist_form.html` — add `<input type="checkbox" name="public" ...>` to the form, checking `.Playlist.Public` for the edit case.
4. Run `go test ./...` and smoke-test against a live Navidrome instance.

No other files need to change. The `navidrome.Client` sends whatever fields are set in the request struct; Navidrome ignores absent fields.

**Adding a new batch operation** (e.g., the merge feature):

1. `internal/handlers/importexport.go` — add handler functions.
2. `cmd/server/main.go` — register routes and add any new page template to `buildTemplates()`.
3. `web/templates/` — add the new page template; add a trigger button to `playlist_list.html` if needed.
