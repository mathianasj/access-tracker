# Access Tracker Architecture

## System Overview

Access Tracker is a system for tracking and managing resource access permissions and logs. The system follows a client-server architecture with:

- **Vue 3 Frontend** - Single Page Application (SPA)
- **Go Backend** - REST API with GraphQL
- **PostgreSQL** - Primary database
- **Kubernetes** - Container orchestration
- **Quay.io** - Container registry

## Component Diagram

```mermaid
graph TB
    subgraph "External"
        User["👤 User Browser"]
    end

    subgraph "Kubernetes Cluster"
        subgraph "access-tracker namespace"
            Ingress["🌐 Ingress<br/>nginx-ingress"]
            Frontend["📦 Frontend Pod<br/>Vue 3 SPA"]
            Backend["⚙️ Backend Pod<br/>Go API Server"]
            DB["🗄️ PostgreSQL<br/>Bitnami Chart"]
        end

        subgraph "Monitoring"
            Prometheus["📊 Prometheus"]
            Grafana["📈 Grafana"]
        end
    end

    subgraph "External Services"
        Quay["📦 Quay.io<br/>Container Registry"]
        GitHub["🐙 GitHub<br/>CI/CD"]
    end

    User -->|HTTPS| Ingress
    Ingress -->|/api| Backend
    Ingress -->|/*| Frontend
    Backend -->|GraphQL| DB
    Frontend -->|HTTP| Backend
    GitHub -->|Push Images| Quay
    Prometheus -->|Scrape| Backend
    Prometheus -->|Scrape| Frontend
```

## Request Flow Diagram

```mermaid
sequenceDiagram
    participant User as 👤 User
    participant Browser as 🌐 Browser
    participant Ingress as 🌐 Ingress
    participant Frontend as 📦 Frontend
    participant Backend as ⚙️ Backend
    participant DB as 🗄️ PostgreSQL

    User->>Browser: Opens application
    Browser->>Ingress: GET /
    Ingress->>Frontend: Route to frontend pod
    Frontend->>Browser: Returns SPA
    Browser->>Frontend: User fills request form
    Frontend->>Backend: POST /graphql mutation
    Backend->>DB: INSERT access_request
    DB-->>Backend: Return created record
    Backend-->>Frontend: GraphQL response
    Frontend-->>Browser: Update UI
    Browser-->>User: Show confirmation
```

## CI/CD Pipeline

```mermaid
flowchart LR
    subgraph "GitHub"
        Push["🐙 Push to main"]
        CI["✅ CI Workflows"]
        CD["🚀 Deploy Workflow"]
    end

    subgraph "Build"
        BackendImage["📦 Build Backend<br/>Dockerfile"]
        FrontendImage["📦 Build Frontend<br/>Dockerfile"]
    end

    subgraph "Registry"
        Quay["📦 Quay.io<br/>Container Registry"]
    end

    subgraph "Kubernetes"
        Cluster["☸️ K8s Cluster"]
        Helm["Helm Upgrade"]
    end

    Push --> CI
    CI --> BackendImage
    CI --> FrontendImage
    BackendImage --> Quay
    FrontendImage --> Quay
    Push --> CD
    CD --> Helm
    Helm --> Cluster
```

## Data Model

```mermaid
erDiagram
    USERS {
        uuid id PK
        string username
        string password_hash
        timestamptz created_at
        timestamptz updated_at
    }

    ACCESS_REQUESTS {
        uuid id PK
        uuid request_id
        string requester
        string system_resource
        string access_level
        string justification
        string status
        timestamptz created_at
        timestamptz updated_at
    }

    AUDIT_LOGS {
        uuid id PK
        uuid access_request_id FK
        string action
        string old_value
        string new_value
        string changed_by
        timestamptz changed_at
    }

    USERS ||--o{ ACCESS_REQUESTS : "creates"
    ACCESS_REQUESTS ||--o{ AUDIT_LOGS : "has"
```

## Frontend Architecture

```mermaid
graph TD
    subgraph "Frontend Components"
        App["📱 App.vue"]
        Router["🔀 Vue Router"]
        Store["📋 Pinia Store"]
        Apollo["🔗 Apollo Client"]
    end

    subgraph "Pages"
        Home["🏠 Home/Request List"]
        Form["📝 Request Form"]
    end

    subgraph "Components"
        RequestList["📋 RequestList"]
        StatusBadge["🏷️ StatusBadge"]
        AccessRequestForm["📄 AccessRequestForm"]
    end

    App --> Router
    Router --> Home
    Router --> Form
    Home --> RequestList
    Home --> StatusBadge
    Form --> AccessRequestForm
    RequestList --> Store
    AccessRequestForm --> Store
    Store --> Apollo
    Apollo -->|GraphQL| Backend["⚙️ Backend API"]
```

## Backend Architecture

```mermaid
graph TD
    subgraph "Go Backend"
        Main["🚀 main.go"]
        Router["🛣️ Chi Router"]
        Middleware["🔧 Middleware"]
        Auth["🔐 Auth Middleware"]
        RequestID["🆔 Request ID"]
        Logger["📝 Request Logger"]
    end

    subgraph "GraphQL Layer"
        Handler["📡 GraphQL Handler"]
        Schema["📋 Schema"]
        Resolver["⚙️ Resolver"]
        Query["📖 Query Resolvers"]
        Mutation["✏️ Mutation Resolvers"]
    end

    subgraph "Data Layer"
        DB["🗄️ Database"]
        Pool["连接池 Pool"]
    end

    Main --> Router
    Router --> Middleware
    Middleware --> Auth
    Middleware --> RequestID
    Middleware --> Logger
    Router --> Handler
    Handler --> Schema
    Handler --> Resolver
    Resolver --> Query
    Resolver --> Mutation
    Mutation --> DB
    Query --> DB
    DB --> Pool
```

## Kubernetes Deployment

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "access-tracker Namespace"
            subgraph "Services"
                Ingress["🌐 Ingress<br/>:80/:443"]
                SvcFrontend["⚙️ Service: Frontend<br/>ClusterIP :80"]
                SvcBackend["⚙️ Service: Backend<br/>ClusterIP :8080"]
                SvcDB["⚙️ Service: PostgreSQL<br/>ClusterIP :5432"]
            end

            subgraph "Deployments"
                DeployFrontend["📦 Frontend<br/>replicas: 2"]
                DeployBackend["📦 Backend<br/>replicas: 2"]
                DeployDB["📦 PostgreSQL<br/>StatefulSet"]
            end

            subgraph "Config"
                ConfigMap["📋 ConfigMap<br/>DB URL, GraphQL endpoint"]
                Secret["🔒 Secret<br/>DB credentials"]
            end

            subgraph "Autoscaling"
                HPAFrontend["📈 HPA Frontend<br/>min: 2, max: 10"]
                HPABackend["📈 HPA Backend<br/>min: 2, max: 10"]
            end

            subgraph "Monitoring"
                SMBackend["📊 ServiceMonitor<br/>Backend"]
                SMFrontend["📊 ServiceMonitor<br/>Frontend"]
            end
        end
    end

    Ingress --> SvcFrontend
    Ingress --> SvcBackend
    SvcFrontend --> DeployFrontend
    SvcBackend --> DeployBackend
    SvcDB --> DeployDB
    DeployBackend --> ConfigMap
    DeployBackend --> Secret
    DeployBackend --> DeployDB
    HPAFrontend -.-> DeployFrontend
    HPABackend -.-> DeployBackend
    SMBackend -.-> SvcBackend
    SMFrontend -.-> SvcFrontend
```

## Environment Configuration

```mermaid
graph LR
    subgraph "Development"
        DevVals["values-dev.yaml"]
        DevDB["ephemeral PostgreSQL"]
        DevTLS["TLS disabled"]
        DevReplica["replicas: 1"]
    end

    subgraph "Production"
        ProdVals["values-prod.yaml"]
        ProdDB["persistent PostgreSQL<br/>10Gi PVC"]
        ProdTLS["TLS enabled<br/>Let's Encrypt"]
        ProdReplica["replicas: 2+"]
        ProdHPA["HPA enabled"]
    end

    DevVals -.->|override| BaseVals["values.yaml"]
    ProdVals -.->|override| BaseVals
```

## Tool Selection Rationale

### Go (Backend)

| Reason | Benefit |
|--------|---------|
| Performance | Excellent for API workloads with minimal memory footprint |
| Concurrency | Built-in goroutines for handling multiple simultaneous connections |
| Deployment | Single binary deployment simplifies Kubernetes operations |
| Tooling | Mature ecosystem for web services |

### Vue 3 (Frontend)

| Reason | Benefit |
|--------|---------|
| Developer Experience | Clean `<script setup>` syntax |
| Reactivity | Efficient UI state management |
| TypeScript | First-class integration improves code quality |
| Ecosystem | Mature router, state management, build tooling (Vite) |

### GraphQL (API)

| Reason | Benefit |
|--------|---------|
| Flexible Queries | Clients request exactly the data they need |
| Type Safety | Schema-first development catches errors early |
| Single Endpoint | Simplifies frontend-backend communication |
| Tooling | GraphiQL provides excellent developer experience |

### PostgreSQL (Database)

| Reason | Benefit |
|--------|---------|
| Reliability | ACID compliance and data integrity |
| Bitnami Chart | Easy Kubernetes deployment |
| Performance | Excellent read/write performance |

### Kubernetes + Helm (Infrastructure)

| Reason | Benefit |
|--------|---------|
| Orchestration | Deployment, scaling, health management |
| Helm | Reproducible, versioned deployments |
| HPA | Supports 10x traffic scaling |
| Ingress + Cert-Manager | Automatic TLS |

### Quay.io (Container Registry)

| Reason | Benefit |
|--------|---------|
| Security | Image scanning and vulnerability detection |
| GitHub Actions | Excellent CI/CD integration |

## Project Structure

```
access-tracker/
├── backend/
│   ├── cmd/server/
│   │   └── main.go           # Entry point, router setup
│   └── internal/
│       ├── auth/
│       │   └── auth.go       # JWT + bcrypt utilities
│       ├── db/
│       │   └── conn.go       # PostgreSQL connection pool
│       ├── graphql/
│       │   ├── graphql.go    # Schema, resolvers
│       │   └── server.go     # GraphQL handler
│       └── models/
│           ├── access_request.go
│           └── user.go        # User + AuditLog models
├── frontend/
│   └── src/
│       ├── components/
│       │   ├── AccessRequestForm.vue
│       │   ├── RequestList.vue
│       │   └── StatusBadge.vue
│       ├── apollo.ts         # Apollo/GraphQL client
│       ├── router.ts         # Vue Router config
│       └── main.ts           # Entry point
├── k8s/access-tracker/
│   ├── values.yaml           # Base values
│   ├── values-dev.yaml       # Dev overrides
│   ├── values-prod.yaml     # Prod overrides
│   └── templates/
│       ├── deployment-backend.yaml
│       ├── deployment-frontend.yaml
│       ├── service-backend.yaml
│       ├── service-frontend.yaml
│       ├── ingress.yaml
│       ├── configmap.yaml
│       ├── hpa.yaml          # HorizontalPodAutoscaler
│       └── servicemonitor.yaml # Prometheus
├── .github/
│   ├── workflows/
│   │   ├── backend-ci.yml
│   │   ├── frontend-ci.yml
│   │   ├── deploy.yml
│   │   └── codeql.yml
│   └── dependabot.yml
└── docs/
    ├── architecture.md
    ├── runbook.md
    └── scaling.md
```

## API Reference

### GraphQL Queries

```graphql
# Get all access requests with optional filter
query GetAccessRequests($filter: AccessRequestFilter) {
  accessRequests(filter: $filter) {
    id
    requestId
    requester
    systemResource
    accessLevel
    justification
    status
    createdAt
    updatedAt
  }
}

# Get audit logs for an access request
query GetAuditLogs($accessRequestId: ID!) {
  auditLogs(accessRequestId: $accessRequestId) {
    id
    action
    oldValue
    newValue
    changedBy
    changedAt
  }
}
```

### GraphQL Mutations

```graphql
# Login and get JWT token
mutation Login($input: LoginInput!) {
  login(input: $input) {
    token
    username
  }
}

# Create new access request
mutation CreateAccessRequest($input: CreateAccessRequestInput!) {
  createAccessRequest(input: $input) {
    id
    requestId
    status
  }
}

# Update request status (creates audit log)
mutation UpdateStatus($input: UpdateAccessRequestStatusInput!) {
  updateAccessRequestStatus(input: $input) {
    id
    status
  }
}
```

## Security

### Authentication Flow

```mermaid
sequenceDiagram
    participant User as 👤 User
    participant Frontend as 📦 Frontend
    participant Backend as ⚙️ Backend
    participant DB as 🗄️ PostgreSQL

    User->>Frontend: Enter credentials
    Frontend->>Backend: mutation login(username, password)
    Backend->>DB: SELECT password_hash WHERE username=?
    DB-->>Backend: password_hash
    Backend->>Backend: bcrypt.Compare(password, hash)
    Backend->>Backend: jwt.Sign(username)
    Backend-->>Frontend: { token, username }
    Frontend-->>User: Store token, logged in

    Note over User,Backend: Subsequent requests include Authorization header
    Frontend->>Backend: GET /graphql + Authorization: Bearer <token>
    Backend->>Backend: jwt.Validate(token)
    Backend-->>Frontend: GraphQL response
```

### Password Security

- Passwords hashed with **bcrypt** (cost factor 10)
- Tokens signed with **JWT** (HS256)
- `JWT_SECRET` environment variable required in production

## Monitoring

### Metrics Endpoint

The backend exposes metrics at `/metrics` for Prometheus scraping:

```bash
# Prometheus Operator scrapes ServiceMonitor
oc get servicemonitor -n access-tracker

# View metrics directly
kubectl exec -it <backend-pod> -n access-tracker -- wget -O- http://localhost:8080/metrics
```

### Health Checks

| Probe | Endpoint | Purpose |
|-------|----------|---------|
| Liveness | `/healthz` | Pod is alive |
| Readiness | `/healthz` | Pod can serve traffic |
| Startup | `/healthz` | Application initialized |

### Log Aggregation

Logs are written to stdout and collected by OpenShift aggregated logging:

```bash
# View logs via CLI
kubectl logs -l app.kubernetes.io/component=backend -n access-tracker -f

# Via OpenShift Console
# Observe -> Logging
```
