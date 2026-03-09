# Navilist

<p align="center">
  <img src="docs/demo.png" alt="" width="600">
</p>

<h3 align="center">Self-hosted web UI for managing Navidrome M3U playlists and Smart Playlists — no filesystem access required.</h3>

---

## Quick Start

```bash
# Clone and configure
git clone https://github.com/fabienpiette/navilist.git
cd navilist
cp docker-compose.yml docker-compose.override.yml
# Edit docker-compose.override.yml: set NAVIDROME_URL, NAVIDROME_USER, NAVIDROME_PASS

# Run
docker compose up -d

# Open
open http://localhost:8080
```

Or from source (Go 1.22+ required):

```bash
NAVIDROME_URL=http://localhost:4533 \
NAVIDROME_USER=admin \
NAVIDROME_PASS=yourpassword \
go run ./cmd/server
```

---

## Features

- **M3U playlist editor** — create and edit playlists with live search-as-you-type track search
- **Smart playlist builder** — visual rule editor (field / operator / value) with raw JSON fallback for the full NSP format
- **M3U import** — upload a `.m3u` file and get a match report: exact path lookup, then title fuzzy-search for unmatched entries
- **Merge playlists** — select two or more playlists, combine their tracks (deduplicated), optionally delete the sources
- **Batch ops** — delete selected, delete all empty, or export multiple playlists as a single `.zip` in one click
- **Dark mode** — theme toggle stored in `localStorage`; no cookies, no server round-trip
- **Single binary** — all templates and static assets embedded; no volume mounts needed in Docker

## Install

**Prerequisites:** a running [Navidrome](https://www.navidrome.org/) instance (v0.50+), Docker or Go 1.22+.

> **Note:** the Navidrome account used must have **admin** privileges. Non-admin accounts can only see public playlists and their own — not the full library.

### Docker Compose (recommended)

```yaml
services:
  navilist:
    build: .
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      NAVIDROME_URL: http://navidrome:4533
      NAVIDROME_USER: admin
      NAVIDROME_PASS: yourpassword
    networks:
      - navidrome-net

networks:
  navidrome-net:
    external: true
```

```bash
docker compose up -d
```

### From source

```bash
git clone https://github.com/fabienpiette/navilist.git
cd navilist
go build -o navilist ./cmd/server
NAVIDROME_URL=http://localhost:4533 NAVIDROME_USER=admin NAVIDROME_PASS=secret ./navilist
```

## Usage

```bash
# Environment variables (all required except PORT)
NAVIDROME_URL=http://navidrome:4533   # Navidrome base URL
NAVIDROME_USER=admin                  # Navidrome username (must be admin)
NAVIDROME_PASS=secret                 # Navidrome password
PORT=8080                             # Listening port (default: 8080)
```

**Playlist list** (`/`) — filter by All / M3U / Smart, inline delete, batch delete/export, merge selected, delete all empty.

**New M3U** (`/playlists/new`) — type a name, search tracks, reorder with drag, save.

**New Smart** (`/playlists/new/smart`) — add rules visually or paste NSP JSON directly; set sort field and limit.

**Import** (`/import`) — upload a `.m3u` file; review the match report; confirm to create the playlist.

**Merge** — check two or more playlists on the list page, click "merge selected", name the result, optionally delete sources.

## Known Issues

- Smart playlists require Navidrome v0.50+ (earlier versions don't expose the `rules` field via `/api/playlist`).
- The M3U import fuzzy-match uses title search only; tracks with identical titles may match the wrong song.
- No authentication layer — intended for private/LAN use only. Do not expose port 8080 to the internet without a reverse proxy with auth.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Navidrome Smart Playlist format](https://www.navidrome.org/docs/usage/features/smart-playlists/)

## Contributing

Contributions welcome. Open an issue first for non-trivial changes so we can align on direction before you write code.

## License

[MIT](LICENSE)
