# Docker Deployment Guide

This guide explains how to deploy the CD Tool server using Docker for production environments.

## Quick Start

### Option 1: Automated Setup Script (Recommended)

```bash
# Run the automated setup script
./docker-setup.sh

# Or with custom options
./docker-setup.sh --port 9000 --logs
```

### Option 2: Docker Compose

```bash
# Basic setup
docker-compose up -d

# With monitoring and caching
docker-compose --profile cache --profile monitoring up -d
```

### Option 3: Manual Docker Commands

```bash
# Build the image
docker build -t cd-tool:latest .

# Run the container
docker run -d \
  --name cd-server \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/config.toml:/app/config.toml:ro \
  cd-tool:latest
```

## Prerequisites

### System Requirements

- **Docker**: Version 20.10 or later
- **Docker Compose**: Version 2.0 or later (optional)
- **Operating System**: Linux, macOS, or Windows with WSL2
- **Memory**: Minimum 512MB RAM, recommended 1GB+
- **Disk Space**: 2GB for images and temporary files

### Installation

**macOS:**
```bash
brew install --cask docker
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update
sudo apt-get install docker.io docker-compose-plugin
sudo usermod -aG docker $USER
```

**Windows:**
Download and install Docker Desktop from [docker.com](https://docker.com)

## Configuration

### Environment Variables

The Docker container supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `SERVER_PORT` | Internal server port | `8080` |
| `MAX_DEPLOYMENTS` | Maximum concurrent deployments | `5` |
| `DOCKER_HOST` | Docker daemon socket | Auto-detected |

### Config File

Create a `config.toml` file in your project directory:

```toml
# CD Tool Server Configuration
listen_addr = ':8080'

# API keys for authentication (CHANGE THESE!)
api_keys = [
    'your-secure-api-key-here',
    'another-secure-key'
]

# Optional settings
log_level = 'info'
max_deployments = 5
```

### Volume Mounts

The container requires these volume mounts:

| Host Path | Container Path | Purpose |
|-----------|----------------|---------|
| `/var/run/docker.sock` | `/var/run/docker.sock` | Docker access |
| `./config.toml` | `/app/config.toml` | Configuration |
| `./logs` | `/app/logs` | Log files (optional) |

## Security Considerations

### Docker Socket Access

The container needs access to the Docker socket to manage deployments. This is equivalent to root access on the host system.

**Security measures:**
- Run container as non-root user (configured in Dockerfile)
- Use read-only config mounts
- Implement API key authentication
- Consider using Docker-in-Docker instead of socket mounting

### Network Security

```bash
# Create custom network
docker network create cd-network

# Run with custom network
docker run -d \
  --name cd-server \
  --network cd-network \
  -p 127.0.0.1:8080:8080 \  # Bind to localhost only
  cd-tool:latest
```

### API Key Management

```bash
# Generate secure API keys
openssl rand -hex 32

# Set via environment variable
docker run -d \
  --name cd-server \
  -e API_KEYS="$(cat /path/to/secure/api-keys)" \
  cd-tool:latest
```

## Production Deployment

### Using Docker Compose (Recommended)

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  cd-server:
    build: .
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./config.toml:/app/config.toml:ro
      - ./logs:/app/logs
    environment:
      - LOG_LEVEL=warn
    networks:
      - cd-network
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - cd-server
    networks:
      - cd-network

networks:
  cd-network:
    driver: bridge
```

### SSL/TLS Configuration

Create `nginx.conf` for HTTPS:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream cd-server {
        server cd-server:8080;
    }

    server {
        listen 80;
        server_name your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        location / {
            proxy_pass http://cd-server;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## Monitoring and Logging

### Health Checks

The container includes built-in health checks:

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' cd-server

# View health check logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' cd-server
```

### Log Management

```bash
# View logs
docker logs cd-server

# Follow logs
docker logs -f cd-server

# View last 100 lines
docker logs --tail 100 cd-server

# Configure log rotation
docker run -d \
  --name cd-server \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  cd-tool:latest
```

### Monitoring with Prometheus

Enable monitoring profile:

```bash
docker-compose --profile monitoring up -d
```

Access Grafana at http://localhost:3000 (admin/admin)

## Scaling and High Availability

### Load Balancing

```yaml
# docker-compose.scale.yml
version: '3.8'

services:
  cd-server:
    build: .
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./config.toml:/app/config.toml:ro
    environment:
      - LOG_LEVEL=info
    networks:
      - cd-network
    deploy:
      replicas: 3

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "8080:80"
    volumes:
      - ./nginx-lb.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - cd-server
    networks:
      - cd-network

networks:
  cd-network:
    driver: bridge
```

### Nginx Load Balancer Config

```nginx
events {
    worker_connections 1024;
}

http {
    upstream cd-servers {
        server cd-server_1:8080;
        server cd-server_2:8080;
        server cd-server_3:8080;
    }

    server {
        listen 80;
        location / {
            proxy_pass http://cd-servers;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

## Troubleshooting

### Common Issues

**Container fails to start:**
```bash
# Check logs
docker logs cd-server

# Check Docker socket permissions
ls -la /var/run/docker.sock

# Test Docker access
docker run hello-world
```

**Port already in use:**
```bash
# Find what's using the port
lsof -i :8080

# Use different port
docker run -p 9000:8080 cd-tool:latest
```

**Docker socket access denied:**
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker

# On macOS, restart Docker Desktop
```

### Debug Mode

Run container in debug mode:

```bash
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/config.toml:/app/config.toml:ro \
  -e LOG_LEVEL=debug \
  cd-tool:latest
```

### Performance Optimization

**Resource Limits:**
```bash
docker run -d \
  --name cd-server \
  --memory=1g \
  --cpus=1.0 \
  --oom-kill-disable=false \
  cd-tool:latest
```

**Build Optimization:**
```bash
# Multi-stage build (already configured in Dockerfile)
# Use .dockerignore to reduce context size
# Cache dependencies separately from source code
```

## Backup and Recovery

### Configuration Backup

```bash
# Backup config and volumes
docker run --rm \
  -v cd-server_config:/source \
  -v $(pwd)/backup:/backup \
  alpine tar czf /backup/cd-config-$(date +%Y%m%d).tar.gz -C /source .
```

### Container Recovery

```bash
# Save container image
docker save cd-tool:latest | gzip > cd-tool-backup.tar.gz

# Load container image
gunzip -c cd-tool-backup.tar.gz | docker load
```

## Maintenance

### Updates

```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
./docker-setup.sh --rebuild

# Or with compose
docker-compose down
docker-compose up --build -d
```

### Cleanup

```bash
# Remove unused images
docker image prune -f

# Remove unused volumes
docker volume prune -f

# Remove unused containers
docker container prune -f

# Full cleanup
./docker-setup.sh --remove
```

## Integration Examples

### CI/CD Pipeline (GitHub Actions)

```yaml
name: Deploy CD Server

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build and deploy
        run: |
          ./docker-setup.sh --rebuild
          
      - name: Health check
        run: |
          sleep 10
          curl -f http://localhost:8080/health || exit 1
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cd-server
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cd-server
  template:
    metadata:
      labels:
        app: cd-server
    spec:
      containers:
      - name: cd-server
        image: cd-tool:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: docker-sock
          mountPath: /var/run/docker.sock
        - name: config
          mountPath: /app/config.toml
          subPath: config.toml
      volumes:
      - name: docker-sock
        hostPath:
          path: /var/run/docker.sock
      - name: config
        configMap:
          name: cd-config
```

## Support

For issues and questions:
- Check logs: `docker logs cd-server`
- Run diagnostics: `./bin/cd-cli -diagnose`
- Review documentation: [README.md](../README.md)
- Open issues: [GitHub Issues](https://github.com/RestartFU/cd/issues)