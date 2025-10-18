# CD - Continuous Deployment Tool

A lightweight continuous deployment tool using pure TCP for real-time communication. Deploy applications from Git repositories to Docker containers with live log streaming.

## Features

- **Pure TCP Protocol** - Fast, persistent connections for real-time updates
- **Real-time Logging** - Stream deployment logs as they happen
- **Docker Integration** - Automatic container building and deployment
- **Environment Variables** - Inject configuration and secrets into containers
- **Interactive CLI** - User-friendly interface with colored output and progress bars
- **API Authentication** - Secure server access with API keys
- **SSH Key Support** - Server-side SSH key configuration for private repositories

## Installation

### Download Binary

```bash
curl -L -o cd-cli https://github.com/RestartFU/cd/releases/latest/download/cd-cli-linux-amd64
chmod +x cd-cli
```

### Build from Source

```bash
git clone https://github.com/RestartFU/cd.git
cd cd
make build
```

Binaries will be in `bin/`:
- `bin/cd-server` - Server component
- `bin/cd-cli` - CLI client

## Quick Start

### 1. Configure Server

Create `config/config.toml`:

```toml
listen_addr = ':8080'
api_keys = ['your-secret-api-key']
ssh_key_path = ''  # Optional: path to SSH private key for git cloning
```

### 2. Start Server

```bash
./bin/cd-server
```

The server will:
- Listen on the configured port (default: 8080)
- Accept authenticated TCP connections
- Clone repositories and build Docker containers

### 3. Deploy with CLI

**Interactive Mode:**

```bash
./bin/cd-cli
```

**Direct Deployment:**

```bash
./bin/cd-cli \
  -server=localhost:8080 \
  -key=your-secret-api-key \
  -git=git@github.com:user/repo.git
```

**With Environment Variables:**

```bash
./bin/cd-cli \
  -server=localhost:8080 \
  -key=your-secret-api-key \
  -git=git@github.com:user/repo.git \
  -env=production \
  -env-vars="API_KEY=value,DEBUG=false" \
  -secrets-file=.secrets
```

## CLI Options

```
-server string       Server address (default "localhost:8080")
-key string          API key for authentication
-git string          Git repository URL
-env string          Environment name (default "production")
-env-file string     Path to environment variables file (.env format)
-env-vars string     Comma-separated environment variables (KEY=VALUE)
-secrets-file string Path to secrets file (.env format)
-secrets string      Comma-separated secrets (KEY=VALUE)
-diagnose            Run Docker diagnostics
-help                Show help message
```

## SSH Key Configuration

SSH keys are configured on the server side in `config/config.toml`:

```toml
ssh_key_path = '/home/user/.ssh/id_ed25519'
```

**Fallback behavior:**
- If `ssh_key_path` is set, the server uses that key
- If empty or not set, defaults to `/home/restart/.ssh/id_ed25519`

This allows the server to clone private repositories without exposing SSH keys to clients.

## Environment Variables

### From File

Create `.env` file:

```
NODE_ENV=production
PORT=3000
API_URL=https://api.example.com
```

Deploy:

```bash
./bin/cd-cli -env-file=.env -git=git@github.com:user/repo.git
```

### From Command Line

```bash
./bin/cd-cli -env-vars="PORT=3000,DEBUG=true" -git=git@github.com:user/repo.git
```

### Secrets

Separate secrets from environment variables:

```bash
./bin/cd-cli \
  -env-file=.env.production \
  -secrets-file=.secrets \
  -git=git@github.com:user/repo.git
```

Secrets override environment variables with the same name and are masked in logs.

## Repository Requirements

Your repository must contain a `Dockerfile`:

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]
```

The server will:
1. Clone the repository
2. Build the Docker image
3. Start a container with environment variables injected

## Protocol

The client and server communicate using JSON messages over TCP:

**Authentication:**
```json
{"type":"auth","timestamp":"2024-01-01T00:00:00Z","data":{"api_key":"secret"}}
```

**Deploy Request:**
```json
{
  "type":"deploy",
  "timestamp":"2024-01-01T00:00:00Z",
  "data":{
    "git_url":"git@github.com:user/repo.git",
    "environment":"production",
    "env_vars":{"PORT":"3000"},
    "secrets":{"DB_PASSWORD":"secret"},
    "ssh_key_path":"/path/to/key"
  }
}
```

**Server Responses:**
```json
{"type":"log","data":{"level":"info","message":"Cloning repository","source":"git"}}
{"type":"status","data":{"stage":"building","progress":50,"message":"Building image"}}
{"type":"result","data":{"success":true,"message":"SUCCESS","duration":"45s"}}
{"type":"error","data":{"code":"DEPLOY_FAILED","message":"Build failed"}}
```

## Architecture

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  CLI Client │◄───TCP──►│  CD Server  │────────►│   Docker    │
│             │         │             │         │   Daemon    │
│ • Auth      │         │ • Git clone │         │             │
│ • Deploy    │         │ • Build     │         │ • Build     │
│ • Stream    │         │ • Deploy    │         │ • Run       │
└─────────────┘         └─────────────┘         └─────────────┘
```

## Development

```bash
# Build both components
make build

# Build separately
go build -o bin/cd-server ./cmd/server
go build -o bin/cd-cli ./cmd/cli

# Format code
make fmt

# Run tests
make test

# Clean build artifacts
make clean
```

## Docker Deployment

Run the CD server in Docker:

```bash
docker run -d \
  --name cd-server \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/config:/app/config \
  -v ~/.ssh:/root/.ssh:ro \
  cd-server:latest
```

## Troubleshooting

### Docker Socket Issues

Run diagnostics:

```bash
./bin/cd-cli -diagnose
```

### SSH Key Problems

Verify key permissions:

```bash
chmod 600 ~/.ssh/id_ed25519
ssh-add ~/.ssh/id_ed25519
```

Test Git access:

```bash
ssh -T git@github.com
```

### Connection Issues

Check server is running:

```bash
netstat -tuln | grep 8080
```

Test connection:

```bash
telnet localhost 8080
```

## License

MIT License - see LICENSE file for details.