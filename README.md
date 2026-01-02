# Gobite

Gobite is a Go backend for authentication, user profiles, and notifications. It includes MFA, JWT auth, pluggable messaging/storage backends, and OpenTelemetry instrumentation.

## Features
- User registration, email verification, login, refresh tokens
- MFA with TOTP and backup code rotation
- Password reset/change flows and profile management
- RESTful JSON API with Swagger/OpenAPI specs
- Casbin-backed authorization with Postgres storage
- Pluggable messaging (NSQ/Kafka/NATS/Pub/Sub) and storage (S3/GCS/MinIO)
- Observability via OpenTelemetry

## Tech Stack
- Go 1.25.4
- PostgreSQL 17+ (pgx), goose migrations, sqlc
- Redis 8+ (session/cache)
- JWT (HS512), SMTP mailer
- HTTP router: `httprouter`

## Project Structure
```sh
.
├── main.go           # application entrypoint
├── api               # Swagger/OpenAPI artifacts
├── config            # YAML configuration
├── database          # migrations, sqlc queries, seed scripts
├── deploy            # observability stack configs
├── docs              # documentation
├── internal          # application modules and shared packages
│   ├── app           # bootstrapping and wiring
│   ├── identity      # auth and profile domain
│   ├── notification  # notification and email domain
│   ├── media         # media module
│   ├── pkg           # shared utilities
│   └── shared        # shared contracts/events
└── tests             # API/integration-style tests
```

## Setup (Local)
1) Copy config:
```bash
cp config/config.example.yaml config/config.yaml
```
2) Start dependencies:
```bash
docker compose up --wait
```
3) Run migrations and seeds:
```bash
export POSTGRES_USER=user
export POSTGRES_PASSWORD=password
export POSTGRES_DB=gobite
make migrate-up
make seed-up
```

## Run the Service
Dev (hot reload):
```bash
make run
```
Direct run:
```bash
LOCAL=true go run main.go
```

## Configuration
- Default config path: `/config/config.yaml`
- Local override: `LOCAL=true` uses `./config/config.yaml`
- Explicit override: `CONFIG_PATH=/path/to/config.yaml`

See `config/config.example.yaml` for all keys.

## Database & Codegen
- `make migrate-up` / `make migrate-down` uses goose against Postgres on `localhost:5432`.
- `make gen-sql` regenerates sqlc models and queries.
- `make gen-api` regenerates Swagger via `swag`.

## Tests
- Unit tests: `make test` or `make test-race`
- API tests: `make test-real` `go test ./tests/...`

## API Docs & Quick Checks
- Swagger: `api/swagger.yaml` or `api/swagger.json`
- Health:
```bash
curl http://localhost:8080/health
```
- Register:
```bash
curl -X POST http://localhost:8080/api/v1/identity/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","full_name":"User Example","password":"P@ssw0rd!"}'
```

## Docker Usage
This repo ships a `compose.yaml` with local dependencies (Postgres, Redis, Mailpit, MinIO, NSQ, NATS, Tempo/Prometheus/Loki, OTEL collector, Grafana):
```bash
docker compose up --wait
```
There is no backend container in the compose file; run the Go service locally.
MinIO defaults to `MINIO_ROOT_USER=user` and `MINIO_ROOT_PASSWORD=password`; align `storage.minio.*` in `config/config.yaml` if you use MinIO locally.

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
