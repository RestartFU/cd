# CD Tool - TCP Continuous Deployment

A high-performance continuous deployment tool built on pure TCP for real-time communication between client and server. Deploy applications from Git repositories to Docker containers with live log streaming and status updates.

## Features

- **Pure TCP Communication** - Fast, persistent connections for real-time updates
- **Real-time Log Streaming** - See deployment logs as they happen
- **Interactive CLI** - User-friendly command-line interface with colored output
- **Docker Integration** - Automatic Docker image building and container deployment
- **Git Repository Support** - Clone and deploy from any Git repository
- **Progress Tracking** - Visual progress bars and status updates
- **API Key Authentication** - Secure access control
- **Environment Variables & Secrets** - Secure injection of configuration and credentials
- **GitHub Actions Integration** - Ready-to-use workflow templates
- **Multi-environment Support** - Deploy to different environments

## Quick Start

### 1. Install

```bash
# Download latest release
curl -L -o cd-cli https://github.com/RestartFU/cd/releases/latest/download/cd-cli-linux-amd64
chmod +x cd-cli

# Or build from source
git clone https://github.com/RestartFU/cd.git
cd cd
make build
```

### 2. Start Server

```bash
# Configure server
cat > config/config.toml << EOF
listen_addr = ':8080'
api_keys = ['your-secret-key']
EOF

# Start server
./bin/cd-server
```

### 3. Deploy with CLI

```bash
# Interactive mode
./cd-cli

# Direct deployment
./cd-cli -server=localhost:8080 -key=your-secret-key -git=https://github.com/user/repo.git

# Deploy with environment variables
./cd-cli -server=localhost:8080 -key=your-secret-key -git=https://github.com/user/repo.git \
         -env-vars="API_KEY=value,DEBUG=true" -env-file=".env.production"
```

## GitHub Actions Integration

### Setup Secrets

Add these secrets to your repository:

- `CD_SERVER_HOST` - Your CD server address (e.g., `cd-server.example.com:8080`)
- `CD_API_KEY` - Your API key for authentication

### Simple Deployment

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy Application

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    uses: RestartFU/cd/.github/workflows/deploy.yml@main
    with:
      environment: production
      env_vars: "BUILD_NUMBER=${{ github.run_number }}"
      include_github_secrets: true
    secrets:
      CD_SERVER_HOST: ${{ secrets.CD_SERVER_HOST }}
      CD_API_KEY: ${{ secrets.CD_API_KEY }}
```

### Advanced Deployment

Use the advanced template from `workflow-templates/deploy-advanced.yml` for:
- Multi-environment support
- Pre-deployment validation
- Integration testing
- Rollback capabilities

## Configuration

```toml
listen_addr = ':8080'
api_keys = ['your-secret-key']
```

## CLI Usage

```bash
# Options
./cd-cli -server=HOST:PORT -key=API_KEY -git=REPO_URL -env=ENVIRONMENT

# Examples
./cd-cli -key=secret -git=https://github.com/user/repo.git
./cd-cli -key=secret -git=https://github.com/user/repo.git -env=staging
./cd-cli -server=remote:8080 -key=secret -git=https://github.com/user/repo.git

# With environment variables
./cd-cli -key=secret -git=https://github.com/user/repo.git \
         -env-vars="API_KEY=value,DEBUG=false" \
         -env-file=".env.production"

# With secrets file
./cd-cli -key=secret -git=https://github.com/user/repo.git \
         -secrets-file=".secrets" \
         -env="production"
```

## Docker Deployment

### Quick Docker Setup (Recommended)

Use the automated setup script for easy Docker deployment:

```bash
# Automated setup with all dependencies
./scripts/docker-setup.sh

# Or with custom options
./scripts/docker-setup.sh --port 9000 --logs
```

### Docker Compose

```bash
# Basic setup
docker-compose up -d

# With monitoring and caching
docker-compose --profile cache --profile monitoring up -d
```

### Manual Docker Setup

```bash
# Build the image
docker build -t cd-tool:latest .

# Run the container
docker run -d \
  --name cd-server \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/config/config.toml:/app/config.toml:ro \
  cd-tool:latest
```

### Docker Features

- **Automated OS-specific Docker socket detection**
- **Health checks and automatic restart**
- **Non-root container security**
- **Volume mounting for configuration**
- **Multi-stage builds for smaller images**
- **Optional monitoring with Prometheus/Grafana**

📖 **For complete Docker deployment guide, see [docs/DOCKER_DEPLOYMENT.md](docs/DOCKER_DEPLOYMENT.md)**

### Repository Requirements

Your repository needs a `Dockerfile`:

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
```

### Docker Management Commands

```bash
# View container status
docker ps --filter name=cd-server

# View logs
docker logs -f cd-server

# Stop/start container
docker stop cd-server
docker start cd-server

# Update deployment
./scripts/docker-setup.sh --rebuild

# Complete cleanup
./scripts/docker-setup.sh --remove
```

## Environment Variables & Secrets

The CD tool supports secure injection of environment variables and secrets into your containers:

### Environment Files

Create `.env` files for your configuration:

```bash
# .env.production
NODE_ENV=production
PORT=3000
API_ENDPOINT=https://api.prod.example.com
```

### GitHub Actions with Environment Variables

```yaml
deploy:
  uses: RestartFU/cd/.github/workflows/deploy.yml@main
  with:
    environment: production
    env_file: ".env.production"
    env_vars: "BUILD_ID=${{ github.run_id }},COMMIT_SHA=${{ github.sha }}"
    include_github_secrets: true
  secrets:
    CD_SERVER_HOST: ${{ secrets.CD_SERVER_HOST }}
    CD_API_KEY: ${{ secrets.CD_API_KEY }}
```

### CLI with Environment Variables

```bash
# From environment file
./cd-cli -env-file=".env.production" -git=https://github.com/user/repo.git

# From command line
./cd-cli -env-vars="API_KEY=value,DEBUG=true" -git=https://github.com/user/repo.git

# With secrets file
./cd-cli -secrets-file=".secrets" -git=https://github.com/user/repo.git
```

### Security Features

- **Secrets are masked** in deployment logs
- **Environment-specific configurations** with different `.env` files
- **GitHub secrets integration** - automatically include repository secrets
- **Precedence handling** - secrets override environment variables

📖 **For complete documentation, see [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md)**

## Development

```bash
# Build
make build

# Test
make test

# Run tests
./scripts/test.sh

# Format & vet
make fmt
make vet

# Cross-platform build
make release
```

## Protocol

JSON messages over TCP:

```json
{"type":"auth","data":{"api_key":"secret"}}
{"type":"deploy","data":{"git_url":"https://github.com/user/repo.git","environment":"production","env_vars":{"API_KEY":"value"},"secrets":{"DB_PASSWORD":"secret"}}}
{"type":"log","data":{"level":"info","message":"Cloning repository...","source":"git"}}
{"type":"status","data":{"stage":"building","progress":50,"message":"Building image"}}
{"type":"result","data":{"success":true,"message":"SUCCESS","duration":"45s"}}
```

## Architecture

```
┌─────────────────┐    TCP     ┌─────────────────┐
│   CLI Client    │◄──────────►│   TCP Server    │
│                 │            │                 │
│ • Interactive   │            │ • Authentication│
│ • Colored logs  │            │ • Log streaming │
│ • Progress bars │            │ • Git cloning   │
└─────────────────┘            │ • Docker builds │
                               └─────────────────┘
                                        │
                                        ▼
                               ┌─────────────────┐
                               │ Docker Daemon   │
                               │                 │
                               │ • Build images  │
                               │ • Run containers│
                               └─────────────────┘
```

## What's New - Environment Variables & Secrets Support

We've significantly enhanced the CD tool with comprehensive environment variables and secrets management:

### 🔐 New Security Features
- **Environment file support** - Load variables from `.env`, `.env.staging`, `.env.production` files
- **Command-line variables** - Pass variables directly via CLI arguments
- **GitHub secrets integration** - Automatically include repository secrets as environment variables
- **Secrets masking** - Sensitive values are automatically hidden in deployment logs

### 🚀 Enhanced GitHub Actions
- **Automatic secret injection** - Set `include_github_secrets: true` to pass all GitHub secrets to containers
- **Environment-specific configs** - Use different `.env` files for staging/production
- **Build metadata** - Automatically inject build numbers, commit SHAs, and deployment context
- **Flexible configuration** - Combine environment files, command-line vars, and secrets

### 💻 CLI Enhancements
New CLI options for environment variable management:
- `-env-file` - Load variables from environment files
- `-env-vars` - Pass comma-separated environment variables
- `-secrets-file` - Load secrets from protected files
- `-secrets` - Pass secrets via command line

### 🛡️ Security Best Practices
- Environment variables and secrets are handled separately
- Secrets always override environment variables
- Automatic masking of sensitive values in logs
- Support for environment-specific configuration files

For complete documentation and examples, see:
- [Environment Variables Guide](docs/ENVIRONMENT_VARIABLES.md)
- [Example Workflow](examples/deploy-with-env-vars.yml)
- [Environment File Template](.env.example)

## License

MIT License - see LICENSE file for details.