# Project Structure

This document describes the organization of the CD Tool project files and directories.

## Directory Structure

```
cd/
├── .github/                    # GitHub Actions workflows and templates
│   └── workflows/
├── bin/                        # Compiled binaries (generated)
├── cmd/                        # Main application entry points
│   ├── cli/                    # CLI client source code
│   └── server/                 # TCP server source code
├── config/                     # Configuration files
│   └── config.toml            # Server configuration
├── deploy/                     # Docker and deployment files
│   ├── Dockerfile             # Docker container definition
│   └── docker-compose.yml     # Multi-service Docker setup
├── docs/                       # Documentation
├── examples/                   # Example configurations and workflows
├── internal/                   # Internal Go packages (not for external use)
├── scripts/                    # Build and utility scripts
│   ├── docker-setup.sh        # Automated Docker deployment
│   ├── example.sh             # Example deployment script
│   ├── quick-start.sh         # Quick start script
│   └── test.sh               # Test runner script
├── workflow-templates/         # GitHub Actions workflow templates
├── .dockerignore              # Docker build ignore file
├── .gitignore                 # Git ignore file
├── CHANGELOG.md              # Version history
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── Makefile                  # Build automation
├── PROJECT_STRUCTURE.md      # This file
└── README.md                 # Main project documentation
```

## Directory Purposes

### `/config/`
Contains all configuration files for the CD Tool server:
- `config.toml` - Main server configuration with API keys, listen address, etc.

### `/deploy/`
Contains Docker and deployment-related files:
- `Dockerfile` - Container definition for the CD server
- `docker-compose.yml` - Multi-service deployment with optional monitoring

### `/scripts/`
Contains utility and automation scripts:
- `docker-setup.sh` - Comprehensive Docker deployment automation
- `test.sh` - Full test suite runner
- `quick-start.sh` - Quick development setup
- `example.sh` - Example deployment demonstrations

### `/cmd/`
Contains the main application entry points:
- `cli/` - Command-line interface client
- `server/` - TCP server application

### `/internal/`
Contains internal Go packages (following Go conventions):
- Private packages not intended for external use
- Protocol handlers, configuration, adapters, etc.

### `/docs/`
Contains detailed documentation:
- Deployment guides, API documentation, troubleshooting, etc.

### `/examples/`
Contains example configurations and usage patterns:
- Sample workflow files, environment configurations, etc.

### `/.github/`
Contains GitHub-specific files:
- CI/CD workflows, issue templates, etc.

### `/workflow-templates/`
Contains reusable GitHub Actions workflow templates:
- Templates that other repositories can use for CD deployments

## Key Files

### Configuration
- `config/config.toml` - Server configuration (API keys, ports, etc.)

### Build & Development
- `Makefile` - Build targets and development commands
- `go.mod` / `go.sum` - Go module dependencies
- `scripts/test.sh` - Test automation

### Deployment
- `deploy/Dockerfile` - Production container image
- `deploy/docker-compose.yml` - Multi-service deployment
- `scripts/docker-setup.sh` - Automated Docker setup

### Documentation
- `README.md` - Main project documentation
- `CHANGELOG.md` - Version history and changes
- `docs/` - Detailed guides and documentation

## Usage Patterns

### Development
```bash
# Build the project
make build

# Run tests
make test
# or
./scripts/test.sh

# Start development server
./bin/cd-server
```

### Docker Deployment
```bash
# Automated setup
./scripts/docker-setup.sh

# Manual Docker Compose
cd deploy/
docker-compose up -d
```

### Configuration
```bash
# Edit server configuration
vim config/config.toml

# Restart server to apply changes
docker restart cd-server
```

## Migration Notes

If you're updating from an older version of the project, note these path changes:

| Old Location | New Location |
|--------------|--------------|
| `./config.toml` | `./config/config.toml` |
| `./Dockerfile` | `./deploy/Dockerfile` |
| `./docker-compose.yml` | `./deploy/docker-compose.yml` |
| `./docker-setup.sh` | `./scripts/docker-setup.sh` |
| `./test.sh` | `./scripts/test.sh` |
| `./quick-start.sh` | `./scripts/quick-start.sh` |
| `./example.sh` | `./scripts/example.sh` |

## Benefits of This Structure

1. **Clear Separation**: Different types of files are organized into logical directories
2. **Scalability**: Easy to find and maintain files as the project grows
3. **Standard Conventions**: Follows Go and Docker community best practices
4. **Clean Root**: Reduces clutter in the project root directory
5. **Easy Navigation**: Developers can quickly locate files by purpose
6. **Build Efficiency**: Better Docker layer caching and ignore patterns

## Contributing

When adding new files, please follow this structure:
- Configuration files → `config/`
- Build/deployment files → `deploy/`
- Utility scripts → `scripts/`
- Documentation → `docs/`
- Examples → `examples/`
- Source code → `cmd/` or `internal/`
