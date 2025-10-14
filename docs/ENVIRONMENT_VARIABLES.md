# Environment Variables and Secrets

This document explains how to use environment variables and secrets with the CD tool to securely pass configuration and credentials to your deployed containers.

## Overview

The CD tool supports multiple ways to inject environment variables and secrets into your containers during deployment:

1. **Environment Files** - Load variables from `.env` files
2. **Command Line Variables** - Pass variables directly via CLI
3. **GitHub Secrets** - Automatically include GitHub repository secrets
4. **Deployment Context** - Include GitHub workflow context as variables

## Quick Start

### Basic Usage with GitHub Actions

```yaml
jobs:
  deploy:
    uses: ./.github/workflows/deploy.yml
    with:
      environment: production
      env_vars: "API_KEY=my-api-key,DEBUG=false"
      include_github_secrets: true
    secrets:
      CD_SERVER_HOST: ${{ secrets.CD_SERVER_HOST }}
      CD_API_KEY: ${{ secrets.CD_API_KEY }}
```

### CLI Usage

```bash
# Deploy with environment variables
cd-cli -server="cd-server:8080" -key="api-key" \
       -git="https://github.com/user/repo.git" \
       -env-vars="API_KEY=value,DEBUG=true"

# Deploy with environment file
cd-cli -server="cd-server:8080" -key="api-key" \
       -git="https://github.com/user/repo.git" \
       -env-file=".env.production"

# Deploy with secrets file
cd-cli -server="cd-server:8080" -key="api-key" \
       -git="https://github.com/user/repo.git" \
       -secrets-file=".secrets"
```

## Environment Files

### Format

Environment files use the standard `.env` format:

```bash
# .env file example
API_KEY=your-api-key-here
DATABASE_URL=postgresql://user:pass@localhost:5432/db
DEBUG=false
PORT=3000

# Comments are supported
# Empty lines are ignored

# Quotes are optional but recommended for values with spaces
APP_NAME="My Application"
DESCRIPTION='Application with spaces in description'
```

### File Naming Conventions

- `.env` - Default environment file
- `.env.local` - Local development overrides
- `.env.staging` - Staging environment variables
- `.env.production` - Production environment variables
- `.env.testing` - Testing environment variables

### Security Best Practices

⚠️ **Never commit sensitive environment files to git!**

Add them to your `.gitignore`:

```gitignore
.env
.env.local
.env.*.local
.secrets
.secrets.*
```

## GitHub Actions Integration

### Basic Configuration

```yaml
name: Deploy with Environment Variables

on:
  push:
    branches: [main]

jobs:
  deploy:
    uses: ./.github/workflows/deploy.yml
    with:
      environment: production
      env_file: ".env.production"
      env_vars: "BUILD_NUMBER=${{ github.run_number }}"
      include_github_secrets: true
    secrets:
      CD_SERVER_HOST: ${{ secrets.CD_SERVER_HOST }}
      CD_API_KEY: ${{ secrets.CD_API_KEY }}
```

### Advanced Configuration

```yaml
jobs:
  deploy:
    uses: ./.github/workflows/deploy.yml
    with:
      environment: ${{ inputs.environment }}
      # Load environment-specific file
      env_file: ".env.${{ inputs.environment }}"
      # Pass custom variables
      env_vars: |
        BUILD_NUMBER=${{ github.run_number }}
        COMMIT_SHA=${{ github.sha }}
        DEPLOYMENT_TIME=${{ github.event.head_commit.timestamp }}
      # Include all GitHub secrets
      include_github_secrets: true
    secrets:
      CD_SERVER_HOST: ${{ secrets.CD_SERVER_HOST }}
      CD_API_KEY: ${{ secrets.CD_API_KEY }}
```

### Including GitHub Secrets

When `include_github_secrets: true` is set, the following variables are automatically included:

```bash
GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }}
GITHUB_REPOSITORY=${{ github.repository }}
GITHUB_REF=${{ github.ref }}
GITHUB_SHA=${{ github.sha }}
GITHUB_ACTOR=${{ github.actor }}
GITHUB_WORKFLOW=${{ github.workflow }}
GITHUB_RUN_ID=${{ github.run_id }}
GITHUB_RUN_NUMBER=${{ github.run_number }}
```

Plus all repository secrets defined in your GitHub repository settings.

## Environment Variable Precedence

When multiple sources provide the same variable, the precedence order is:

1. **Secrets** (highest priority)
2. **Command line variables** (`env_vars`)
3. **Environment files** (`env_file`)
4. **GitHub context variables**
5. **Default values** (lowest priority)

Example:
```yaml
# If you have DATABASE_URL in multiple places:
env_file: ".env"              # DATABASE_URL=file-value
env_vars: "DATABASE_URL=cli-value"  # This overrides the file
# But secrets would override both
```

## Security Considerations

### Secrets vs Environment Variables

- **Environment Variables**: Non-sensitive configuration (API endpoints, feature flags, ports)
- **Secrets**: Sensitive data (API keys, passwords, tokens, certificates)

### Best Practices

1. **Use Secrets for Sensitive Data**
   ```yaml
   # ❌ Don't put sensitive data in env_vars
   env_vars: "API_KEY=secret-key-123"
   
   # ✅ Use GitHub secrets instead
   secrets:
     API_KEY: ${{ secrets.API_KEY }}
   ```

2. **Validate Environment Variables**
   ```dockerfile
   # In your Dockerfile
   RUN test -n "$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
   ```

3. **Use Environment-Specific Files**
   ```yaml
   # Different configs for different environments
   env_file: ".env.${{ inputs.environment }}"
   ```

4. **Log Safely**
   ```bash
   # The CD tool automatically masks sensitive values in logs
   # Secrets are never shown in deployment logs
   ```

## Container Access

### In Your Application

Environment variables are available to your application just like any other container environment variables:

**Node.js:**
```javascript
const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
const isDebug = process.env.DEBUG === 'true';
```

**Python:**
```python
import os

api_key = os.environ.get('API_KEY')
db_url = os.environ.get('DATABASE_URL')
is_debug = os.environ.get('DEBUG', 'false').lower() == 'true'
```

**Go:**
```go
import "os"

apiKey := os.Getenv("API_KEY")
dbURL := os.Getenv("DATABASE_URL")
isDebug := os.Getenv("DEBUG") == "true"
```

### Docker Environment

Variables are passed to Docker containers using the `-e` flag:

```bash
# The CD tool automatically converts this:
docker run -e API_KEY=value -e DEBUG=true your-image

# You can also use them in your Dockerfile:
ENV NODE_ENV=production
```

## Examples

### Example 1: Web Application with Database

**.env.production:**
```bash
NODE_ENV=production
PORT=3000
DATABASE_URL=postgresql://prod-db:5432/myapp
REDIS_URL=redis://prod-redis:6379
LOG_LEVEL=info
```

**GitHub Workflow:**
```yaml
deploy:
  uses: ./.github/workflows/deploy.yml
  with:
    environment: production
    env_file: ".env.production"
    env_vars: "BUILD_NUMBER=${{ github.run_number }}"
  secrets:
    CD_SERVER_HOST: ${{ secrets.CD_SERVER_HOST }}
    CD_API_KEY: ${{ secrets.CD_API_KEY }}
    DATABASE_PASSWORD: ${{ secrets.DATABASE_PASSWORD }}
    REDIS_PASSWORD: ${{ secrets.REDIS_PASSWORD }}
```

### Example 2: Microservice with External APIs

**Workflow with Manual Input:**
```yaml
on:
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [staging, production]
      feature_flags:
        description: "Feature flags (FLAG1=true,FLAG2=false)"
        type: string

jobs:
  deploy:
    uses: ./.github/workflows/deploy.yml
    with:
      environment: ${{ inputs.environment }}
      env_vars: |
        DEPLOYMENT_ID=${{ github.run_id }}
        ${{ inputs.feature_flags }}
      include_github_secrets: true
```

### Example 3: CLI Deployment with Multiple Sources

```bash
# Deploy with combined sources
cd-cli \
  -server="cd-server.example.com:8080" \
  -key="$CD_API_KEY" \
  -git="https://github.com/myorg/myapp.git" \
  -env="production" \
  -env-file=".env.production" \
  -env-vars="BUILD_ID=123,FEATURE_X=enabled" \
  -secrets-file=".secrets.production"
```

## Troubleshooting

### Common Issues

1. **Variables Not Available in Container**
   ```bash
   # Check if variables are being passed correctly
   docker exec -it container-name env | grep YOUR_VAR
   ```

2. **File Not Found Errors**
   ```bash
   # Ensure environment files exist and are accessible
   ls -la .env*
   ```

3. **Invalid Format Errors**
   ```bash
   # Validate .env file format
   grep -v '^#' .env | grep -v '^$' | grep -v '='
   ```

### Debugging

Enable debug logging to see what variables are being processed:

```yaml
env_vars: "DEBUG=true,LOG_LEVEL=debug"
```

Check deployment logs for environment variable processing:
- Number of variables loaded from files
- Number of variables from command line
- Number of secrets provided (count only, not values)

### Security Troubleshooting

If sensitive data appears in logs:
1. Use secrets instead of environment variables
2. Check that the variable name doesn't contain "password", "key", "secret", or "token"
3. Report the issue - secrets should never appear in logs

## Migration Guide

### From v1.0 (HTTP-based)

If migrating from the old HTTP-based CD tool:

**Old approach:**
```bash
curl -X POST cd-server/deploy \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"git_url":"...", "env": {"KEY":"VALUE"}}'
```

**New approach:**
```bash
cd-cli -server="cd-server:8080" -key="$API_KEY" \
       -git="..." -env-vars="KEY=VALUE"
```

### From Manual Docker Deployment

**Old approach:**
```bash
docker run -e API_KEY=value -e DB_URL=url my-app
```

**New approach:**
```bash
# Create .env file
echo "API_KEY=value" > .env
echo "DB_URL=url" >> .env

# Deploy with CD tool
cd-cli -env-file=".env" ...
```

## API Reference

### CLI Flags

| Flag | Description | Example |
|------|-------------|---------|
| `-env-file` | Path to environment file | `-env-file=".env.prod"` |
| `-env-vars` | Comma-separated variables | `-env-vars="A=1,B=2"` |
| `-secrets-file` | Path to secrets file | `-secrets-file=".secrets"` |
| `-secrets` | Comma-separated secrets | `-secrets="KEY=value"` |

### GitHub Actions Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `env_file` | Environment file path | `""` |
| `env_vars` | Command line variables | `""` |
| `include_github_secrets` | Include GitHub secrets | `false` |

### Environment File Format

```bash
# Valid formats:
KEY=value
KEY="value with spaces"
KEY='single quoted value'

# Invalid formats:
KEY value        # Missing =
=value          # Missing key
KEY=            # Empty value (use KEY="" instead)
```

---

For more information, see the main [README.md](../README.md) or [GitHub Actions documentation](https://docs.github.com/en/actions).