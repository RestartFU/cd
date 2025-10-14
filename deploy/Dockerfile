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
COPY config/config.toml* ./

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
