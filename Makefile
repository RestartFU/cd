.PHONY: build build-server build-cli clean install help

# Default target
all: build

# Build both server and client
build: build-server build-cli

# Build the TCP server
build-server:
	@echo "Building TCP server..."
	go build -o bin/cd-server ./cmd/server

# Build the CLI client
build-cli:
	@echo "Building CLI client..."
	go build -o bin/cd-cli ./cmd/cli

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Install binaries to GOPATH/bin
install: build
	@echo "Installing binaries..."
	go install ./cmd/server
	go install ./cmd/cli

# Run the server
run-server: build-server
	@echo "Starting TCP server..."
	./bin/cd-server

# Run tests
test:
	@echo "Running tests..."
	./scripts/test.sh

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test ./internal/config -v
	go test ./internal/protocol -run "TestNewMessage|TestMessage_ParseData|TestMessageSerialization|TestMessageTypes|TestInvalidMessageData|TestEmptyMessageData" -v
	go test ./internal/adapters/handler -run "TestNewAdapter|TestDeployResult_Fields|TestAdapter_Deploy_EmptyGitURL" -v

# Run all tests with timeout
test-all:
	@echo "Running all tests with timeout..."
	go test ./... -timeout 30s

# Run specific test package
test-config:
	@echo "Testing configuration..."
	go test ./internal/config -v

test-protocol:
	@echo "Testing protocol..."
	go test ./internal/protocol -v

test-handler:
	@echo "Testing handler..."
	go test ./internal/adapters/handler -v

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test ./... -bench=. -benchmem

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Check for issues
vet:
	@echo "Running go vet..."
	go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Create release builds for multiple platforms
release: clean
	@echo "Building release binaries..."
	mkdir -p dist

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/cd-server-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/cd-cli-linux-amd64 ./cmd/cli

	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/cd-server-linux-arm64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/cd-cli-linux-arm64 ./cmd/cli

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/cd-server-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/cd-cli-darwin-amd64 ./cmd/cli

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/cd-server-darwin-arm64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/cd-cli-darwin-arm64 ./cmd/cli

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/cd-server-windows-amd64.exe ./cmd/server
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/cd-cli-windows-amd64.exe ./cmd/cli

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build both server and client"
	@echo "  build-server - Build only the TCP server"
	@echo "  build-cli    - Build only the CLI client"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install binaries to GOPATH/bin"
	@echo "  run-server   - Build and run the server"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  release      - Build release binaries for multiple platforms"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Test targets:"
	@echo "  test         - Run comprehensive test suite"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-all     - Run all tests with timeout"
	@echo "  test-config  - Test configuration package"
	@echo "  test-protocol- Test protocol package"
	@echo "  test-handler - Test handler package"
	@echo "  bench        - Run benchmarks"
