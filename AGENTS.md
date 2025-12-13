# Repository Guidelines

## Project Structure & Module Organization
- Entry point: `main.go` boots the `internal/app` setup and starts the HTTP server.
- Core wiring: `internal/app` handles configuration, dependency injection, server start/stop, and module registration.
- Auth domain: `internal/auth` contains inbound HTTP handlers (`inbound`), use cases (`usecase` + `domain`), and persistence/cache adapters (`outbound`).
- Shared utilities: `internal/pkg` holds cross-cutting helpers (config, JWT, logging, router middlewares, validation, OTP, UID, etc.).
- Data layer: SQL migrations live in `database/migrations`; sqlc inputs in `database/queries`; generated Go models/queries in `internal/pkg/pkgsql`.
- Configuration: sample at `config/config.example.yaml`; runtime config resolved from `/config/config.yaml` in containers or `./config/config.yaml` when `LOCAL=true`.

## Build, Test, and Development Commands
- `make run` — local dev server with hot reload via `reflex`; uses `LOCAL=true go run main.go`.
- `make lint` — lint with `golangci-lint`.

## Coding Style & Naming Conventions
- Go 1.25+ idioms; keep code `gofmt`-clean and `golangci-lint`-clean before pushing.
- Package names are lowercase, short, and domain-driven (`auth`, `pkgjwt`, `pkgrouter`); files use snake_case when helpful.
- Favor `context.Context` threading and `slog` for logging; keep handlers thin and business logic in use cases.
- Validate inputs with the shared validator and reuse helpers from `internal/pkg` instead of duplicating logic.

## Security & Configuration Tips
- Do not commit secrets; rely on env vars and `config/config.yaml` kept out of VCS.
- JWT secrets, DB/Redis credentials, and SMTP settings must be set per environment; ensure `LOCAL=true` is used only for local dev paths.
- Run migrations before starting the app; verify Redis availability since startup pings it.***
