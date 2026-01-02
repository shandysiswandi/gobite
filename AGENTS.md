# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the entrypoint; bootstrapping lives in `internal/app`.
- Core domains live under `internal/identity` (auth/profile) and `internal/notification` (email/alerts), with shared helpers in `internal/pkg`.
- Database artifacts are in `database/migrations`, `database/queries` (sqlc inputs), and `database/seeds`.
- API specs are in `api` (`swagger.yaml`/`swagger.json`); docs live in `docs`.
- Tests are Go files in `tests` and unit tests are under `internal/...`.

## Build, Test, and Development Commands
- `make run` runs the API with hot reload (uses `reflex` and `LOCAL=true`).
- `make test` and `make test-race` run unit tests in `./internal/...`.
- `go test ./tests/...` runs API/integration-style tests under `tests`.
- `make compose-up` starts local dependencies (Postgres, Redis, MinIO, etc).
- `make migrate-up` / `make seed-up` apply DB migrations/seeds (requires `POSTGRES_*` env vars).
- `make gen-sql` regenerates sqlc models; `make gen-api` regenerates Swagger via `swag`.

## Coding Style & Naming Conventions
- Go formatting: run `gofmt` (tabs, standard Go formatting).
- Package names are lowercase; file names use snake_case (`password_reset_test.go`).
- SQL migrations and seeds use numeric prefixes (`00001_init.sql`).
- Linting is done with `golangci-lint` via `make lint`.

## Testing Guidelines
- Unit tests live with packages (`internal/...`); API flow tests live in `tests`.
- Test files use `*_test.go`, and helpers live in `tests/helpers_test.go`.
- Prefer running `make test` before changes; run `go test ./tests/...` when touching API flows.

## Commit & Pull Request Guidelines
- Git history only shows an “initial commit”, so no formal convention is established.
- Use concise, imperative subjects (e.g., “add refresh token rotation”).
- PRs should include a short summary, rationale, and test evidence.
- If API behavior changes, update `api/swagger.yaml` and regenerate via `make gen-api`.

## Configuration & Local Dependencies
- Copy `config/config.example.yaml` to `config/config.yaml` and set `LOCAL=true` to use it.
- Use `CONFIG_PATH=/path/to/config.yaml` for non-local configs.
- Local services are expected via `docker compose` (see `compose.yaml`).
