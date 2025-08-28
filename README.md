# Envini - Secure Environment Management System

A comprehensive system for managing environment variables and secrets across GitHub repositories with secure authentication, encryption, and CLI tools.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      CLI        │    │   BackendGate   │    │   AuthService   │
│   (Go Client)   │◄──►│   (NestJS API)  │◄──►│     (gRPC)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │SecretOperation  │    │   PostgreSQL    │
                       │Service (gRPC)   │    │   (Sessions)    │
                       └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   PostgreSQL    │
                       │   (Audit DB)    │
                       └─────────────────┘
```

## 🚀 Components

### 1. **AuthService** (Go gRPC Server)
- **Purpose**: Handles GitHub OAuth device flow authentication
- **Features**:
  - GitHub device code authentication flow
  - JWT-based session management
  - PostgreSQL-backed session storage
  - Token refresh and validation
  - Secure logout functionality
  - **NEW**: Username extraction from JWT tokens
  - **NEW**: Session-based username retrieval

### 2. **BackendGate** (NestJS REST API)
- **Purpose**: REST API gateway that communicates with gRPC services
- **Features**:
  - REST endpoints for authentication and secrets management
  - Repository listing via SecretsService
  - JWT token validation and forwarding
  - Secrets upload, download, and version management
  - Tag-based secret retrieval
  - Clean separation between frontend and backend services
  - **NEW**: Enhanced secrets management with tag-specific versioning
  - **NEW**: Repository listing with version information
  - **NEW**: Username propagation through gRPC metadata

### 3. **SecretOperationService** (Go gRPC Server)
- **Purpose**: Handles GitHub repository operations and secrets management
- **Features**:
  - List GitHub repositories
  - Secure access token handling
  - Repository metadata retrieval
  - **Secrets Management**:
    - Upload `.env` files with versioning
    - Download secrets by version or tag
    - List secret versions with metadata
    - Delete specific versions or all versions
    - **NEW**: Tag-specific versioning (separate version sequences per tag)
    - **NEW**: Repository listing with all secret versions
  - **Security Features**:
    - AES-256 encryption for all secrets
    - Per-secret encryption keys
    - Master key encryption for key management
    - SHA256 checksums for integrity verification
    - Comprehensive audit logging
    - **NEW**: Username tracking in audit logs
    - **NEW**: Service name and request ID tracking

### 4. **CLI** (Go Client)
- **Purpose**: Command-line interface for users
- **Features**:
  - GitHub authentication flow
  - Repository listing
  - File upload capabilities
  - Interactive loading animations
  - Help system
  - **NEW**: Git repository auto-detection
  - **NEW**: Tag-specific secret management
  - **NEW**: Smart defaults (development tag, latest version)
  - **NEW**: Flag-based command options
  - **NEW**: Comprehensive secret operations (upload, download, delete, versions)

### 5. **Audit Database** (PostgreSQL)
- **Purpose**: Stores secrets, repositories, and audit logs
- **Features**:
  - Encrypted secrets storage
  - Repository metadata
  - Comprehensive audit trail
  - Version control for secrets
  - Tag-based organization
  - **NEW**: Username tracking in audit logs
  - **NEW**: Service name and request ID tracking
  - **NEW**: Tag-specific version constraints

## 📋 Prerequisites

- **Go** 1.21+
- **Node.js** 18+
- **PostgreSQL** 14+
- **Docker** (for PostgreSQL containers)
- **GitHub OAuth App** credentials

## 🛠️ Installation & Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd Envini
```

### 2. Environment Variables

Create `.env` files for each component (these files are gitignored, so you need to create them manually):

#### AuthService (.env)
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=envini
DB_PASSWORD=envini
DB_NAME=envini
DB_SSL_MODE=disable
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
JWT_SECRET=your_jwt_secret_key
```

#### BackendGate (.env)
```env
AUTH_SERVICE_URL=localhost:50052
SECRETS_SERVICE_URL=localhost:50053
PORT=3000
```

#### SecretOperationService (.env)
```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=envini
DB_PASSWORD=envini
DB_NAME=envini_audit
GITHUB_API_URL=https://api.github.com
GRPC_PORT=50053
# Master encryption key (32 bytes base64 encoded)
MASTER_ENCRYPTION_KEY=your_master_encryption_key_here
```

### 3. Database Setup

#### Auth Database (Port 5432)
```bash
docker run -d \
  --name postgres-auth \
  -e POSTGRES_PASSWORD=envini \
  -e POSTGRES_USER=envini \
  -e POSTGRES_DB=envini \
  -p 5432:5432 \
  postgres:14
```

#### Audit Database (Port 5433)
```bash
cd Database_AuditService
make run
```

Or manually:
```bash
docker run -d \
  --name postgres-audit \
  -e POSTGRES_PASSWORD=envini \
  -e POSTGRES_USER=envini \
  -e POSTGRES_DB=envini_audit \
  -p 5433:5433 \
  postgres:14
```

### 4. Database Migrations

Run the following SQL commands to set up the new schema:

#### Update Secrets Table Constraint
```sql
-- Connect to audit database
docker exec -it postgres-audit psql -h localhost -p 5433 -U envini -d envini_audit

-- Drop old constraint and create new tag-specific constraint
DROP INDEX IF EXISTS idx_repo_version;
CREATE UNIQUE INDEX idx_repo_tag_version ON secrets (repo_id, tag, version);
```

#### Add Username Column to Audit Logs
```sql
-- Add username column to audit_logs
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS username VARCHAR(255) NOT NULL DEFAULT 'unknown';
```

### 5. Generate Protocol Buffers
```bash
make proto
```

## 🚀 Running the Services

### 1. Start AuthService
```bash
cd AuthService
go mod tidy
go run main.go
```

### 2. Start SecretOperationService
```bash
cd SecretOperationService
go mod tidy
go run main.go
```

### 3. Start BackendGate
```bash
cd BackendGate
npm install
npm run start:dev
```

### 4. Build and Run CLI
```bash
cd CLI
go mod tidy
go build -o envini
./envini
```

## 📖 Usage

### CLI Commands

#### Authentication
```bash
# Start GitHub authentication
envini auth

# This will:
# 1. Display a device code and verification URL
# 2. Show a loading spinner while polling for completion
# 3. Store the JWT token for subsequent requests
```

#### Repository Operations
```bash
# List your GitHub repositories
envini repos

# List repositories with all secret versions
envini repos-with-versions
```

#### Secret Management (NEW!)

##### Upload Secrets
```bash
# Auto-detect repository from git (recommended)
envini upload .env                    # Uses "development" tag by default
envini upload .env --tag=production   # Upload with custom tag
envini upload .env --tag=staging      # Upload with staging tag

# Explicit repository
envini upload kurs0n 8080-emulator .env
envini upload kurs0n 8080-emulator .env --tag=production
```

##### Download Secrets
```bash
# Auto-detect repository from git
envini download .env.downloaded       # Downloads latest from development tag
envini download .env.downloaded --tag=production  # Downloads latest from production tag
envini download .env.downloaded --version=1  # Downloads specific version
envini download .env.downloaded --version=2 --tag=production  # Downloads version 2 from production tag

# Explicit repository
envini download kurs0n 8080-emulator .env.downloaded
envini download kurs0n 8080-emulator .env.downloaded --tag=production
envini download kurs0n 8080-emulator .env.downloaded --version=1
```

##### List Secret Versions
```bash
# Auto-detect repository from git
envini versions

# Explicit repository
envini versions kurs0n 8080-emulator
```

##### Delete Secrets
```bash
# Auto-detect repository from git
envini delete                        # Deletes latest from development tag
envini delete --tag=production       # Deletes latest from production tag
envini delete --version=1           # Deletes specific version

# Explicit repository
envini delete kurs0n 8080-emulator
envini delete kurs0n 8080-emulator --tag=production
envini delete kurs0n 8080-emulator --version=1
```

#### Help
```bash
# Show available commands
envini help
```

### API Endpoints (BackendGate)

#### Authentication
- `POST /auth/start` - Start GitHub device flow
- `POST /auth/poll` - Poll for authentication completion
- `POST /auth/validate` - Validate JWT session
- `POST /auth/logout` - Logout and clear session

#### Repository Operations
- `GET /repos/list` - List GitHub repositories (requires JWT Bearer token)
- `GET /repos/list-with-versions` - List repositories with all secret versions

#### Secrets Management
- `POST /secrets/upload/:ownerLogin/:repoName` - Upload `.env` file
  - Body: `{ "tag": "production", "envFileContent": "base64_encoded_content" }`
- `GET /secrets/versions/:ownerLogin/:repoName` - List secret versions
- `GET /secrets/download/:ownerLogin/:repoName` - Download secret by version or tag
  - Query: `?version=1`, `?tag=production`, or `?version=2&tag=production`
- `GET /secrets/content/:ownerLogin/:repoName` - Get secret content as JSON
  - Query: `?version=1`, `?tag=production`, or `?version=2&tag=production`
- `DELETE /secrets/delete/:ownerLogin/:repoName` - Delete secret
  - Query: `?version=1`, `?tag=production`, or `?version=2&tag=production`

## 🔐 Security Features

### Authentication & Authorization
- **JWT-based Authentication**: Secure session management with JWTs
- **GitHub OAuth Device Flow**: Secure authentication without client secrets
- **PostgreSQL Session Storage**: Persistent session management
- **Repository Access Control**: Verify user has access to specific repositories
- **NEW**: Username tracking in audit logs

### Data Protection
- **AES-256 Encryption**: All secrets are encrypted at rest
- **Per-Secret Keys**: Each secret has its own unique encryption key
- **Master Key Encryption**: Secret keys are encrypted with a master key
- **SHA256 Checksums**: Integrity verification for all secrets
- **Base64 Encoding**: Secure transmission of encrypted data
- **NEW**: Tag-specific versioning for better organization

### Audit & Compliance
- **Comprehensive Audit Logging**: All operations are logged with metadata
- **User Tracking**: Track who performed each operation
- **Service Tracking**: Track which service performed operations
- **Request ID Tracking**: Unique request identifiers for tracing
- **Success/Failure Tracking**: Monitor operation outcomes
- **NEW**: Username extraction from JWT tokens
- **NEW**: Service name and request ID in audit logs

## 📊 Database Schema

### Audit Database (envini_audit)

#### Repositories Table
```sql
CREATE TABLE repositories (
  id BIGSERIAL PRIMARY KEY,
  owner_login VARCHAR(255) NOT NULL,
  repo_name VARCHAR(255) NOT NULL,
  repo_id BIGINT NOT NULL,
  full_name VARCHAR(500) NOT NULL,
  html_url VARCHAR(1000) NOT NULL,
  description TEXT,
  is_private BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  UNIQUE(owner_login, repo_name)
);
```

#### Secrets Table (UPDATED)
```sql
CREATE TABLE secrets (
  id BIGSERIAL PRIMARY KEY,
  repo_id BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
  tag VARCHAR(255) NOT NULL,
  version BIGINT NOT NULL,
  env_data TEXT NOT NULL, -- Encrypted data
  checksum VARCHAR(64) NOT NULL,
  uploaded_by VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ,
  encrypted_key VARCHAR(255), -- Encrypted per-secret key
  UNIQUE(repo_id, tag, version) -- NEW: Tag-specific versioning
);
```

#### Audit Logs Table (UPDATED)
```sql
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  operation VARCHAR(50) NOT NULL,
  repo_id BIGINT REFERENCES repositories(id),
  secret_id BIGINT REFERENCES secrets(id),
  username VARCHAR(255) NOT NULL, -- NEW: Username tracking
  service_name VARCHAR(100) NOT NULL, -- NEW: Service tracking
  request_id VARCHAR(255), -- NEW: Request ID tracking
  success BOOLEAN NOT NULL,
  error_message TEXT,
  created_at TIMESTAMPTZ
);
```

## 🆕 New Features

### Tag-Specific Versioning
- **Separate version sequences** for each tag (development, production, staging, etc.)
- **Example**: v1(development), v2(development), v1(production), v2(production)
- **Smart defaults**: "development" tag by default, "latest" version by default

### Git Repository Auto-Detection
- **Automatic repository detection** from git remote
- **Supports both HTTPS and SSH** git remote formats
- **Fallback to explicit repository** if git detection fails
- **Visual feedback** showing detected repository

### Enhanced CLI Commands
- **Simplified command names**: `upload`, `download`, `delete`, `versions`
- **Flag-based options**: `--tag=production`, `--version=1`
- **Smart defaults**: Development tag, latest version
- **Comprehensive help**: Detailed usage examples

### Improved Audit Logging
- **Username tracking**: Extract and log usernames from JWT tokens
- **Service tracking**: Track which service performed operations
- **Request ID tracking**: Unique identifiers for request tracing
- **Cleaner output**: Removed redundant information from CLI output

### Security Enhancements
- **Removed redundant encryption flag**: Encryption status determined by `EncryptedKey` field
- **Enhanced constraint**: Tag-specific unique constraints
- **Better error handling**: More descriptive error messages

## 🔧 Development

### Protocol Buffer Generation
```bash
# Generate Go code from proto files
make proto
```

### Database Management
```bash
# Start PostgreSQL containers
make postgres-start

# Stop PostgreSQL containers
make postgres-stop

# Remove PostgreSQL containers
make postgres-clean
```

### Testing
```bash
# Test BackendGate
cd BackendGate
npm test

# Test Go services
cd AuthService
go test ./...

cd ../SecretOperationService
go test ./...
```

## 📁 Project Structure

```
Envini/
├── AuthService/                 # Go gRPC authentication service
│   ├── internal/
│   │   ├── server.go           # Enhanced with username extraction
│   │   └── session.go          # Session management
│   ├── proto/
│   └── main.go
├── BackendGate/                 # NestJS REST API gateway
│   ├── src/
│   │   ├── auth/
│   │   ├── repos/              # Enhanced with version listing
│   │   ├── secrets/            # Complete secrets management
│   │   └── grpc/
│   └── package.json
├── SecretOperationService/       # Go gRPC secrets service
│   ├── internal/
│   │   ├── server.go           # Enhanced with tag-specific versioning
│   │   └── database.go         # Enhanced with new constraints
│   ├── proto/
│   └── main.go
├── CLI/                         # Go command-line client
│   ├── auth/
│   ├── list/                   # Enhanced with version listing
│   ├── secrets/                # NEW: Complete secrets management
│   ├── upload/                 # Legacy upload
│   └── main.go                 # Enhanced with git detection
├── Database_AuthService/        # Auth database setup
├── Database_AuditService/       # Audit database setup
├── proto/                       # Protocol buffer definitions
│   ├── auth.proto              # Enhanced with username methods
│   └── secrets.proto           # Enhanced with version listing
├── Makefile                     # Build and deployment scripts
└── README.md
```

## 🔧 Troubleshooting

### Connection Issues
If you see connection errors like `ECONNREFUSED 127.0.0.1:5000`, ensure:

1. **AuthService is running** on port 50052:
   ```bash
   cd AuthService
   go run main.go
   ```

2. **SecretOperationService is running** on port 50053:
   ```bash
   cd SecretOperationService
   go run main.go
   ```

3. **BackendGate environment** is correctly configured:
   ```bash
   # Create BackendGate/.env file
   echo "AUTH_SERVICE_URL=localhost:50052" > BackendGate/.env
   echo "SECRETS_SERVICE_URL=localhost:50053" >> BackendGate/.env
   echo "PORT=3000" >> BackendGate/.env
   ```

4. **Databases are running** with correct credentials:
   ```bash
   # Check if PostgreSQL containers are running
   docker ps | grep postgres
   ```

### Service Startup Order
Start services in this order:
1. PostgreSQL databases (auth and audit)
2. AuthService (port 50052)
3. SecretOperationService (port 50053)
4. BackendGate (port 3000)
5. CLI

### Encryption Issues
If you encounter encryption-related errors:

1. **Check Master Key**: Ensure `MASTER_ENCRYPTION_KEY` is set in SecretOperationService `.env`
2. **Generate New Key**: Use `openssl rand -base64 32` to generate a new master key
3. **Database Migration**: Restart SecretOperationService to apply schema changes

### Database Constraint Issues
If you see duplicate key constraint errors:

1. **Run Migration**: Execute the database migration scripts
2. **Check Constraints**: Verify the new tag-specific constraints are in place
3. **Restart Services**: Restart SecretOperationService after migration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For issues and questions:
- Check the existing issues
- Create a new issue with detailed information
- Contact the development team

---

**Note**: This system is designed for secure environment variable management across GitHub repositories with enterprise-grade authentication, encryption, and audit capabilities. The new tag-specific versioning system provides better organization and the git auto-detection feature makes the CLI much more user-friendly.
