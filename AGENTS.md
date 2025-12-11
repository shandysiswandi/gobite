# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the entrypoint that boots the app via `internal/app`, which wires config (Viper), logging, HTTP server, and module registration.
- `internal/auth` holds the authentication module with `inbound` HTTP handlers, `usecase` orchestration, and `domain` contracts/entities. Keep business rules here, not in handlers.
- `internal/pkg` contains shared utilities (config, logging, router, validator, clock/uid/hash helpers, and sqlc-generated `pkgsql`). Add reusable code here instead of cross-module imports.
- `config/config.yaml` and `config/config.example.yaml` carry defaults; environment variables (via `.env`) override sensitive values. Database schemas live in `database/migrations`; sqlc query specs sit in `database/queries`.

## Build, Test, and Development Commands
- `make run` — hot-reload dev server with `reflex`; uses `LOCAL=true` and reads `.env` if present.
- `go run main.go` — single-run local start (useful when reflex isn’t installed).
- `make compose-up` / `make compose-down` — start/stop dockerized deps such as Postgres.
- `make migrate-up` / `make migrate-down` — apply/rollback goose migrations using `DB_*` env vars.
- `make gen-sql` — regenerate typed DB access layer from `database/queries` via sqlc.
- `go test ./...` — run unit tests across modules when present.

## Coding Style & Naming Conventions
- Go 1.25; always `gofmt` and clean imports (`goimports`). Use tabs (Go default) and keep lines concise.
- Package names stay lower_snake (e.g., `pkgsql`); exported identifiers follow Go capitalization. Respect boundaries between inbound/usecase/domain layers.
- Keep handlers thin; push logic into usecases and domain contracts. Prefer constructor functions for dependencies over globals.

## Testing Guidelines
- Favor table-driven tests and explicit fixtures; avoid shared mutable state. Short-lived `context.Context` in tests.
- Use a dedicated test database; apply migrations per suite and clean up after. Mock shared helpers (`pkgclock`, `pkghash`, `pkguid`) when asserting logic.
- Cover HTTP handlers (inbound) and usecases; prefer black-box tests that hit the module surface.

## Commit & Pull Request Guidelines
- Commits: short, imperative subjects (`add oauth callback handler`), scoped to one concern. Include migration/config updates in the same commit.
- PRs: explain intent, list breaking changes, note commands run (`go test ./...`, `make migrate-up`), and attach API samples or screenshots when user-visible.
- Link related issues; update `config/config.example.yaml` and docs when adding settings or endpoints.

## Configuration & Security
- Supply `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, and `DB_NAME` via `.env` or environment; do not commit secrets.
- Keep defaults non-sensitive; rotate keys when sharing environments. Avoid logging secrets—audit `pkglog` fields before adding.
