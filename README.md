# Access Tracker

A monorepo for the Access Tracker project, containing backend, frontend, and Kubernetes deployment configurations.

## Structure

```
.
├── backend/          # Go backend API
├── frontend/         # Vue 3 frontend application
├── k8s/              # Kubernetes manifests
└── .github/workflows/ # CI/CD workflows
```

## Quick Start

### Backend

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Tech Stack

- **Backend**: Go
- **Frontend**: Vue 3 + TypeScript + Vite
- **Infrastructure**: Kubernetes
- **CI/CD**: GitHub Actions
