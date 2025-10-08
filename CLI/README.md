# Envini CLI - Secure Environment Management Tool

A command-line interface for managing environment variables and secrets across GitHub repositories with secure authentication and encryption.

## 🚀 Quick Start

### 1. Build the CLI
```bash
cd CLI
go mod tidy
go build -o envini
```

### 2. Authenticate with GitHub
```bash
./envini auth
```
This will guide you through the GitHub OAuth device flow.

### 3. Start managing secrets
```bash
# Upload a secret file (auto-detects current git repo)
./envini upload .env

# Download the latest version
./envini download .env.downloaded

# List all versions
./envini versions
```

## 📋 Commands

### Authentication
```bash
envini auth                 # Start GitHub OAuth authentication flow
```

### Repository Management
```bash
envini repos                # List your GitHub repositories
envini repos-with-versions  # List repositories with secret version info
```

### Secret Management

#### Upload Secrets
```bash
# Auto-detect repository from git remote
envini upload .env                    # Upload to development tag (default)
envini upload .env --tag=production   # Upload to production tag
envini upload .env --tag=staging      # Upload to staging tag

# Explicit repository specification
envini upload <owner> <repo> .env
envini upload <owner> <repo> .env --tag=production
```

#### Download Secrets
```bash
# Auto-detect repository from git remote
envini download .env.downloaded                      # Latest from development tag
envini download .env.downloaded --tag=production     # Latest from production tag
envini download .env.downloaded --version=1          # Specific version number
envini download .env.downloaded --version=2 --tag=production  # Version 2 from production tag

# Explicit repository specification
envini download <owner> <repo> .env.downloaded
envini download <owner> <repo> .env.downloaded --tag=production
envini download <owner> <repo> .env.downloaded --version=1
```

#### List Secret Versions
```bash
# Auto-detect repository from git remote
envini versions

# Explicit repository specification
envini versions <owner> <repo>
```

#### Delete Secrets
```bash
# Auto-detect repository from git remote
envini delete                        # Delete latest from development tag
envini delete --tag=production       # Delete latest from production tag
envini delete --version=1           # Delete specific version

# Explicit repository specification
envini delete <owner> <repo>
envini delete <owner> <repo> --tag=production
envini delete <owner> <repo> --version=1
```

#### Help
```bash
envini help                 # Show detailed help and examples
```

## 🛠️ Options

### Common Flags
- `--tag=<value>` - Specify tag for upload/download/delete operations (default: development for latest operations)
- `--version=<value>` - Specify version number or 'latest' (default: latest)

### Examples
```bash
# Tag-based operations
envini upload .env --tag=production
envini download .env.prod --tag=production
envini delete --tag=staging

# Version-specific operations
envini download .env.v1 --version=1
envini delete --version=2

# Combined operations
envini download .env.prod.v2 --version=2 --tag=production
```

## 🔍 Auto-Detection

The CLI can automatically detect your GitHub repository from the current directory's git remote:

```bash
# Works in any git repository directory
cd /path/to/your/git/repo
envini upload .env          # Automatically uses current repo
envini download .env.local  # Downloads from current repo
```

**Supported git remote formats:**
- HTTPS: `https://github.com/owner/repo.git`
- SSH: `git@github.com:owner/repo.git`

## 🏷️ Tag System

Envini uses a tag-based versioning system:

- **Tags**: Logical environments (development, production, staging, etc.)
- **Versions**: Sequential numbers within each tag (1, 2, 3, ...)
- **Default tag**: `development` for upload operations
- **Latest operations**: Use `development` tag when no tag specified

### Tag Examples
```
Repository: myuser/myapp
├── development tag
│   ├── v1 (uploaded yesterday)
│   ├── v2 (uploaded today)
│   └── v3 (latest)
├── production tag
│   ├── v1 (uploaded last week)
│   └── v2 (latest)
└── staging tag
    └── v1 (latest)
```

## 🔐 Security

- **JWT Authentication**: Secure session management
- **AES-256 Encryption**: All secrets encrypted at rest
- **GitHub OAuth**: No password storage required
- **Repository Access Control**: Only access repositories you own
- **Audit Logging**: All operations tracked server-side

## 📁 Configuration

### Environment Variables
The CLI can be configured using environment variables:

```bash
# Set backend URL (default: http://localhost:3000)
export BACKEND_URL=http://your-backend-url:3000
```

### Authentication Storage
The CLI stores authentication data in:
```
CLI/temp/auth.json
```

This file contains your JWT token and should not be shared.

## 🔧 Troubleshooting

### Authentication Issues
```bash
# Clear stored authentication and re-authenticate
rm temp/auth.json
envini auth
```

### Git Detection Issues
If auto-detection fails:
```bash
# Use explicit repository format instead
envini upload owner repo .env
```

### Connection Issues
Ensure the Envini backend services are running:
- AuthService (port 50052)
- SecretOperationService (port 50053)  
- BackendGate (port 3000)

### Session Expired
```bash
# Re-authenticate if you see "Session expired" messages
envini auth
```

## 🚀 Advanced Usage

### Batch Operations
```bash
# Upload multiple environments
envini upload .env.development --tag=development
envini upload .env.staging --tag=staging
envini upload .env.production --tag=production

# Download specific configurations
envini download .env.dev --tag=development
envini download .env.staging --tag=staging
envini download .env.prod --tag=production
```

### Version Management
```bash
# Check what versions exist
envini versions

# Download previous version for rollback
envini download .env.rollback --version=1 --tag=production

# Clean up old versions
envini delete --version=1
envini delete --version=2
```

## 📖 More Information

For complete system documentation, see the main [Envini README](../README.md).
