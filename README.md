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

### 2. **BackendGate** (NestJS REST API)
- **Purpose**: REST API gateway that communicates with gRPC services
- **Features**:
  - REST endpoints for authentication and secrets management
  - Repository listing via SecretsService
  - JWT token validation and forwarding
  - Secrets upload, download, and version management
  - Tag-based secret retrieval
  - Clean separation between frontend and backend services

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
  - **Security Features**:
    - AES-256 encryption for all secrets
    - Per-secret encryption keys
    - Master key encryption for key management
    - SHA256 checksums for integrity verification
    - Comprehensive audit logging

### 4. **CLI** (Go Client)
- **Purpose**: Command-line interface for users
- **Features**:
  - GitHub authentication flow
  - Repository listing
  - File upload capabilities
  - Interactive loading animations
  - Help system

### 5. **Audit Database** (PostgreSQL)
- **Purpose**: Stores secrets, repositories, and audit logs
- **Features**:
  - Encrypted secrets storage
  - Repository metadata
  - Comprehensive audit trail
  - Version control for secrets
  - Tag-based organization

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

### 4. Generate Protocol Buffers
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
go build -o envini-cli
./envini-cli
```

## 📖 Usage

### CLI Commands

#### Authentication
```bash
# Start GitHub authentication
./envini-cli auth

# This will:
# 1. Display a device code and verification URL
# 2. Show a loading spinner while polling for completion
# 3. Store the JWT token for subsequent requests
```

#### Repository Operations
```bash
# List your GitHub repositories
./envini-cli list

# Upload a file (placeholder for future implementation)
./envini-cli upload <file-path>
```

#### Help
```bash
# Show available commands
./envini-cli help
```

### API Endpoints (BackendGate)

#### Authentication
- `POST /auth/start` - Start GitHub device flow
- `POST /auth/poll` - Poll for authentication completion
- `POST /auth/validate` - Validate JWT session
- `POST /auth/logout` - Logout and clear session

#### Repository Operations
- `GET /repos/list` - List GitHub repositories (requires JWT Bearer token)

#### Secrets Management
- `POST /secrets/upload/:ownerLogin/:repoName` - Upload `.env` file
  - Body: `{ "tag": "production", "envFileContent": "base64_encoded_content" }`
- `GET /secrets/versions/:ownerLogin/:repoName` - List secret versions
- `GET /secrets/download/:ownerLogin/:repoName` - Download secret by version
  - Query: `?version=1` or `?tag=production`
- `GET /secrets/content/:ownerLogin/:repoName` - Get secret content as JSON
  - Query: `?version=1` or `?tag=production`
- `DELETE /secrets/delete/:ownerLogin/:repoName` - Delete secret
  - Query: `?version=1` or `?all=true`

## 🔐 Security Features

### Authentication & Authorization
- **JWT-based Authentication**: Secure session management with JWTs
- **GitHub OAuth Device Flow**: Secure authentication without client secrets
- **PostgreSQL Session Storage**: Persistent session management
- **Repository Access Control**: Verify user has access to specific repositories

### Data Protection
- **AES-256 Encryption**: All secrets are encrypted at rest
- **Per-Secret Keys**: Each secret has its own unique encryption key
- **Master Key Encryption**: Secret keys are encrypted with a master key
- **SHA256 Checksums**: Integrity verification for all secrets
- **Base64 Encoding**: Secure transmission of encrypted data

### Audit & Compliance
- **Comprehensive Audit Logging**: All operations are logged with metadata
- **User Tracking**: Track who performed each operation
- **IP Address Logging**: Record client IP addresses
- **User Agent Logging**: Track client applications
- **Success/Failure Tracking**: Monitor operation outcomes

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

#### Secrets Table
```sql
CREATE TABLE secrets (
  id BIGSERIAL PRIMARY KEY,
  repo_id BIGINT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
  version BIGINT NOT NULL,
  tag VARCHAR(255),
  env_data TEXT NOT NULL, -- Encrypted data
  checksum VARCHAR(64) NOT NULL,
  uploaded_by VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ,
  is_encrypted BOOLEAN DEFAULT FALSE,
  encrypted_key VARCHAR(255), -- Encrypted per-secret key
  UNIQUE(repo_id, version)
);
```

#### Audit Logs Table
```sql
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  operation VARCHAR(50) NOT NULL,
  repo_id BIGINT REFERENCES repositories(id),
  secret_id BIGINT REFERENCES secrets(id),
  user_login VARCHAR(255) NOT NULL,
  ip_address VARCHAR(45),
  user_agent TEXT,
  success BOOLEAN NOT NULL,
  error_message TEXT,
  created_at TIMESTAMPTZ
);
```

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
│   ├── proto/
│   └── main.go
├── BackendGate/                 # NestJS REST API gateway
│   ├── src/
│   │   ├── auth/
│   │   ├── repos/
│   │   ├── secrets/             # NEW: Secrets management
│   │   └── grpc/
│   └── package.json
├── SecretOperationService/       # Go gRPC secrets service
│   ├── internal/
│   │   ├── server.go            # Enhanced with secrets management
│   │   └── database.go          # NEW: Database operations
│   ├── proto/
│   └── main.go
├── CLI/                         # Go command-line client
│   ├── auth/
│   ├── list/
│   ├── upload/
│   └── main.go
├── Database_AuthService/        # Auth database setup
├── Database_AuditService/       # NEW: Audit database setup
├── proto/                       # Protocol buffer definitions
│   ├── auth.proto
│   └── secrets.proto            # Enhanced with new operations
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

**Note**: This system is designed for secure environment variable management across GitHub repositories with enterprise-grade authentication, encryption, and audit capabilities.
