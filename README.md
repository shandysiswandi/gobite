# Gobite

Gobite is a Go backend that provides authentication and user profile APIs with MFA support.
It ships with modular identity, iam and notification domains, OpenTelemetry instrumentation, and pluggable storage and messaging backends.

## Features
- [x] User registration, email verification, login, and refresh tokens
- [x] JWT bearer authentication for protected endpoints
- [x] MFA with TOTP (plus backup code rotation)
- [x] Password reset and change flows
- [x] User profile and avatar management
- [x] Notification module with email delivery
- [x] RESTful JSON API with Swagger/OpenAPI specs
- [x] Observability via OpenTelemetry
- [x] Pluggable messaging (NSQ/Kafka/NATS/PubSub) and storage (S3/GCS/MinIO)

## Tech Stack
- Go 1.25+ (module: `github.com/shandysiswandi/gobite`)
- PostgreSQL (pgx) + SQL migrations (goose) + sqlc
- Redis (session/cache)
- JWT (HS512)
- OpenTelemetry (tracing/metrics/logging)
- Messaging: NSQ, Kafka, NATS, or Google Pub/Sub
- Storage: S3-compatible, GCS, or MinIO
- SMTP mailer
- HTTP router: `httprouter`

## Folder Structure
- `main.go` - application entrypoint
- `internal/app` - configuration, dependency wiring, server lifecycle
- `internal/iam` - notification domain
- `internal/identity` - auth domain
- `internal/notification` - notification domain
- `internal/pkg` - shared utilities (config, router, jwt, mfa, logging, etc.)
- `database/migrations` - SQL migrations (goose)
- `database/queries` - sqlc input queries
- `internal/pkg/sqlc` - generated sqlc models/queries
- `api` - Swagger/OpenAPI artifacts
- `tests` - integration tests
- `web` - frontend (separate Vite app)

## Architecture and Lifecycle
- `main.go` builds the app via `internal/app.New()` and blocks on the shutdown signal channel.
- `internal/app` owns configuration, dependency wiring, and graceful shutdown (HTTP server, goroutines, messaging, storage, and telemetry).
- Modules are initialized in `internal/app/module.go` and are gated by `modules.*.enabled` config flags.
- Casbin policies are stored in Postgres and synchronized via the pgx watcher.

## Requirements
- Go 1.25+
- PostgreSQL 17+
- Redis 7+
- SMTP server (Mailpit is used in compose)
- Messaging broker (NSQ by default)
- Object storage (MinIO for local)
- Docker (recommended for local dependencies)
- Optional tooling: `goose`, `sqlc`, `swag`, `reflex`, `golangci-lint`

## Setup (Local)
1) Copy the sample config:
```bash
cp config/config.example.yaml config/config.yaml
```
2) Update `config/config.yaml` with your local settings (DB, Redis, JWT/MFA secrets, storage, messaging).
3) Start dependencies:
```bash
docker compose up --wait
```
4) Run migrations and seeds (requires DB env vars):
```bash
export DB_USER=user
export DB_PASSWORD=password
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=gobite

make migrate-up
make seed-up
```

## Configuration
The app loads configuration from a YAML file using Viper.
- Default path: `/config/config.yaml`
- Local override: set `LOCAL=true` to use `./config/config.yaml`
- Explicit override: set `CONFIG_PATH=/path/to/config.yaml`

See `config/config.example.yaml` for all keys. Key areas include:
- `server.http.address` - HTTP bind address
- `database.url` - Postgres DSN
- `database.pool.*` - pgx pool tuning for max/min conns and lifetimes
- `redis.url` - Redis DSN
- `jwt.*` and `mfa.*` - auth secrets and TTLs
- `messaging.*` - broker configuration
- `storage.*` - S3/GCS/MinIO configuration
- `otel.*` - OpenTelemetry exporter settings

## Environment Variables
- `CONFIG_PATH` - path to the YAML config file (overrides defaults)
- `LOCAL` - if `true`, uses `./config/config.yaml`
- `DB_USER` - used by `make migrate-*` and `make seed-*`
- `DB_PASSWORD` - used by `make migrate-*` and `make seed-*`
- `DB_HOST` - used by `make migrate-*` and `make seed-*`
- `DB_PORT` - used by `make migrate-*` and `make seed-*`
- `DB_NAME` - used by `make migrate-*` and `make seed-*`

## Running the App
Dev (hot reload):
```bash
make run
```
Direct run:
```bash
LOCAL=true go run main.go
```
Production-style run (example):
```bash
CONFIG_PATH=/config/config.yaml ./gobite
```

## Deployment Notes
1) Build the binary:
```bash
go build -o gobite ./...
```
2) Provide a config file at `/config/config.yaml` or set `CONFIG_PATH`.
3) Run migrations and seeds against the target database before starting the service.
4) Ensure supporting services (Redis, messaging broker, object storage, SMTP, OTEL collector) are reachable in the deployment environment.

## Running Tests
Unit tests:
```bash
make test
```
Race tests:
```bash
make test-race
```
Integration tests (uses Docker via testcontainers):
```bash
make test-integration
```

## API Docs
- Swagger spec: `api/swagger.yaml` or `api/swagger.json`
- Regenerate: `make gen-api`

Health check:
```bash
curl http://localhost:8080/health
```
Register user:
```bash
curl -X POST http://localhost:8080/api/v1/identity/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","full_name":"User Example","password":"P@ssw0rd!"}'
```
Login:
```bash
curl -X POST http://localhost:8080/api/v1/identity/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"P@ssw0rd!"}'
```
Get profile (requires bearer token):
```bash
curl http://localhost:8080/api/v1/identity/profile \
  -H 'Authorization: Bearer <access_token>'
```

## Docker Usage
This repo ships a `compose.yaml` with local dependencies (Postgres, Redis, Mailpit, MinIO, NSQ, NATS, OTEL stack, Grafana):
```bash
docker compose up --wait
```
There is no backend container in the compose file; run the Go service locally.

## Makefile Commands
- `make run` - run API with hot reload (reflex)
- `make test` / `make test-race` / `make test-integration`
- `make lint` - run golangci-lint
- `make migrate-up` / `make migrate-down`
- `make seed-up` / `make seed-down`
- `make compose-up` / `make compose-down`
- `make gen-sql` - generate sqlc artifacts
- `make gen-api` - regenerate Swagger

## Troubleshooting
- `failed to init config`: check `CONFIG_PATH` and ensure `config/config.yaml` exists.
- `failed to init redis`: confirm `redis.url` is reachable.
- `failed to init mfacrypto`: `mfa.secret` must be base64-encoded 32 bytes.
- `failed to init jwt token`: `jwt.secret` must be base64-encoded 64 bytes.
- `authentication required`: missing `Authorization: Bearer <token>` header for protected endpoints.

## Contributing
1) Create a feature branch.
2) Run `gofmt`, `make lint`, and tests relevant to your change.
3) Open a PR with a concise description and test evidence.
