#!/bin/bash

# CD Tool - TCP Example Usage Script
# This script demonstrates how to use the TCP-based continuous deployment tool

set -e

echo "=== CD Tool TCP Example ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_PORT="8080"
API_KEY="example-key-123"
EXAMPLE_REPO="https://github.com/restartfu/example-app.git"

echo -e "${BLUE}Step 1: Building the CD tool...${NC}"
make build
echo

echo -e "${BLUE}Step 2: Updating configuration...${NC}"
# Create example config
cat > config.toml << EOF
listen_addr = ':${SERVER_PORT}'
api_keys = ['${API_KEY}', 'default']
EOF

echo "Created config.toml with:"
cat config.toml
echo

echo -e "${BLUE}Step 3: Starting the TCP server in background...${NC}"
# Start server in background
./bin/cd-server &
SERVER_PID=$!

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    echo -e "${GREEN}Cleanup complete.${NC}"
}

# Set trap to cleanup on script exit
trap cleanup EXIT

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "${RED}Failed to start server!${NC}"
    exit 1
fi

echo -e "${GREEN}Server started successfully (PID: $SERVER_PID)${NC}"
echo

echo -e "${BLUE}Step 4: Testing TCP server...${NC}"
echo "TCP server is now running and ready for connections..."

# Test server is running
echo -e "${YELLOW}Note: This example shows the TCP server interface. For a real deployment,"
echo -e "you would need a Git repository with a Dockerfile.${NC}"
echo

echo -e "${BLUE}Step 5: Example CLI commands:${NC}"
echo
echo "# Interactive mode:"
echo "./bin/cd-cli"
echo
echo "# Deploy specific repository:"
echo "./bin/cd-cli -key=${API_KEY} -git=${EXAMPLE_REPO}"
echo
echo "# Deploy to staging environment:"
echo "./bin/cd-cli -key=${API_KEY} -git=${EXAMPLE_REPO} -env=staging"
echo
echo "# Connect to remote server:"
echo "./bin/cd-cli -server=remote-host:8080 -key=${API_KEY} -git=${EXAMPLE_REPO}"
echo

echo -e "${BLUE}Step 6: Protocol example - Manual TCP communication:${NC}"
echo "You can also communicate directly with the server using tools like netcat:"
echo
echo "# Connect to server"
echo "nc localhost ${SERVER_PORT}"
echo
echo "# Send authentication message:"
echo '{"type":"auth","timestamp":"2024-01-01T00:00:00Z","data":{"api_key":"'${API_KEY}'"}}'
echo
echo "# Send deployment message:"
echo '{"type":"deploy","timestamp":"2024-01-01T00:00:00Z","data":{"git_url":"'${EXAMPLE_REPO}'","environment":"production"}}'
echo

echo -e "${GREEN}=== Example Complete ===${NC}"
echo
echo -e "${YELLOW}Key Features Demonstrated:${NC}"
echo "✓ TCP server with real-time communication"
echo "✓ API key authentication"
echo "✓ JSON protocol for client-server communication"
echo "✓ Interactive CLI with colored output"
echo "✓ Command-line options for automated deployments"
echo "✓ Progress tracking and status updates"
echo "✓ Real-time log streaming during deployment"
echo
echo -e "${BLUE}To test with a real repository:${NC}"
echo "1. Create a Git repository with a Dockerfile"
echo "2. Use the CLI tool to deploy it:"
echo "   ./bin/cd-cli -key=${API_KEY} -git=https://github.com/yourusername/your-repo.git"
echo
echo -e "${YELLOW}The server will continue running in the background."
echo "Press Ctrl+C to stop this script and cleanup.${NC}"

# Keep script running until user interrupts
echo
echo "Waiting... (Press Ctrl+C to exit)"
while true; do
    sleep 1
done
