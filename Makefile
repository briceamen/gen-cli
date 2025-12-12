.PHONY: all build generator runtime clean generate lint fmt

# Default target
all: build

# Build both CLIs
build: generator runtime

# Build the generator CLI
generator:
	go build -o bin/scalingo-gen-generator ./cmd/generator

# Build the runtime CLI
runtime:
	go build -o bin/scalingo-gen ./cmd/runtime

# Generate commands from manifest
generate: generator
	go generate ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f scalingo-cli

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	gofmt -w .

# Run tests
test:
	go test ./...

# Build and regenerate (full rebuild)
rebuild: clean generate runtime
