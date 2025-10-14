#!/bin/bash

# Docker Setup Script for CD Tool Server
# This script builds and runs the CD server as a Docker container

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="cd-server"
IMAGE_NAME="cd-tool"
IMAGE_TAG="latest"
SERVER_PORT="8080"
CONFIG_FILE="config.toml"

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}  CD Tool Docker Setup Script${NC}"
    echo -e "${CYAN}================================${NC}"
    echo
}

# Check if Docker is installed and running
check_docker() {
    print_status "Checking Docker installation..."

    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first:"
        echo "  macOS: brew install --cask docker"
        echo "  Linux: sudo apt-get install docker.io (Ubuntu/Debian)"
        echo "  Windows: Download from https://docker.com"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running. Please start Docker first."
        exit 1
    fi

    print_success "Docker is installed and running"
}

# Create Dockerfile for the CD server
create_dockerfile() {
    print_status "Creating Dockerfile..."

    cat > Dockerfile << 'EOF'
# Simple Dockerfile that uses pre-built binary
FROM ubuntu:22.04

# Install required packages and clean up
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    netcat \
    curl && \
    rm -rf /var/lib/apt/lists/* && \
    update-ca-certificates

# Create non-root user (but we'll run as root for Docker socket access)
RUN groupadd -r cduser && \
    useradd -r -g cduser cduser

# Set working directory
WORKDIR /app

# Copy pre-built binary (must be built before running docker build)
COPY bin/cd-server ./cd-server

# Copy config file (if exists)
COPY config.toml* ./

# Create directories for temporary files
RUN mkdir -p /tmp/cd-deployments && \
    chown -R cduser:cduser /app /tmp/cd-deployments

# Run as root for Docker socket access (security note: required for Docker-in-Docker)
USER root

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 8080 || exit 1

# Run the server
CMD ["./cd-server"]
EOF

    print_success "Dockerfile created"
}

# Create .dockerignore file
create_dockerignore() {
    print_status "Creating .dockerignore..."

    cat > .dockerignore << 'EOF'
# Build artifacts (but allow bin/cd-server for Docker build)
bin/*
!bin/cd-server
*.exe
*.dll
*.so
*.dylib

# Test files
*_test.go
test/

# Documentation
*.md
docs/

# Git
.git/
.gitignore

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Environment files
.env*
!.env.example

# Docker files
Dockerfile*
docker-compose*
*.dockerfile

# CI/CD
.github/
.gitlab-ci.yml

# Examples and templates
examples/
workflow-templates/

# Logs
*.log
logs/
EOF

    print_success ".dockerignore created"
}

# Create default config file if it doesn't exist
create_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        print_status "Creating default configuration file..."

        cat > "$CONFIG_FILE" << 'EOF'
# CD Tool Server Configuration

# Server listen address (format: :port or host:port)
listen_addr = ':8080'

# API keys for authentication (add your secure keys here)
api_keys = [
    'default-key-change-this',
    'another-secure-key'
]

# Optional: Log level (debug, info, warn, error)
# log_level = 'info'

# Optional: Maximum concurrent deployments
# max_deployments = 5
EOF

        print_success "Default configuration created at $CONFIG_FILE"
        print_warning "Please edit $CONFIG_FILE and change the default API keys!"
    else
        print_status "Using existing configuration file: $CONFIG_FILE"
    fi
}

# Stop and remove existing container
cleanup_existing() {
    print_status "Cleaning up existing containers..."

    if docker ps -q -f name="$CONTAINER_NAME" | grep -q .; then
        print_status "Stopping existing container: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" > /dev/null
    fi

    if docker ps -aq -f name="$CONTAINER_NAME" | grep -q .; then
        print_status "Removing existing container: $CONTAINER_NAME"
        docker rm "$CONTAINER_NAME" > /dev/null
    fi

    print_success "Cleanup completed"
}

# Build binaries first
build_binaries() {
    print_status "Building CD server binary for Linux..."

    # Create bin directory if it doesn't exist
    mkdir -p bin

    # Build the server binary for Linux (cross-compile)
    print_status "Cross-compiling for Linux AMD64..."
    if ! GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/cd-server ./cmd/server; then
        print_error "Failed to build CD server binary for Linux"
        exit 1
    fi

    print_success "CD server binary built successfully for Linux"
}

# Build Docker image
build_image() {
    print_status "Building Docker image: $IMAGE_NAME:$IMAGE_TAG"

    docker build -t "$IMAGE_NAME:$IMAGE_TAG" . || {
        print_error "Failed to build Docker image"
        exit 1
    }

    print_success "Docker image built successfully"
}

# Run container
run_container() {
    print_status "Starting CD server container..."

    docker run -d \
        --name "$CONTAINER_NAME" \
        --restart unless-stopped \
        -p "$SERVER_PORT:8080" \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v "$(pwd)/$CONFIG_FILE:/app/config.toml:ro" \
        "$IMAGE_NAME:$IMAGE_TAG" || {
        print_error "Failed to start container"
        exit 1
    }

    print_success "Container started successfully"
}

# Wait for server to be ready
wait_for_server() {
    print_status "Waiting for server to be ready..."

    for i in {1..30}; do
        if docker exec "$CONTAINER_NAME" nc -z localhost 8080 2>/dev/null; then
            print_success "Server is ready!"
            return 0
        fi
        sleep 1
        echo -n "."
    done

    echo
    print_warning "Server may still be starting up. Check logs with: docker logs $CONTAINER_NAME"
}

# Display status and usage information
show_status() {
    echo
    print_header

    print_success "CD Tool Server is now running!"
    echo

    echo -e "${YELLOW}Container Information:${NC}"
    echo "  Name: $CONTAINER_NAME"
    echo "  Image: $IMAGE_NAME:$IMAGE_TAG"
    echo "  Port: http://localhost:$SERVER_PORT"
    echo

    echo -e "${YELLOW}Useful Commands:${NC}"
    echo "  View logs:           docker logs $CONTAINER_NAME"
    echo "  View live logs:      docker logs -f $CONTAINER_NAME"
    echo "  Stop server:         docker stop $CONTAINER_NAME"
    echo "  Start server:        docker start $CONTAINER_NAME"
    echo "  Restart server:      docker restart $CONTAINER_NAME"
    echo "  Remove container:    docker rm -f $CONTAINER_NAME"
    echo "  Access container:    docker exec -it $CONTAINER_NAME sh"
    echo

    echo -e "${YELLOW}Test Deployment:${NC}"
    echo "  ./bin/cd-cli -server=localhost:$SERVER_PORT -key=default-key-change-this -git=https://github.com/user/repo.git"
    echo

    echo -e "${YELLOW}Configuration:${NC}"
    echo "  Edit config: $CONFIG_FILE"
    echo "  Restart after config changes: docker restart $CONTAINER_NAME"
    echo

    # Show current status
    echo -e "${YELLOW}Current Status:${NC}"
    docker ps --filter name="$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
}

# Show help
show_help() {
    echo "CD Tool Docker Setup Script"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -p, --port PORT     Set server port (default: 8080)"
    echo "  -n, --name NAME     Set container name (default: cd-server)"
    echo "  -t, --tag TAG       Set image tag (default: latest)"
    echo "  --rebuild           Force rebuild of Docker image"
    echo "  --clean             Clean up and rebuild everything"
    echo "  --logs              Show container logs after setup"
    echo "  --stop              Stop the running container"
    echo "  --remove            Remove container and image"
    echo
    echo "Examples:"
    echo "  $0                  # Standard setup"
    echo "  $0 --port 9000      # Use port 9000"
    echo "  $0 --rebuild        # Force rebuild"
    echo "  $0 --clean          # Clean rebuild"
    echo "  $0 --logs           # Setup and show logs"
    echo "  $0 --stop           # Stop container"
    echo "  $0 --remove         # Remove everything"
}

# Parse command line arguments
REBUILD=false
CLEAN=false
SHOW_LOGS=false
STOP_ONLY=false
REMOVE_ALL=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -p|--port)
            SERVER_PORT="$2"
            shift 2
            ;;
        -n|--name)
            CONTAINER_NAME="$2"
            shift 2
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        --rebuild)
            REBUILD=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --logs)
            SHOW_LOGS=true
            shift
            ;;
        --stop)
            STOP_ONLY=true
            shift
            ;;
        --remove)
            REMOVE_ALL=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_header

    # Handle special modes
    if [ "$STOP_ONLY" = true ]; then
        print_status "Stopping CD server container..."
        docker stop "$CONTAINER_NAME" 2>/dev/null || print_warning "Container not running"
        print_success "Container stopped"
        exit 0
    fi

    if [ "$REMOVE_ALL" = true ]; then
        print_status "Removing CD server container and image..."
        docker rm -f "$CONTAINER_NAME" 2>/dev/null || true
        docker rmi "$IMAGE_NAME:$IMAGE_TAG" 2>/dev/null || true
        print_success "Cleanup completed"
        exit 0
    fi

    # Check prerequisites
    check_docker

    # Clean rebuild if requested
    if [ "$CLEAN" = true ]; then
        print_status "Performing clean rebuild..."
        docker rm -f "$CONTAINER_NAME" 2>/dev/null || true
        docker rmi "$IMAGE_NAME:$IMAGE_TAG" 2>/dev/null || true
        REBUILD=true
    fi

    # Setup files
    create_dockerfile
    create_dockerignore
    create_config

    # Container management
    cleanup_existing

    # Build binaries first
    build_binaries

    # Build image if needed
    if [ "$REBUILD" = true ] || ! docker images "$IMAGE_NAME:$IMAGE_TAG" | grep -q "$IMAGE_TAG"; then
        build_image
    else
        print_status "Using existing Docker image: $IMAGE_NAME:$IMAGE_TAG"
    fi

    # Run container
    run_container
    wait_for_server

    # Show status
    show_status

    # Show logs if requested
    if [ "$SHOW_LOGS" = true ]; then
        echo
        print_status "Showing container logs (Ctrl+C to exit):"
        docker logs -f "$CONTAINER_NAME"
    fi
}

# Run main function
main "$@"
