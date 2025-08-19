APP=coffee
BIN_DIR=bin

.PHONY: run build test tidy

# Run the application using the `go run` command.
# This target starts the server defined in the `cmd/coffee` package.
run:
	go run ./cmd/coffee serve

# Build the application binary and place it in the bin directory.
# This target compiles the Go code into an executable binary.
build:
	mkdir -p $(BIN_DIR) # Ensure the bin directory exists.
	go build -o $(BIN_DIR)/$(APP) ./cmd/coffee

# Run all tests with race detection enabled.
# This target ensures the code is tested thoroughly for concurrency issues.
test:
	go test ./... -race -count=1

# Clean up unused dependencies in go.mod and go.sum.
# This target ensures the project's dependency files are up-to-date.
tidy:
	go mod tidy
