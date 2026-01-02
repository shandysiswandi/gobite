# Include the .env file if it exists, but don't fail if it's missing.
# This allows DATABASE_URL to be set either in .env or directly in the shell.
-include .env
export

.PHONY: help restart run test test-race test-integration lint migrate-up migrate-down seed-up seed-down compose-up compose-down gen-sql gen-api

## meta: Show available make targets.
help:
	@awk 'BEGIN {FS=":.*## "; \
		print "+----------------------+------------------------------------------+"; \
		print "| Target               | Description                              |"; \
		print "+----------------------+------------------------------------------+"} \
		/^[a-zA-Z0-9_-]+:.*## / { \
			target=$$1; desc=$$2; \
			if (length(target) > 20) target=substr(target, 1, 20); \
			if (length(desc) > 40) desc=substr(desc, 1, 40); \
			printf "| %-20s | %-40s |\n", target, desc \
		} \
		END {print "+----------------------+------------------------------------------+"}' $(MAKEFILE_LIST)


## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****
## Development
## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****

restart: ## Restart local stack and refresh generated assets.
	@docker compose down -v
	@docker compose up --wait
	@$(MAKE) migrate-up
	@$(MAKE) seed-up
	@$(MAKE) gen-sql
	@$(MAKE) gen-api
	@go mod tidy
	@gofmt -w .
	@mcli mb local/gobite-assets

run: ## Run the API with hot reload (LOCAL=true).
	@reflex -r '\.go$$' -s -R 'config|database|deploy|docs|tests|web' -- sh -c "LOCAL=true go run main.go"

test: ## Run unit tests for internal packages.
	@go test ./internal/...

test-race: ## Run unit tests with the race detector enabled.
	@go test -race ./internal/...

test-real: ## Run real tests under ./tests/real.
	@go test -count=1 ./tests/... -parallel 4 -v

lint: ## Lint the codebase with golangci-lint.
	@golangci-lint run

## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****
## Migration
## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****

migrate-up: ## Apply database migrations.
	@goose -dir database/migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" up


migrate-down: ## Roll back the most recent migration.
	@goose -dir database/migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" down

## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****
## Seeder
## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****

seed-up: ## Apply database seed scripts.
	@goose -dir database/seeds -table "goose_seed_db_version" postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" up


seed-down: ## Roll back the most recent seed script.
	@goose -dir database/seeds -table "goose_seed_db_version" postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" down

## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****
## Docker
## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****

compose-up: ## Start the docker compose stack.
	@docker compose up --wait

compose-down: ## Stop the docker compose stack.
	@docker compose down

## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****
## Generator
## ***** ***** ***** ***** ***** ***** ***** ***** ***** *****

gen-sql: ## Generate sqlc models and queries.
	@sqlc generate

gen-api: ## Generate OpenAPI docs via swag.
	@swag init --v3.1 -o api
