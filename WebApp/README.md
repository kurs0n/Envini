# Envini WebApp

A modern React-based web application for managing environment variables across repositories. This frontend connects to the BackendGate API to provide a user-friendly interface for the Envini platform.

## Features

- **GitHub Authentication**: Secure OAuth flow using device code authentication
- **Repository Management**: View and manage repositories with environment variables
- **Environment Variables**: Upload, download, and manage environment files for different environments (development, staging, production, testing)
- **Real-time Updates**: Live data from the BackendGate API
- **Responsive Design**: Modern UI built with Tailwind CSS

## Technology Stack

- **React 18**: Modern React with hooks
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool and development server
- **Tailwind CSS**: Utility-first CSS framework
- **Lucide React**: Beautiful icons
- **Axios**: HTTP client for API communication
- **React Router**: Client-side routing

## Getting Started

### Prerequisites

- Node.js 18+ 
- BackendGate service running (default: `http://localhost:3000`)

### Environment Configuration

The WebApp can be configured using environment variables:

```bash
# Set backend URL (default: http://localhost:3000)
VITE_BACKEND_URL=http://your-backend-url:3000
```

Create a `.env` file in the WebApp directory:
```bash
# .env
VITE_BACKEND_URL=http://localhost:3000
```

### Installation

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

3. Open your browser to `http://localhost:5173`

### Building for Production

```bash
npm run build
```

## API Integration

The WebApp connects to the following BackendGate endpoints:

### Authentication
- `POST /auth/github/start` - Start GitHub OAuth flow
- `GET /auth/github/poll` - Poll for authentication token
- `GET /auth/validate` - Validate JWT session
- `POST /auth/logout` - Logout user

### Repositories
- `GET /repos/list` - List user repositories
- `GET /repos/list-with-versions` - List repositories with secret versions

### Secrets
- `POST /secrets/upload/:ownerLogin/:repoName` - Upload environment file
- `GET /secrets/versions/:ownerLogin/:repoName` - List secret versions
- `GET /secrets/download/:ownerLogin/:repoName` - Download environment file
- `GET /secrets/content/:ownerLogin/:repoName` - Get secret content
- `DELETE /secrets/delete/:ownerLogin/:repoName` - Delete secret

## Project Structure

```
src/
├── components/          # Reusable UI components
│   └── Navbar.tsx     # Navigation bar
├── contexts/           # React contexts
│   └── AuthContext.tsx # Authentication state management
├── pages/              # Page components
│   ├── Login.tsx       # GitHub authentication page
│   ├── RepositoryList.tsx # Repository listing page
│   └── RepositoryDetails.tsx # Repository details and environment management
├── services/           # API services
│   └── api.ts         # Axios configuration and API functions
├── App.tsx            # Main app component with routing
└── main.tsx           # Application entry point
```

## Authentication Flow

1. User clicks "Continue with GitHub" on login page
2. BackendGate generates device code and verification URL
3. User visits verification URL and enters code
4. WebApp polls BackendGate for authentication token
5. JWT token is stored in localStorage
6. User is redirected to repositories page

## Environment Management

- **Upload**: Drag and drop or select .env files to upload
- **Download**: Download environment files for specific tags
- **View**: Browse environment variables with show/hide functionality
- **Edit**: Inline editing of environment variables (local changes)
- **Organize**: Group environments by tags (development, staging, production, testing)

## Development

### Adding New Features

1. Create new components in `src/components/`
2. Add new pages in `src/pages/`
3. Extend API services in `src/services/api.ts`
4. Update routing in `App.tsx`

### Styling

The app uses Tailwind CSS for styling. Custom animations are defined in `src/index.css`.

### State Management

Authentication state is managed through React Context (`AuthContext.tsx`). API calls are handled through service functions in `api.ts`.

## Troubleshooting

### Common Issues

1. **Authentication fails**: Ensure BackendGate is running on port 3000
2. **API calls fail**: Check browser console for CORS errors
3. **Build errors**: Ensure all TypeScript types are properly defined

### Debug Mode

Enable debug logging by opening browser console and checking for API request/response logs.

## Contributing

1. Follow the existing code structure and patterns
2. Use TypeScript for all new code
3. Add proper error handling for API calls
4. Test authentication flow thoroughly
5. Ensure responsive design for mobile devices
