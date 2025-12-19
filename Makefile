# Include the .env file if it exists, but don't fail if it's missing.
# This allows DATABASE_URL to be set either in .env or directly in the shell.
-include .env
export

.PHONY: run
run:
	@reflex -r '\.go$$' -s -R 'web|database|config' -- sh -c "LOCAL=true go run main.go"

.PHONY: test
test:
	@go test ./...

.PHONY: test-race
test-race:
	@go test -race ./...

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: migrate-up
migrate-up:
	@goose -dir database/migrations postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" up


.PHONY: migrate-down
migrate-down:
	@goose -dir database/migrations postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" down

.PHONY: seed-up
seed-up:
	@goose -dir database/seeds -table "goose_seed_db_version" postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" up


.PHONY: seed-down
seed-down:
	@goose -dir database/seeds -table "goose_seed_db_version" postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" down

.PHONY: compose-up
compose-up:
	@docker compose up -d

.PHONY: compose-down
compose-down:
	@docker compose down

.PHONY: gen-sql
gen-sql:
	@sqlc generate

.PHONY: gen-api
gen-api:
	@swag init -o api
