APP      := coffee
BIN_DIR  := bin
DB_URL   := $(DB_URL)

include .env
export

.PHONY: run build test tidy db-up db-down db-logs migrate-up migrate-down seed all


## Build the Go binary
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) ./cmd/coffee

## Run the app directly
run:
	go run ./cmd/coffee serve

## Run tests with race detection
test:
	go test ./... -race -count=1

## Clean dependencies
tidy:
	go mod tidy

## Start/stop database with Docker Compose
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

db-logs:
	docker-compose logs -f postgres

wait-for-db:
	@echo "Waiting for Postgres to be ready..."
	@until docker exec coffee_postgres pg_isready -U postgres -d coffee > /dev/null 2>&1; do \
		echo "Postgres is unavailable - sleeping"; \
		sleep 2; \
	done
	@echo "Postgres is up!"

## Run database migrations
migrate-up: wait-for-db

	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down

## Seed the database (inside container)
seed:
	docker exec -i coffee_postgres psql -U postgres -d coffee -f /docker-entrypoint-initdb.d/seed.sql

## Full workflow: start DB, migrate, seed, run app
all: db-up migrate-up seed run
