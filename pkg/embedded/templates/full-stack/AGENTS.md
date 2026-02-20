# Full-Stack Application Template

## Overview

This template provides a complete full-stack web application with separated backend and frontend directories, Docker Compose for local development, and a modern architecture.

## Technology Stack

### Backend (`backend/`)
- **Language**: Go 1.24+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL (via Docker)
- **Architecture**: Clean architecture with handlers, services, models, middleware

### Frontend (`frontend/`)
- **Framework**: React/TypeScript (or your preferred framework)
- **Package Manager**: npm/yarn
- **Build Tool**: Vite or Next.js

## Directory Structure

```
.
├── backend/
│   ├── cmd/server/         # Application entry point
│   ├── internal/
│   │   ├── handlers/       # HTTP request handlers
│   │   ├── services/       # Business logic
│   │   ├── models/         # Data models
│   │   ├── database/       # Database connections
│   │   └── middleware/     # HTTP middleware
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── pages/          # Page components
│   │   ├── hooks/          # Custom React hooks
│   │   └── services/       # API client services
│   └── package.json
├── docker-compose.yml      # Local dev environment
└── README.md
```

## Development Commands

### Start Local Environment
```bash
docker-compose up -d          # Start PostgreSQL and dependencies
cd backend && go run ./cmd/server   # Start backend
cd frontend && npm run dev    # Start frontend
```

### Testing
```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm test
```

### Building
```bash
# Backend binary
cd backend && go build -o server ./cmd/server

# Frontend production build
cd frontend && npm run build
```

## Code Guidelines

- Keep handlers thin; business logic belongs in services
- Use dependency injection for testability
- Frontend components should be small and focused
- API contracts documented in backend/internal/handlers

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
