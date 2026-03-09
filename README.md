# Navilist

<p align="center">
  <img src="docs/demo.gif" alt="" width="600">
</p>

<p align="center">
  <a href="https://github.com/fabienpiette/navilist/actions/workflows/ci.yml"><img src="https://github.com/fabienpiette/navilist/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

<h3 align="center">Self-hosted web UI for managing Navidrome playlists, M3U editor, smart rule builder, dedup, merge, and bulk ops. No filesystem access required.</h3>

---

## Quick Start

```bash
# Clone and configure
git clone https://github.com/fabienpiette/navilist.git
cd navilist
# Edit docker-compose.yml: set NAVIDROME_URL, NAVIDROME_USER, NAVIDROME_PASS

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

- **Smart playlist builder**, visual rule editor with 40+ filterable fields, live autocomplete for genre/artist/album, 12 preset templates, raw NSP JSON fallback
- **M3U playlist editor**, create and edit playlists with live search-as-you-type track search
- **Dedup**, find playlists with identical track lists, pick which to keep, delete the rest in one click
- **Merge**, combine two or more playlists into one (tracks deduplicated); optionally delete the sources
- **Smart rename**, suggest `Artist – Album` names based on track metadata; edit before applying
- **M3U import**, upload a `.m3u` file, get a match report (exact path → title fuzzy fallback), confirm to create
- **Batch ops**, delete selected, delete all empty, export multiple playlists as a `.zip`
- **Single binary**, all templates and static assets embedded; no volume mounts needed in Docker
- **Dark mode**, stored in `localStorage`; no cookies, no server round-trip

## Install

**Prerequisites:** a running [Navidrome](https://www.navidrome.org/) instance (v0.50+). Docker or Go 1.22+.

> **Note:** the Navidrome account must have **admin** privileges. Non-admin accounts can only see public playlists and their own.

### Docker Compose (recommended)

```yaml
services:
  navilist:
    image: ghcr.io/fabienpiette/navilist:latest
    container_name: navilist
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

**Playlist list** (`/`), filter by All / M3U / Smart, client-side search, inline delete, batch delete/export, merge selected, dedup selected, smart rename selected, delete all empty.

**New M3U** (`/playlists/new`), type a name, search tracks, reorder with drag, save.

**New Smart** (`/playlists/new/smart`), add rules visually or paste NSP JSON directly; set sort field and limit. In a rule row, set field to genre/artist/album for live autocomplete suggestions.

**Import** (`/import`), upload a `.m3u` file; review the match report; confirm to create the playlist.

**Merge**, check two or more playlists, click **merge selected**, name the result, optionally delete sources.

**Dedup**, check two or more playlists, click **find duplicates**; for each group of identical playlists choose which to keep.

**Smart rename**, check smart playlists, click **smart rename**; suggested names are derived from track metadata; edit before applying.

## Known Issues

- Smart playlists require Navidrome v0.50+ (earlier versions don't expose the `rules` field via `/api/playlist`).
- The M3U import fuzzy-match uses title search only; tracks with identical titles may match the wrong song.
- No authentication layer, intended for private/LAN use only. Do not expose port 8080 to the internet without a reverse proxy with auth.
- Genre autocomplete fetches the full genre list on each keystroke (Navidrome has no server-side genre filter); performance degrades with very large genre libraries.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Contributing](docs/CONTRIBUTING.md), dev setup, hook install, commit conventions
- [Navidrome Smart Playlist format](https://www.navidrome.org/docs/usage/features/smart-playlists/)

## Contributing

Contributions welcome. See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for dev setup and guidelines.

## Acknowledgments

Thanks to all [contributors](https://github.com/fabienpiette/navilist/graphs/contributors).

<p align="center">
<a href="https://buymeacoffee.com/fabienpiette" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" height="60"></a>
</p>

## License

[AGPL-3.0](docs/LICENSE), if you distribute a modified version, you must release its source under the same terms.