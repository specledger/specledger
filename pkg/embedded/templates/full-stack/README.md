# Full-Stack Application Template

Go backend + TypeScript/React frontend project structure.

## Structure

```
backend/
├── cmd/server/main.go    # HTTP server entrypoint
└── internal/
    ├── handlers/         # HTTP request handlers
    ├── services/         # Business logic
    ├── models/           # Data models
    ├── middleware/       # HTTP middleware
    └── database/         # Database layer

frontend/
├── src/
│   ├── components/       # React components
│   ├── pages/            # Page components
│   ├── services/         # API client services
│   └── hooks/            # Custom React hooks
└── public/               # Static assets
```

## Technologies

- **Backend**: Go + Gin (web framework) + GORM (ORM)
- **Frontend**: React 18+ + TypeScript + Vite + TanStack Query
- **Development**: Docker Compose for local services

## Getting Started

1. Backend: `cd backend && go run cmd/server/main.go`
2. Frontend: `cd frontend && npm install && npm run dev`
