# AGENTS.md

This file provides guidance to AI agents when working with code in this repository.

## Commands

- Build: `make build`
- Format: `make fmt`
- Lint: `make lint`
- Tidy modules: `make tidy`
- Install locally: `make install`
- Run tests: `make test`
- Run the CLI from source: `go run ./cmd/nrb -- help`
- Run build/watch/serve from source: `go run ./cmd/nrb -- build|watch|serve`
- Run one test: `go test ./path/to/package -run TestName -v`

## Architecture

- `cmd/nrb` is the CLI entrypoint. `main.go` parses flags and dispatches to `build`, `watch`, or `serve`.
- `cmd/nrb/utils.go` is the main wiring layer: it loads `.env` files, reads `package.json` `nrb` config, merges CLI overrides, sets up esbuild options, and detects HTTPS certs from `.cert/` or `DEV_SERVER_CERT` / `DEV_SERVER_KEY`.
- `cmd/nrb/build.go` handles production builds: it clears the output dir, copies static files, runs esbuild, writes `version.json`, optionally writes `build-meta.json`, and injects JS/CSS/modulepreload tags into `index.html`.
- `cmd/nrb/watch.go` runs the dev workflow: esbuild serve, `fsnotify` file watching, and an SSE broker for browser reloads. It reloads on source changes and also restarts esbuild when config files change.
- `cmd/nrb/serve.go` serves the already-built `outputDir` with optional TLS.
- `lib/` contains the shared helpers used by the CLI: config parsing, flag helpers, filesystem and HTTP wrappers, logging, version handling, and index.html injection.
- `lib/config.go` defines the `package.json` `nrb` schema. Supported keys are `alias`, `resolve`, `preload`, `inject`, `loaders`, `inline`, and `splitting`.
- `lib/plugins/` holds the esbuild plugins. `alias.go` rewrites package resolution, and `inline.go` inlines matching assets as data URLs below a size threshold.
- The repo is a small esbuild-based replacement for CRA-style frontend builds. Watch mode is reload-based, not HMR-based.

## Working Notes

- Keep changes aligned with the existing `package.json`-driven configuration model unless the task explicitly changes that contract.
