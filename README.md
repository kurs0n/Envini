# Envini - Secure Environment Management System

A comprehensive system for managing environment variables and secrets across GitHub repositories with secure authentication and CLI tools.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      CLI        │    │   BackendGate   │    │   AuthService   │
│   (Go Client)   │◄──►│   (NestJS API)  │◄──►│   (gRPC Server) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │SecretOperation  │    │   PostgreSQL    │
                       │Service (gRPC)   │    │   (Sessions)    │
                       └─────────────────┘    └─────────────────┘
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
  - REST endpoints for authentication
  - Repository listing via SecretsService
  - JWT token validation and forwarding
  - Clean separation between frontend and backend services

### 3. **SecretOperationService** (Go gRPC Server)
- **Purpose**: Handles GitHub repository operations
- **Features**:
  - List GitHub repositories
  - Secure access token handling
  - Repository metadata retrieval

### 4. **CLI** (Go Client)
- **Purpose**: Command-line interface for users
- **Features**:
  - GitHub authentication flow
  - Repository listing
  - File upload capabilities
  - Interactive loading animations
  - Help system

## 📋 Prerequisites

- **Go** 1.21+
- **Node.js** 18+
- **PostgreSQL** 14+
- **Docker** (for PostgreSQL container)
- **GitHub OAuth App** credentials

## 🛠️ Installation & Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd Envini
```

### 2. Environment Variables

Create `.env` files for each component:

#### AuthService (.env)
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=authservice
DB_SSL_MODE=disable
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
JWT_SECRET=your_jwt_secret_key
```

#### BackendGate (.env)
```env
AUTH_SERVICE_URL=localhost:50051
SECRETS_SERVICE_URL=localhost:50052
PORT=3000
```

#### SecretOperationService (.env)
```env
GITHUB_API_URL=https://api.github.com
```

### 3. Database Setup

Start PostgreSQL using Docker:
```bash
make postgres-start
```

Or manually:
```bash
docker run -d \
  --name postgres-auth \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=authservice \
  -p 5432:5432 \
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

## 🔧 Development

### Protocol Buffer Generation
```bash
# Generate Go code from proto files
make proto
```

### Database Management
```bash
# Start PostgreSQL container
make postgres-start

# Stop PostgreSQL container
make postgres-stop

# Remove PostgreSQL container
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

## 🔐 Security Features

- **JWT-based Authentication**: Secure session management with JWTs
- **GitHub OAuth Device Flow**: Secure authentication without client secrets
- **PostgreSQL Session Storage**: Persistent session management
- **gRPC Communication**: Type-safe inter-service communication
- **Token Refresh**: Automatic token renewal
- **Secure Logout**: Proper session cleanup

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
│   │   └── grpc/
│   └── package.json
├── SecretOperationService/       # Go gRPC secrets service
│   ├── internal/
│   ├── proto/
│   └── main.go
├── CLI/                         # Go command-line client
│   ├── auth/
│   ├── list/
│   ├── upload/
│   └── main.go
├── proto/                       # Protocol buffer definitions
│   ├── auth.proto
│   └── secrets.proto
├── Database_AuthService/        # Database setup scripts
├── Makefile                     # Build and deployment scripts
└── README.md
```

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

**Note**: This system is designed for secure environment variable management across GitHub repositories with enterprise-grade authentication and session management.
