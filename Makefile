APP=coffee
BIN_DIR=bin

.PHONY: run build test tidy

# Run the application using the `go run` command.
run:
	go run ./cmd/coffee serve

# Build the application binary and place it in the bin directory.
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) ./cmd/coffee

# Run all tests with race detection enabled.
test:
	go test ./... -race -count=1

# Clean up unused dependencies in go.mod and go.sum.
tidy:
	go mod tidy
