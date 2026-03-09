# Contributing

Thanks for your interest. Here's everything you need to get started.

## Before You Start

Open an issue first for non-trivial changes. This avoids wasted effort if the direction doesn't fit the project.

Bug fixes and documentation improvements can go straight to a PR.

## Setup

**Prerequisites:** Go 1.23+, a running [Navidrome](https://www.navidrome.org/) instance.

```bash
git clone https://github.com/fabienpiette/navilist.git
cd navilist

# Copy and fill in your Navidrome credentials
cp .env.example .env   # or export the vars directly

# Run locally
make run
```

Required environment variables:

```bash
NAVIDROME_URL=http://localhost:4533
NAVIDROME_USER=admin
NAVIDROME_PASS=yourpassword
```

## Development Workflow

```bash
make build          # compile binary
make run            # build and run
make test           # go test -v -race ./...
make fmt            # gofmt -w
make vet            # go vet ./...
```

The binary embeds all templates and static assets at compile time (`//go:embed` in `web/web.go`). A `make build` is required to pick up template changes when running the compiled binary; `make run` does this automatically.

## Code Structure

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a full walkthrough. Short version:

```
cmd/server/        router wiring, entry point
internal/handlers/ HTTP handlers (one file per feature group)
pkg/navidrome/     Navidrome REST client (reusable, no HTTP knowledge)
internal/m3u/      M3U parser and writer (pure I/O, no HTTP)
web/templates/     HTML templates (Go html/template)
web/static/        CSS and JS
```

## Key Rules

These are enforced by CI and will block a merge if violated:

- **`make fmt`** — all Go code must be `gofmt`-formatted.
- **`make vet`** — must pass with no warnings.
- **`make test`** — all tests must pass.
- **No raw user data in responses** — all user-supplied strings must reach the browser through `html/template`. Never use `fmt.Fprintf(w, ..., userInput)`.
- **`pkg/navidrome/` is a library** — it must not import `internal/`. It must not read `http.Request`.

## Adding a Feature

A typical change touches these layers in order:

1. **`pkg/navidrome/`** — add or extend a client method if new Navidrome API calls are needed.
2. **`internal/handlers/`** — add handler function(s); wire form parsing, call the client, render template.
3. **`cmd/server/main.go`** — register the new route(s) and, if adding a full-page template, add it to `buildTemplates()`.
4. **`web/templates/`** — add or update the HTML template.
5. **Tests** — add tests in `_test.go` files in the same package. Use `httptest.NewServer` as a mock for Navidrome API tests.

See the "A Typical Change" section in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for concrete examples.

## Pull Requests

- One logical change per PR.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/): `type(scope): description` — lowercase imperative, max 50 chars, no trailing period.
- CI runs `fmt`, `vet`, `test`, and a Docker build on every PR. All checks must pass.

## License

By contributing, you agree that your changes will be licensed under [AGPL-3.0](LICENSE).
