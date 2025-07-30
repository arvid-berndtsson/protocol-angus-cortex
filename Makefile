# Protocol Argus Cortex Makefile

.PHONY: build clean test lint fmt deps run docker-build docker-run help

# Build variables
BINARY_NAME=protocol-argus-cortex
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Default target
all: build

# Build the application
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/protocol-argus-cortex/

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p ${BUILD_DIR}
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/protocol-argus-cortex/

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p ${BUILD_DIR}
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ./cmd/protocol-argus-cortex/

build-windows:
	@echo "Building for Windows..."
	@mkdir -p ${BUILD_DIR}
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ./cmd/protocol-argus-cortex/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf ${BUILD_DIR}
	rm -f coverage.out coverage.html

# Run the application (requires config.yml)
run: build
	@echo "Running ${BINARY_NAME}..."
	@if [ ! -f config.yml ]; then \
		echo "Error: config.yml not found. Please copy config.yml.example to config.yml and configure it."; \
		exit 1; \
	fi
	./${BUILD_DIR}/${BINARY_NAME} --config config.yml

# Run with verbose logging
run-verbose: build
	@echo "Running ${BINARY_NAME} with verbose logging..."
	./${BUILD_DIR}/${BINARY_NAME} --config config.yml --verbose

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t protocol-argus-cortex:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -d --name argus-cortex \
		-p 8080:8080 \
		-p 9090:9090 \
		--cap-add=NET_ADMIN \
		--cap-add=NET_RAW \
		protocol-argus-cortex:latest

# Docker stop
docker-stop:
	@echo "Stopping Docker container..."
	docker stop argus-cortex || true
	docker rm argus-cortex || true

# Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	godoc -http=:6060 &
	@echo "Documentation available at http://localhost:6060"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/godoc@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-all      - Build for all platforms"
	@echo "  deps           - Install dependencies"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  clean          - Clean build artifacts"
	@echo "  run            - Run the application"
	@echo "  run-verbose    - Run with verbose logging"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-stop    - Stop Docker container"
	@echo "  docs           - Generate documentation"
	@echo "  install-tools  - Install development tools"
	@echo "  help           - Show this help" 