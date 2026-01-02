# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the service entrypoint; runtime wiring lives in `internal/app`.
- Domain modules live under `internal/identity`, `internal/iam`, and `internal/notification`.
- Shared utilities (config, router, JWT, MFA, logging) are in `internal/pkg`.
- Database assets: migrations in `database/migrations`, seeds in `database/seeds`, sqlc inputs in `database/queries`, generated code in `internal/pkg/sqlc`.
- API artifacts are in `api/`; integration tests are in `tests/`; the separate frontend lives in `web/`.

## Build, Test, and Development Commands
- `make run` runs the API with hot reload (uses `reflex` and `LOCAL=true`).
- `make test` / `make test-race` run unit tests under `internal/`.
- `make test-integration` runs integration tests in `tests/` (Docker/testcontainers required).
- `make lint` runs `golangci-lint`.
- `make migrate-up` / `make seed-up` apply database migrations/seeds using the `DB_*` env vars.
- `make gen-sql` regenerates sqlc models; `make gen-api` regenerates Swagger docs.

## Coding Style & Naming Conventions
- Go code should be formatted with `gofmt` (tabs, standard Go layout).
- Follow Go naming: `CamelCase` for exported identifiers, `camelCase` for local vars, lowercase package names.
- Test files use `*_test.go`, and test names follow `TestXxx` conventions.
- Lint with `golangci-lint run` before opening a PR.

## Testing Guidelines
- Unit tests live in `internal/**`; integration tests live in `tests/`.
- Prefer adding tests adjacent to the affected module; keep setup helpers in `tests/auth_util_test.go`.
- Run the smallest relevant test command, then full suite if behavior is cross-cutting.

## Commit & Pull Request Guidelines
- The Git history does not define a commit message convention yet; use concise, imperative summaries (e.g., "Add MFA backup code rotation").
- Create feature branches and open PRs with a clear description of changes and testing evidence (commands and results).
- Link relevant issues and note any API or config changes (e.g., new `config/config.yaml` keys).

## Configuration & Local Dependencies
- Local config uses `config/config.yaml` (copy from `config/config.example.yaml`); set `LOCAL=true` to use it.
- Required services: Postgres, Redis, SMTP, messaging broker, and object storage; use `docker compose up --wait` for local deps.
