# Access Tracker Architecture

## System Overview

Access Tracker is a system for tracking and managing resource access permissions and logs. The system follows a client-server architecture with a Vue.js frontend, Go backend, GraphQL API, and PostgreSQL database.

## Architecture Diagram

```
┌─────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌────────────┐
│ Client  │────▶│   Vue.js    │────▶│  GraphQL    │────▶│    Go       │────▶│ PostgreSQL │
│ Browser │     │  Frontend   │     │   API       │     │   Backend   │     │  Database  │
└─────────┘     └─────────────┘     └─────────────┘     └─────────────┘     └────────────┘
     │                                                                              │
     │                                                                              │
     ▼                                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                      Kubernetes Cluster                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │
│  │   Ingress   │  │   Frontend  │  │   Backend   │  │     DB      │  │  Helm Charts    │  │
│  │  + TLS      │  │   Service   │  │   Service   │  │   StatefulSet│  │  - access-tracker│  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │  - postgresql    │  │
│                                                                    └─────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
                                            │
                                            ▼
                                    ┌─────────────┐
                                    │   Quay.io   │
                                    │  Container  │
                                    │   Registry  │
                                    └─────────────┘
```

## Data Flow

### Access Request Creation Flow

1. **Client Request**: User fills out the access request form in the Vue.js frontend
2. **GraphQL Mutation**: Frontend sends a `createAccessRequest` mutation via GraphQL
3. **Backend Processing**:
   - GraphQL server receives the request
   - Validates the input schema
   - Writes to PostgreSQL database
   - Returns the created access request
4. **Response**: Frontend updates the UI with the new request status

### Access Request Query Flow

1. **Client Request**: User views the access request list
2. **GraphQL Query**: Frontend sends an `accessRequests` query with optional filters
3. **Backend Processing**:
   - GraphQL server receives the query
   - Applies filters (status, requester)
   - Reads from PostgreSQL database
   - Returns filtered results
4. **Response**: Frontend displays the list

### Status Update Flow

1. **Client Request**: Admin clicks approve/deny on an access request
2. **GraphQL Mutation**: Frontend sends `updateAccessRequestStatus` mutation
3. **Backend Processing**:
   - GraphQL server validates the status transition
   - Updates PostgreSQL record
   - Returns updated access request
4. **Response**: Frontend reflects the new status

## Tool Selection Rationale

### Go (Backend)

- **Performance**: Go provides excellent performance for API workloads with minimal memory footprint
- **Concurrency**: Built-in goroutines and channels for handling multiple simultaneous connections
- **Tooling**: Excellent standard library and mature ecosystem for web services
- **Deployment**: Single binary deployment simplifies Kubernetes operations

### Vue 3 (Frontend)

- **Developer Experience**: `<script setup>` syntax provides clean, readable component code
- **Reactivity**: Vue's reactivity system handles UI state changes efficiently
- **TypeScript Support**: First-class TypeScript integration improves code quality
- **Ecosystem**: Mature router, state management, and build tooling (Vite)

### GraphQL (API)

- **Flexible Queries**: Clients can request exactly the data they need
- **Type Safety**: Schema-first development catches errors early
- **Single Endpoint**: Simplifies frontend-backend communication
- **Tooling**: GraphQL Playground/Altair provides excellent developer experience

### PostgreSQL (Database)

- **Reliability**: Proven ACID compliance and data integrity
- **JSON Support**: Can handle semi-structured data alongside relational
- **Bitnami Helm Chart**: Easy Kubernetes deployment with the Bitnami PostgreSQL chart
- **Performance**: Excellent read/write performance for moderate workloads

### Kubernetes + Helm (Infrastructure)

- **Container Orchestration**: Handles deployment, scaling, and health management
- **Helm Charts**: Reproducible deployments with versioned templates
- **Horizontal Pod Autoscaler**: Supports the 10x traffic scaling requirement
- **Ingress + Cert-Manager**: Automatic TLS termination and certificate management

### Quay.io (Container Registry)

- **Security**: Image scanning and vulnerability detection
- **Integration**: Works well with GitHub Actions for CI/CD
- **Namespace**: `quay.io/mathianasj` provides clear organization

## Deployment Architecture

### Container Images

| Component | Image | Tag Strategy |
|-----------|-------|--------------|
| Backend | `quay.io/mathianasj/backend` | `{branch}` on push |
| Frontend | `quay.io/mathianasj/frontend` | `{branch}` on push |

### Helm Release

The `access-tracker` Helm chart deploys:
- Frontend deployment (replica count configurable)
- Backend deployment (replica count configurable, supports HPA)
- PostgreSQL database (via Bitnami subchart)
- Services for both frontend and backend
- Ingress with TLS (cert-manager)

### Scaling

- **Backend**: Configurable replica count with HPA support
- **Frontend**: Configurable replica count
- **Database**: PostgreSQL primary with potential for read replicas
- **Resource Limits**: CPU and memory limits defined in values files for HPA

## Project Structure

```
access-tracker/
├── backend/                  # Go REST API server
│   ├── cmd/server/          # Main entrypoint
│   └── internal/
│       ├── db/              # Database connection and migrations
│       ├── graphql/          # GraphQL schema and resolvers
│       ├── middleware/       # HTTP middleware (logging, etc.)
│       └── models/           # Data models
├── frontend/                 # Vue 3 SPA
│   └── src/
│       ├── components/       # Vue components
│       ├── router.ts         # Vue Router config
│       └── main.ts           # Frontend entrypoint
├── k8s/                      # Kubernetes manifests
│   └── access-tracker/      # Helm chart
└── docs/                     # Documentation
```
