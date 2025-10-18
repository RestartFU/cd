.PHONY: build

build:
	@echo "Building project..."
	go build -o bin/cd-server cmd/server/main.go
	go build -o bin/cd-cli cmd/cli/main.go
	go build -o bin/cd-tui cmd/tui/main.go
