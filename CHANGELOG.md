# Changelog

All notable changes to the CD Tool project will be documented in this file.

## [v2.0.0] - Environment Variables & Secrets Support

### 🚀 Major Features Added

#### Environment Variables & Secrets Management
- **Environment file support** - Load variables from `.env`, `.env.staging`, `.env.production` files
- **Command-line variables** - Pass environment variables directly via CLI arguments
- **Secrets file support** - Secure loading of sensitive configuration from protected files
- **GitHub secrets integration** - Automatically include all GitHub repository secrets as environment variables
- **Variable precedence handling** - Secrets override environment variables with clear precedence rules

#### Enhanced Security
- **Automatic secrets masking** - Sensitive values are automatically hidden in deployment logs
- **Secure environment injection** - Variables and secrets are securely passed to Docker containers
- **Environment-specific configurations** - Support for different configurations per deployment environment
- **Format validation** - Automatic validation of environment file formats with helpful error messages

#### CLI Enhancements
- **New CLI flags**:
  - `-env-file` - Path to environment variables file (.env format)
  - `-env-vars` - Comma-separated environment variables (KEY=VALUE,KEY2=VALUE2)
  - `-secrets-file` - Path to secrets file (.env format)
  - `-secrets` - Comma-separated secrets (KEY=VALUE,KEY2=VALUE2)
- **Enhanced help documentation** - Updated CLI help with examples and usage patterns
- **Interactive mode improvements** - Better prompts and validation in interactive mode

#### GitHub Actions Integration
- **New workflow inputs**:
  - `env_file` - Path to environment file to load
  - `env_vars` - Custom environment variables to pass to container
  - `include_github_secrets` - Include all GitHub secrets as environment variables
- **Automatic GitHub context injection** - Repository, commit, and workflow metadata automatically available
- **Environment-specific deployment** - Support for different environment files per deployment target
- **Enhanced security handling** - Proper secrets management in CI/CD workflows

### 🔧 Technical Improvements

#### Protocol Enhancements
- **Extended DeployMessage** - Added `EnvVars` and `Secrets` fields to deployment protocol
- **New client methods** - Added `DeployWithEnv()` method for environment variable support
- **Backward compatibility** - Existing `Deploy()` method maintained for compatibility

#### Docker Integration
- **Environment variable injection** - Docker containers receive all specified environment variables
- **Secure secrets handling** - Secrets are passed securely to containers without logging
- **Container configuration** - Enhanced container creation with environment variable support

#### Handler Improvements
- **DeployWithEnv method** - New deployment method that accepts environment variables and secrets
- **Environment merging** - Intelligent merging of environment variables from multiple sources
- **Deployment environment injection** - Automatic `DEPLOYMENT_ENV` variable injection

### 📚 Documentation & Examples

#### New Documentation
- **Environment Variables Guide** (`docs/ENVIRONMENT_VARIABLES.md`) - Comprehensive documentation covering:
  - Usage patterns and best practices
  - Security considerations
  - File format specifications
  - Troubleshooting guide
  - Migration guide from previous versions

#### Enhanced Workflow Templates
- **Updated simple template** - Shows basic environment variable usage
- **Enhanced advanced template** - Demonstrates comprehensive environment management
- **Complete example workflow** (`examples/deploy-with-env-vars.yml`) - Real-world deployment example

#### Example Files
- **Environment file template** (`.env.example`) - Sample environment configuration
- **Usage examples** - Multiple examples showing different use cases

### 🛡️ Security Features

#### Secrets Management
- **Automatic masking** - Sensitive values never appear in deployment logs
- **Precedence handling** - Secrets always override environment variables
- **Secure transmission** - All secrets are transmitted securely over TCP protocol
- **File validation** - Environment and secrets files are validated before processing

#### Best Practices
- **Environment-specific files** - Support for `.env.staging`, `.env.production`, etc.
- **Git ignore recommendations** - Clear guidance on what files should not be committed
- **Security scanning** - Basic security checks for exposed secrets in code

### 🔄 Protocol Changes

#### Message Format Updates
```json
{
  "type": "deploy",
  "data": {
    "git_url": "https://github.com/user/repo.git",
    "environment": "production",
    "env_vars": {
      "API_KEY": "value",
      "DEBUG": "false"
    },
    "secrets": {
      "DB_PASSWORD": "secret-value"
    }
  }
}
```

#### Backward Compatibility
- All existing deployments continue to work without changes
- Optional fields ensure compatibility with older clients
- Graceful degradation when environment variables are not provided

### 🧪 Testing Improvements

#### Test Coverage
- **Mock adapter updates** - All mock Docker adapters updated to support environment variables
- **Integration tests** - Enhanced integration tests covering environment variable flows
- **Format validation tests** - Tests for environment file parsing and validation

#### Test Fixes
- **Interface compliance** - All mock implementations updated to match new interface signatures
- **Build verification** - Continuous integration ensures all components build successfully
- **Comprehensive test suite** - All existing functionality continues to pass tests

### 🚀 Usage Examples

#### CLI Usage
```bash
# Deploy with environment file
cd-cli -server=cd-server:8080 -key=api-key \
       -git=https://github.com/user/repo.git \
       -env-file=.env.production

# Deploy with command-line variables
cd-cli -server=cd-server:8080 -key=api-key \
       -git=https://github.com/user/repo.git \
       -env-vars="API_KEY=value,DEBUG=true"

# Deploy with secrets file
cd-cli -server=cd-server:8080 -key=api-key \
       -git=https://github.com/user/repo.git \
       -secrets-file=.secrets
```

#### GitHub Actions Usage
```yaml
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

### 📋 Migration Guide

#### For Existing Users
1. **No immediate action required** - All existing deployments continue to work
2. **Optional enhancement** - Add environment files to your repositories
3. **Secrets migration** - Move hardcoded values to GitHub secrets
4. **Workflow updates** - Update GitHub Actions workflows to use new features

#### For New Users
1. **Create environment files** - Use `.env.example` as a template
2. **Set up GitHub secrets** - Add `CD_SERVER_HOST` and `CD_API_KEY` to repository
3. **Copy workflow templates** - Use provided templates for quick setup
4. **Review documentation** - Follow the comprehensive environment variables guide

### 🎯 What's Next

This release establishes a solid foundation for environment variable and secrets management. Future enhancements may include:

- **Encrypted environment files** - Support for encrypted `.env` files
- **External secrets management** - Integration with HashiCorp Vault, AWS Secrets Manager
- **Dynamic configuration** - Runtime configuration updates without redeployment
- **Audit logging** - Detailed logging of environment variable usage

---

## [v1.0.0] - Initial TCP Release

### Features
- Pure TCP communication protocol
- Real-time log streaming
- Interactive CLI client
- Docker integration
- Git repository support
- GitHub Actions integration
- Multi-environment support

---

**Note**: This project follows [Semantic Versioning](https://semver.org/). The environment variables feature represents a major enhancement that significantly expands the tool's capabilities while maintaining backward compatibility.