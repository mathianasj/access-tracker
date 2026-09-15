# Access Tracker Architecture

## Quick Reference: What Talks to What

```mermaid
graph LR
    User["👤 Browser"] -->|HTTPS| Ingress["🌐 Ingress"]
    Ingress -->|/api/*| Backend["⚙️ Go Backend<br/>:8080"]
    Ingress -->|/*| Frontend["📦 Vue SPA<br/>:80"]
    Frontend -->|GraphQL| Backend
    Backend -->|SQL| PostgreSQL["🗄️ PostgreSQL<br/>:5432"]
    GitHubActions["🐙 GitHub Actions"] -->|docker push| Quay["📦 Quay.io"]
    Quay -->|docker pull| K8s["☸️ K8s Cluster"]
    Prometheus["📊 Prometheus"] -->|scrape| Backend
    Prometheus -->|scrape| Frontend
```

**Connection Summary:**

| From | To | Protocol | Purpose |
|------|-----|----------|---------|
| Browser | Ingress | HTTPS | User traffic |
| Ingress | Frontend Pod | HTTP | Serve SPA |
| Ingress | Backend Pod | HTTP | Proxy /api/* |
| Frontend | Backend | HTTP/GraphQL | API calls |
| Backend | PostgreSQL | SQL | Data persistence |
| GitHub | Quay | HTTPS | Container registry |
| Prometheus | Backend | HTTP | Metrics scraping |

## 1. Architecture Overview (2-4 min)

### System Purpose

Access Tracker manages resource access permissions with:
- **Requesters** submit access requests
- **Approvers** review and approve/deny requests
- **System** maintains audit trail of all changes

### Why These Tools?

```mermaid
flowchart TD
    subgraph "Frontend Choices"
        Vue["Vue 3"] --> VueWhy["Component-based SPA<br/>TypeScript support<br/>Fast reactivity"]
        Apollo["Apollo Client"] --> ApolloWhy["Native GraphQL support<br/>Caching built-in"]
    end

    subgraph "Backend Choices"
        Go["Go"] --> GoWhy["Single binary<br/>Fast startup<br/>Native concurrency"]
        Chi["Chi Router"] --> ChiWhy["Lightweight<br/>Standard library compatible"]
        GraphQL["GraphQL"] --> GraphQLWhy["Flexible queries<br/>Single endpoint<br/>Schema-first"]
    end

    subgraph "Infrastructure Choices"
        K8s["Kubernetes"] --> K8sWhy["Auto-scaling<br/>Self-healing<br/>Rolling updates"]
        Helm["Helm"] --> HelmWhy["Versioned configs<br/>Environment promotion"]
        PostgreSQL["PostgreSQL"] --> PGWhy["ACID compliant<br/>Bitnami chart<br/>JSON support"}
        Quay["Quay.io"] --> QuayWhy["Image scanning<br/>GitHub integration"]
    end
```

**Key Decisions:**

1. **Go over other languages**: Single binary deployment to K8s, fast boot time, excellent concurrency
2. **GraphQL over REST**: Frontend can request exact fields, single endpoint simplifies routing
3. **Vue over React**: Cleaner `<script setup>` syntax, less boilerplate
4. **K8s over raw Docker**: HPA for scaling, self-healing, rolling deploys
5. **Quay over Docker Hub**: Built-in vulnerability scanning, better GitHub Actions integration

---

## 2. Live Walkthrough: Request End-to-End (3-5 min)

### Trace a Create Request

```mermaid
sequenceDiagram
    participant B as 👤 Browser
    participant F as 📦 Vue SPA
    participant I as 🌐 Ingress
    participant BE as ⚙️ Go Backend
    participant DB as 🗄️ PostgreSQL

    B->>F: Fill form: requester, resource, access level
    F->>I: POST /api/graphql
    I->>BE: Forward GraphQL mutation
    BE->>BE: Validate JWT token
    BE->>BE: Parse GraphQL mutation
    BE->>DB: INSERT access_request
    DB->>DB: Write to table
    DB-->>BE: RETURNING id
    BE->>DB: INSERT audit_logs (CREATED)
    BE-->>I: GraphQL response
    I-->>F: HTTP 200 + JSON
    F->>F: Update Pinia store
    F->>B: Show success toast
```

### Code Path Trace

**Frontend:**
```
src/components/AccessRequestForm.vue
  → apollo.ts (Apollo Client)
    → POST /graphql { query: createAccessRequest }
```

**Backend:**
```
cmd/server/main.go
  → chi.Router.Handle("/graphql")
    → graphql.ContextMiddleware (add username to ctx)
    → graphql/server.go Handler
      → Schema.Exec()
        → Resolver.createAccessRequest()
          → Pool.QueryRow("INSERT INTO access_requests")
          → Pool.Exec("INSERT INTO audit_logs")
```

### Trace a Status Update

```mermaid
sequenceDiagram
    participant B as 👤 Browser
    participant F as 📦 Vue SPA
    participant BE as ⚙️ Go Backend
    participant DB as 🗄️ PostgreSQL

    B->>F: Click "Approve"
    F->>BE: mutation updateAccessRequestStatus(id, APPROVED)
    BE->>DB: SELECT status FROM access_requests WHERE id=?
    Note over BE,DB: oldStatus = "PENDING"
    BE->>DB: UPDATE access_requests SET status='APPROVED'
    BE->>DB: INSERT audit_logs (STATUS_CHANGED, PENDING → APPROVED)
    DB-->>BE: Updated row
    BE-->>F: { status: APPROVED }
    F->>B: Badge turns green
```

---

## 3. Deployment Narrative: Troubleshooting (3-5 min)

### Common Failure Scenarios

```mermaid
flowchart TD
    subgraph "Diagnose"
        Check["🔍 kubectl get pods"]
        Events["📋 kubectl describe pod"]
        Logs["📝 kubectl logs"]
    end

    subgraph "Failures"
        Image["🖼️ ImagePullBackOff"]
        Probe["❤️ Liveness probe failed"]
        DB["💾 Database connection"]
        TLS["🔒 Certificate issue"]
    end

    Check --> Events
    Events --> Logs
    Logs --> Image
    Logs --> Probe
    Logs --> DB
```

### Runbook Quick Reference

| Symptom | Command | Likely Cause |
|---------|---------|--------------|
| Pods stuck Pending | `kubectl describe pod` | Resource constraints, image pull |
| CrashLoopBackOff | `kubectl logs --previous` | App crash, bad env var |
| 404 from Ingress | `kubectl get endpoints` | Service has no pods |
| 503 from backend | `kubectl exec curl localhost:8080/healthz` | App unhealthy |
| TLS errors | `kubectl describe certificate` | cert-manager issue |

### Specific Scenario: Backend Pod CrashLoopBackOff

```bash
# 1. Identify the problem
kubectl get pods -n access-tracker
NAME                                READY   STATUS             RESTARTS   AGE
access-tracker-backend-7d8f9c6b5-xkq2l   0/1     CrashLoopBackOff   3          45s

# 2. Get details
kubectl describe pod access-tracker-backend-7d8f9c6b5-xkq2l -n access-tracker
Events:
  Type     Reason     Age   From               Message
  ----     ------     ----  ----               -------
  Warning  BackOff    10s   kubelet            Back-off restarting failed container

# 3. Check logs
kubectl logs access-tracker-backend-7d8f9c6b5-xkq2l -n access-tracker --previous
Error: failed to connect to database: connection refused

# 4. Root cause: PostgreSQL not ready yet
kubectl get pods -n access-tracker -l app.kubernetes.io/name=postgresql
NAME                              READY   STATUS    RESTARTS   AGE
access-tracker-postgresql-0       0/1     Running   0          15s

# 5. Solution: Wait for PostgreSQL, then restart backend
kubectl rollout restart deployment/access-tracker-backend -n access-tracker
```

**Key Insight:** Kubernetes starts pods in parallel - backend may start before PostgreSQL is ready. HPA and health checks handle this in production.

---

## 4. Monitoring Demo (2-4 min)

### Where to Look

```mermaid
flowchart LR
    subgraph "User Symptom"
        Browser["👤 Browser shows error"]
    end

    subgraph "Find the Problem"
        Route["🌐 Check Ingress route"]
        Pods["📦 Check Pods"]
        Logs["📝 Check Logs"]
        Metrics["📊 Check Metrics"]
    end

    subgraph "Tools"
        Openshift["🖥️ OpenShift Console"]
        CLI["⌨️ kubectl"]
        Prometheus["📈 Prometheus"]
        Grafana["📉 Grafana"]
    end

    Browser --> Route
    Route --> Openshift
    Pods --> CLI
    Logs --> Openshift
    Metrics --> Prometheus
    Metrics --> Grafana
```

### Demo Sequence

**1. OpenShift Console → Observe**
- See all pods, their status, CPU/memory
- Click any pod for details

**2. Prometheus Queries**

```promql
# Is the backend healthy?
up{service="access-tracker-backend"}

# Error rate
rate(http_requests_total{status=~"5.."}[5m])

# Latency p99
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# Pod CPU usage
pod:container_cpu_usage_seconds_total:sum{namespace="access-tracker"}
```

**3. View Logs**

```bash
# All backend logs in real-time
kubectl logs -l app.kubernetes.io/component=backend -n access-tracker -f --timestamps

# Search for errors
kubectl logs -l app.kubernetes.io/component=backend -n access-tracker | grep ERROR
```

**4. Check Health Endpoint**

```bash
kubectl exec -it <backend-pod> -n access-tracker -- wget -O- http://localhost:8080/healthz
```

### Alerting

Prometheus alerts fire when:
- Backend pod down > 1 minute
- Error rate > 1% for 5 minutes
- CPU > 90% for 10 minutes

Alertmanager routes to Slack/PagerDuty based on severity.

---

## 5. Forward-Looking: Scaling & Integration (2-4 min)

### 10x Traffic Scaling Path

```mermaid
flowchart TD
    subgraph "Current State"
        BE1["⚙️ Backend x2"]
        FE1["📦 Frontend x2"]
        DB1["🗄️ PostgreSQL x1"]
    end

    subgraph "10x Scale"
        BE2["⚙️ Backend x8"]
        FE2["📦 Frontend x4"]
        DB2["🗄️ PostgreSQL<br/>+ 2 Read Replicas"]
        Cache["💨 Redis Cache"]
        LB["⚖️ Load Balancer"]
    end

    subgraph "Future"
        HPA["📈 HPA"]
        VPA["📈 VPA"]
        PA["📊 Pod Disruption Budget"]
    end

    BE1 -->|HPA| BE2
    FE1 -->|HPA| FE2
    DB1 -->|replication| DB2
    BE2 --> Cache
    Cache -->|cache reads| DB2
```

### Database Connection Pooling

PostgreSQL has ~100 connection limit. At 10x:
- Backend pods: 8 × ~10 conn = 80 connections
- Risk: connection exhaustion

**Solution: pgBouncer**

```yaml
# Add as sidecar to backend
containers:
  - name: pgbouncer
    image: edoburu/pgbouncer:latest
    env:
      - name: DATABASE_URL
        value: "postgres://user:pass@localhost:5432/access_tracker"
      - name: POOL_MODE
        value: "transaction"
      - name: MAX_CLIENT_CONN
        value: "500"
```

Backend connects to `localhost:5432` (pgBouncer), pgBouncer connects to PostgreSQL with pooled connections.

### SSO Integration Path

```mermaid
flowchart LR
    subgraph "Current"
        FE["📦 Frontend"]
        BE["⚙️ Backend"]
    end

    subgraph "With SSO"
        Keycloak["🔐 Keycloak<br/>OIDC Provider"]
        FE2["📦 Frontend +<br/>oidc-client"]
        BE2["⚙️ Backend +<br/>JWT Validation"]
    end

    User["👤 User"] -->|Login| Keycloak
    Keycloak -->|Access Token| FE2
    FE2 -->|Forward Bearer| BE2
    BE2 -->|Validate| Keycloak
```

**Changes needed:**
1. Frontend: Add `oidc-client-js`, redirect to Keycloak
2. Backend: Validate JWT from Keycloak, extract user info
3. GraphQL: Add `requireAuth` directive

### Multi-Region Deployment

```mermaid
flowchart LR
    subgraph "Region US-East"
        FE1["📦 Frontend"]
        BE1["⚙️ Backend"]
        DB1["🗄️ PostgreSQL Primary"]
    end

    subgraph "Region EU-West"
        FE2["📦 Frontend"]
        BE2["⚙️ Backend"]
        DB2["🗄️ PostgreSQL Read Replica"]
    end

    subgraph "Traffic"
        DNS["🌍 Route 53"]
        LB["⚖️ Global LB"]
    end

    DNS --> LB
    LB --> FE1
    LB --> FE2
    BE1 -->|write| DB1
    BE2 -->|read| DB2
    DB1 -->|replication| DB2
```

### API Versioning Strategy

```graphql
# Current (v1)
POST /api/graphql

# Future (v2) - Breaking changes
POST /v2/api/graphql

# Ingress routes by path
/api/*     → backend-v1
/v2/api/*  → backend-v2
```

---

## Component Details

### Frontend (Vue 3 SPA)

**What it does:**
- Serves the UI at `/`
- Communicates with backend via GraphQL at `/api/graphql`
- Maintains local state with Pinia
- Apollo Client handles GraphQL and caching

**Key files:**
```
frontend/src/
├── components/
│   ├── AccessRequestForm.vue   # Create request form
│   ├── RequestList.vue         # List with filters
│   └── StatusBadge.vue         # Status display
├── apollo.ts                   # Apollo Client setup
├── router.ts                   # Vue Router config
└── main.ts                     # App entry point
```

### Backend (Go + GraphQL)

**What it does:**
- Handles all `/api/graphql` requests
- Validates JWT tokens from `Authorization: Bearer <token>`
- Writes to PostgreSQL
- Creates audit log entries on changes

**Key files:**
```
backend/
├── cmd/server/main.go          # Entry point, router
└── internal/
    ├── auth/auth.go            # JWT + bcrypt
    ├── db/conn.go              # PostgreSQL pool
    ├── graphql/
    │   ├── graphql.go          # Schema, resolvers
    │   └── server.go           # Handler
    └── models/
        ├── access_request.go   # Request model
        └── user.go             # User + AuditLog
```

### Kubernetes Deployment

**What it does:**
- Runs frontend and backend as Deployments
- PostgreSQL via Bitnami StatefulSet
- Ingress routes traffic
- HPA scales pods based on CPU/memory
- ServiceMonitors expose metrics to Prometheus

**Key files:**
```
k8s/access-tracker/
├── values.yaml                 # Base config
├── values-dev.yaml             # Dev overrides
├── values-prod.yaml            # Prod overrides
└── templates/
    ├── deployment-*.yaml       # Pod specs
    ├── service-*.yaml          # ClusterIP services
    ├── ingress.yaml             # Routing
    ├── hpa.yaml                 # Autoscaling
    └── servicemonitor.yaml      # Prometheus
```

---

## Environment Differences

| Setting | Development | Production |
|---------|-------------|------------|
| Replicas | 1 | 2+ |
| TLS | Disabled | Let's Encrypt |
| Persistence | EmptyDir | 10Gi PVC |
| Autoscaling | Disabled | HPA enabled |
| Resource Limits | 512Mi | 1Gi |
| Monitoring | Disabled | ServiceMonitors |

---

## API Reference

### GraphQL Schema Summary

**Queries:**
- `accessRequests(filter)` → List requests
- `accessRequest(id)` → Single request
- `auditLogs(accessRequestId)` → History

**Mutations:**
- `login(username, password)` → JWT token
- `createAccessRequest(input)` → New request
- `updateAccessRequestStatus(input)` → Change status

### Environment Variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `JWT_SECRET` | Token signing secret | Prod |
| `GRAPHQL_ENDPOINT` | Backend URL for frontend | Frontend |
