APP      := coffee
BIN_DIR  := bin
DB_URL   := $(DB_URL)
API_BASE := http://localhost:9090/v1

include .env
export

.PHONY: run build test tidy db-up db-down db-logs migrate-up migrate-down seed all \
        health menu orders-list orders-clear orders-create


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

## Run integration tests
test-integration:
	go test -tags=integration ./test/integration/... -count=1 -v

## Run integration tests with verbose output + benchmarks
test-integration-bench:
	go test -tags=integration ./test/integration/... -bench . -benchmem -count=1

## Clean dependencies
tidy:
	go mod tidy

## Start/stop database with Docker Compose
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down -v

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

### -----------------------------
### API TEST TARGETS
### -----------------------------

## Health check endpoints
health:
	curl -i $(API_BASE)/healthz
ready:
	curl -i $(API_BASE)/readyz

## Menu endpoint
menu:
	curl -i $(API_BASE)/menu

## List all orders
orders-list:
	curl -i $(API_BASE)/orders

## Clear all orders
orders-clear:
	curl -X DELETE $(API_BASE)/orders

## Create an order (interactive)
## Create an order (interactive)
orders-create:
	@echo " Place a new order:"; \
	read -p "Customer name: " customer; \
	read -p "Items (comma separated): " items; \
	payload=$$(printf '{"customer":"%s","items":[%s]}' "$$customer" "$$(echo $$items | sed 's/[^,][^,]*/"&"/g')"); \
	\
	echo ""; \
	echo "======================================"; \
	echo "        PRINTING ORDER DOCKET     "; \
	echo "======================================"; \
	echo "$$payload"; \
	echo "======================================"; \
	echo ""; \
	\
	curl -s -X POST -H "Content-Type: application/json" \
		-d "$$payload" \
		$(API_BASE)/orders | jq .



## Show available build/dev commands
main-help:
	@echo ""
	@echo "Main Build & Dev Commands:"
	@echo "  make build          Build the Go binary ($(BIN_DIR)/$(APP))"
	@echo "  make run            Run the app directly with 'go run'"
	@echo "  make test           Run all tests with race detection"
	@echo "  make test-integration Run integration tests"
	@echo "  make tidy           Clean up go.mod/go.sum"
	@echo "  make db-up          Start PostgreSQL container"
	@echo "  make db-down        Stop PostgreSQL container"
	@echo "  make migrate-up     Run database migrations"
	@echo "  make migrate-down   Roll back the last migration"
	@echo "  make seed           Seed the database with initial data"
	@echo "  make all            Full workflow: db-up, migrate, seed, run"
	@echo ""

## Show available API testing commands
api-help:
	@echo ""
	@echo "API Testing Commands:"
	@echo "  make health         Check /healthz endpoint"
	@echo "  make ready          Check /readyz endpoint"
	@echo "  make menu           Get the coffee menu (/menu)"
	@echo "  make orders-list    List all orders (/orders)"
	@echo "  make orders-clear   Clear all orders (DELETE /orders)"
	@echo "  make orders-create  Create a new order (interactive prompt)"
	@echo ""
