# Repository Guidelines

## Project Structure & Module Organization
- `main.go` bootstraps the service; `internal/app` owns config, wiring, and lifecycle.
- Domain modules live in `internal/identity`, `internal/iam`, and `internal/notification`.
- Shared helpers and infrastructure live in `internal/pkg` (router, config, JWT/MFA, logging).
- Database artifacts live in `database/migrations`, `database/seeds`, and `database/queries` (sqlc input) with generated code in `internal/pkg/sqlc`.
- API specs are under `api/`; integration tests are under `tests/`.
- `web/` is a separate Vite frontend.

## Build, Test, and Development Commands
- `make run` runs the API with hot reload (`LOCAL=true`) using `reflex`.
- `make test` runs unit tests in `./internal/...`.
- `make test-race` runs unit tests with the race detector.
- `make test-integration` runs integration tests in `./tests/...` (uses testcontainers).
- `make lint` runs `golangci-lint`.
- `make migrate-up` / `make seed-up` apply database migrations and seeds (requires `POSTGRES_*` env vars).
- `make gen-sql` / `make gen-api` regenerate sqlc models and Swagger.

## Coding Style & Naming Conventions
- Go code follows standard formatting (`gofmt`), which uses tabs for indentation.
- Keep packages and files aligned with domain boundaries (e.g., `internal/identity/*`).
- Tests use `_test.go` suffix; prefer descriptive names like `auth_login_test.go`.
- Lint with `golangci-lint` before opening a PR.

## Testing Guidelines
- Unit tests live alongside code in `internal/...`.
- Integration tests live in `tests/` and assume Docker is available.
- Run `make test` for fast checks, and `make test-integration` before larger changes.

## Commit & Pull Request Guidelines
- Git history does not yet define a strict format; use short, imperative summaries (e.g., “Add refresh token rotation”).
- PRs should include a concise description, relevant test output, and any config or migration notes.
- Link related issues when applicable and call out any breaking changes.

## Configuration & Local Dependencies
- Copy `config/config.example.yaml` to `config/config.yaml`; set `LOCAL=true` to load it.
- Update `CONFIG_PATH` when running with a non-default config location.
- Local dependencies are provided via `compose.yaml`; start them with `docker compose up --wait`.
