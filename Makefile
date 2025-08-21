APP      := coffee
BIN_DIR  := bin
DB_URL   := $(DB_URL)
API_BASE := http://localhost:9090/v1

# ---- Prometheus quick helpers ----
## Prometheus URL (default to local dev instance)
PROM_URL ?= http://localhost:9091
# query window (override with WIN=1m,10m)
WIN ?= 5m


include .env
export

.PHONY: run build test tidy db-up db-down db-logs migrate-up migrate-down seed all \
        health menu orders-list orders-clear orders-create

##Main Makefile targets for Golang Coffee App
## Build the Go binary
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) ./cmd/coffee

## Run the app directly
run:
	go run ./cmd/coffee serve

## Run tests with race detection
test:
	go test ./... -race -count=1 -v

## Run integration tests
test-integration:
	go test -tags=integration ./test/integration/... -count=1 -v

## Run integration tests with verbose output + benchmarks
## Runs integration tests *and* Go benchmarks (`go test -bench . -benchmem`).
## Example output line:
##   BenchmarkMenuEndpoint-8   1704   611095 ns/op   26384 B/op   320 allocs/op
##
## Explanation:
##   • BenchmarkMenuEndpoint-8  → benchmark name; "-8" = ran with 8 logical CPUs (GOMAXPROCS=8).
##   • 1704                     → number of iterations Go ran this benchmark for statistical confidence.
##   • 611095 ns/op             → average time per operation (≈ 0.61 ms/request).
##   • 26384 B/op               → average heap memory allocated per operation (≈ 25.8 KB).
##   • 320 allocs/op            → average number of heap allocations per operation.
##
## Useful for: spotting regressions in latency (ns/op), memory footprint (B/op),
## and allocation count (allocs/op) across code changes.
test-integration-bench:
	go test -tags=integration ./test/integration/... -bench . -benchmem -count=1

## Clean dependencies
tidy:
	go mod tidy



## Start the API server with Docker Compose
api-up:
	@echo "Building API container..."
	@docker-compose build api
	@echo "Starting API server..."
	@docker-compose up -d api


## Stop all containers and remove volumes
all-down:
	docker-compose down -v

## Main targets for PostgreSQL database management

## Start/stop database with Docker Compose
db-up:
	docker-compose up -d postgres

## Get the logs of the Postgres container
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

## Full workflow:  up, migrate, seed, 
complete-db-setup:  db-up migrate-up seed 

## Prometheus query helpers

## Start Prometheus with Docker Compose
prometheus-up: 
	docker-compose up -d prometheus

# 1) See if Prometheus is up and scraping your API (robust)
prom-targets:
	@echo "Prometheus targets @ $(PROM_URL)"
	@if ! curl -fsS $(PROM_URL)/-/healthy >/dev/null; then \
		echo "  ERROR: Prometheus not reachable on $(PROM_URL). Try: make prometheus-up"; \
		exit 1; \
	fi
	@curl -fsS $(PROM_URL)/api/v1/targets \
	 | jq -r '.data.activeTargets[]? | "\(.labels.job) @ \(.labels.instance): \(.health)\t\(.lastError)"'

# 2) Generate simple traffic so numbers aren’t NaN (defaults to 30 hits)
traffic:
	@count=$${COUNT:-30}; \
	echo "Hitting $(API_BASE)/menu $$count times..."; \
	for i in $$(seq 1 $$count); do curl -s $(API_BASE)/menu >/dev/null; done; \
	echo "Done."

# One-shot summary with friendly explanations: requests, RPS, p90 latency
metrics-summary:
	@echo "Metrics Summary (window=$(WIN))"
	@if ! curl -fsS "$(PROM_URL)/-/healthy" >/dev/null; then \
		echo "  ERROR: Prometheus not reachable on $(PROM_URL). Try: make prometheus-up"; \
		exit 1; \
	fi
	@echo "• requests — how many /v1/menu requests happened in the last $(WIN) (uses increase())."
	@tot=$$(curl -fsSG "$(PROM_URL)/api/v1/query" \
	  --data-urlencode 'query=sum(increase(coffee_menu_requests_total['"$(WIN)"']))' \
	  | jq -r 'if (.data.result|length)>0 then .data.result[0].value[1] else "0" end'); \
	printf "  requests: %s in %s\n" "$$tot" "$(WIN)"
	@echo "• rps — average requests/second over $(WIN) (uses rate()). Good for throughput."
	@rps=$$(curl -fsSG "$(PROM_URL)/api/v1/query" \
	  --data-urlencode 'query=sum(rate(coffee_menu_requests_total['"$(WIN)"']))' \
	  | jq -r 'if (.data.result|length)>0 then .data.result[0].value[1] else "0" end'); \
	printf "  rps: %s\n" "$$rps"
	@echo "• p90 — 90% of requests completed faster than this (tail latency via histogram_quantile)."
	@p90=$$(curl -fsSG "$(PROM_URL)/api/v1/query" \
	  --data-urlencode 'query=histogram_quantile(0.90, sum(rate(coffee_menu_latency_seconds_bucket['"$(WIN)"'])) by (le))' \
	  | jq -r 'if (.data.result|length)>0 then (.data.result[0].value[1]|try (tonumber*1000) catch "NaN") else "NaN" end'); \
	printf "  p90: %s ms\n" "$$p90"


# List ALL histogram buckets and show their current cumulative rates
metrics-buckets:
	@echo "Latency histogram buckets for coffee_menu_latency_seconds (window=$(WIN))"
	@if ! curl -fsS "$(PROM_URL)/-/healthy" >/dev/null; then \
		echo "  ERROR: Prometheus not reachable on $(PROM_URL). Try: make prometheus-up"; \
		exit 1; \
	fi
	@echo "• Buckets are cumulative by 'le' (<=). +Inf equals total request rate."
	@echo "• Default buckets (DefBuckets) are: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, +Inf (seconds)."
	@echo "• Interpret shape: big jumps between adjacent buckets mean many requests fall in that latency range."
	@echo ""
	@curl -fsSG "$(PROM_URL)/api/v1/query" \
	  --data-urlencode 'query=sum(rate(coffee_menu_latency_seconds_bucket['"$(WIN)"'])) by (le)' \
	  | jq -r '.data.result | sort_by(.metric.le|try tonumber catch 1e99)[] | "\(.metric.le)s\t\(.value[1]) req/s"'


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
	curl -s $(API_BASE)/menu | jq .

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
	@echo "  make build                    Build the Go binary ($(BIN_DIR)/$(APP))"
	@echo "  make run                      Run the app locally (uses .env DB_URL)"
	@echo "  make test                     Run all tests with race detection"
	@echo "  make test-integration         Run integration tests (uses testcontainers)"
	@echo "  make test-integration-bench   Run integration benchmarks"
	@echo "  make tidy                     Clean up go.mod/go.sum"
	@echo "  make db-up                    Start PostgreSQL container"
	@echo "  make db-logs                  Tail PostgreSQL logs"
	@echo "  make migrate-up               Run database migrations"
	@echo "  make migrate-down             Roll back the last migration"
	@echo "  make seed                     Seed the database with initial data"
	@echo "  make complete-db-setup        Start DB, run migrations, seed"
	@echo "  make api-up                   Build & start the API container (Compose)"
	@echo "  make prometheus-up            Start Prometheus (headless scrape)"
	@echo "  make all-down                 Stop all containers and remove volumes"
	@echo ""
	@echo "Typical flow:"
	@echo "  1) make db-up"
	@echo "  2) make complete-db-setup"
	@echo "  3) make prometheus-up"
	@echo "  4) make api-up"
	@echo ""


## Show available API testing commands
api-help:
	@echo ""
	@echo "API Testing Commands:"
	@echo "  make health                   Check /healthz endpoint"
	@echo "  make ready                    Check /readyz endpoint"
	@echo "  make menu                     Pretty-print the coffee menu (/menu)"
	@echo "  make traffic         		Generate sample traffic to /menu"
	@echo "  make orders-list              List all orders (/orders)"
	@echo "  make orders-clear             Clear all orders (DELETE /orders)"
	@echo "  make orders-create            Create a new order (interactive prompt)"
	@echo ""

## Show Prometheus quick demo commands
prom-help:
	@echo ""
	@echo "Prometheus Quick Demos (PROM_URL=$(PROM_URL), default WIN=$(WIN)):"
	@echo "  make prom-targets                Show scrape targets and health"
	@echo "  make metrics-summary   	   Summary: requests, RPS, p90 latency"
	@echo "  make metrics-buckets   	   Show latency buckets (cumulative rates)"
	@echo ""