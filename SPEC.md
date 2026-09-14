# Access Tracker - Project Specification

## Overview

Access Tracker is a system for tracking and managing resource access permissions and logs.

## Architecture

### Monorepo Structure

```
access-tracker/
├── backend/          # Go REST API server
├── frontend/         # Vue 3 SPA frontend
├── k8s/              # Kubernetes deployment manifests
└── .github/workflows/ # GitHub Actions CI/CD pipelines
```

### Backend

- **Language**: Go 1.24+
- **Framework**: Clean Architecture / Standard Library
- **Database**: TBD

### Frontend

- **Framework**: Vue 3 with `<script setup>`
- **Language**: TypeScript
- **Build Tool**: Vite
- **Styling**: CSS

## Infrastructure

### Kubernetes

- Deployment manifests for backend and frontend
- Service and Ingress configurations
- ConfigMaps and Secrets management

### CI/CD

- GitHub Actions workflows for:
  - Backend tests and build
  - Frontend tests and build
  - Deployment automation

## Development Guidelines

### Git Conventions

- Branch naming: `feature/`, `fix/`, `refactor/`
- Commit messages: Conventional Commits
- PR requirements: Tests pass, code reviewed

### Code Standards

- Backend: Go fmt, golint
- Frontend: ESLint, TypeScript strict mode
