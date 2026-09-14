# Scaling Configuration

This document covers scalability considerations for the Access Tracker system, supporting 10x traffic growth, SSO integration, and production-ready deployments.

## Horizontal Pod Autoscaling (HPA)

The Access Tracker uses Kubernetes HPA to automatically scale pods based on CPU and memory utilization.

### Current Configuration

HPA is configured in `values-prod.yaml`:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### How HPA Works

1. **CPU-based scaling**: When average CPU utilization exceeds 70%, HPA scales up (up to maxReplicas)
2. **Memory-based scaling**: When average memory utilization exceeds 80%, additional pods are added
3. **Scale-down**: HPA scales down when utilization drops below thresholds

### Prerequisites for HPA

```bash
# Ensure metrics-server is installed in your cluster
kubectl get pod -n kube-system -l k8s-app=metrics-server

# If not installed:
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

### Custom HPA Configuration

To customize HPA behavior, modify `values-prod.yaml`:

```yaml
autoscaling:
  enabled: true
  minReplicas: 3          # Minimum pods (higher for production)
  maxReplicas: 20         # Maximum pods for burst traffic
  targetCPUUtilizationPercentage: 60   # More aggressive scaling
  targetMemoryUtilizationPercentage: 70
```

## Resource Limits

Proper resource limits ensure stable performance and enable HPA.

### Backend (Go API Server)

```yaml
backend:
  resources:
    limits:
      cpu: "1000m"        # 1 CPU core
      memory: "1Gi"       # 1 GB RAM
    requests:
      cpu: "250m"        # Guaranteed CPU
      memory: "256Mi"    # Guaranteed memory
```

### Frontend (Vue SPA)

```yaml
frontend:
  resources:
    limits:
      cpu: "500m"
      memory: "512Mi"
    requests:
      cpu: "100m"
      memory: "128Mi"
```

### PostgreSQL

```yaml
postgresql:
  primary:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
```

## Database Connection Pooling

For high-traffic scenarios, consider pgBouncer for PostgreSQL connection pooling.

### Why pgBouncer?

- PostgreSQL has a default limit of ~100 connections
- Each Access Tracker request typically holds a connection briefly
- At 10x traffic, connection exhaustion becomes a risk

### Deployment Option

Deploy pgBouncer as a sidecar or separate service:

```yaml
# Example pgBouncer sidecar for backend
containers:
  - name: pgBouncer
    image: edoburu/pgbouncer:latest
    ports:
      - containerPort: 5432
    env:
      - name: DATABASE_URL
        value: "postgres://user:pass@localhost:5432/access_tracker"
      - name: POOL_MODE
        value: "transaction"
      - name: MAX_CLIENT_CONN
        value: "500"
      - name: DEFAULT_POOL_SIZE
        value: "25"
```

### Connection String Update

When using pgBouncer, update the backend connection string:

```
postgresql://user:pass@pgbouncer:5432/access_tracker?pool_mode=transaction
```

## Caching Strategies

### GraphQL Query Caching

Consider Redis caching for expensive GraphQL queries.

### Implementation Options

1. **DataLoader pattern**: Batch and cache database queries within a request
2. **Redis caching**: Cache frequent queries (e.g., list of pending requests)
3. **CDN caching**: Cache static assets at the edge

### Redis Example

```yaml
# Add Redis to Helm values
redis:
  enabled: true
  architecture: replication
  auth:
    enabled: true
```

Backend code would then cache frequently-accessed data:

```go
// Pseudocode for caching
func getCachedRequests(ctx context.Context, key string) ([]Request, error) {
    cached, err := redis.Get(ctx, key)
    if err == nil {
        return unmarshal(cached)
    }
    requests, err := db.GetRequests()
    redis.Set(ctx, key, marshal(requests), 5*time.Minute)
    return requests, err
}
```

## Multi-Region Deployment

For global deployment, consider:

### Option 1: Multi-Region Read Replicas

- Deploy Access Tracker in multiple regions
- Use PostgreSQL read replicas for each region
- Route writes to a primary region

### Option 2: GitOps with ArgoCD

```yaml
# ArgoCD Application for multi-cluster
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: access-tracker
spec:
  destination:
    server: https://<cluster-api>
    namespace: access-tracker
  source:
    repoURL: https://github.com/mathianasj/access-tracker
    targetRevision: main
    path: k8s/access-tracker
    helm:
      valueFiles:
        - values-prod.yaml
```

## SSO Integration (OIDC)

### Integration Points

1. **Frontend**: Integrate OIDC provider (Keycloak, Auth0, Okta)
2. **Backend**: Validate JWT tokens from OIDC provider

### Keycloak Example

```yaml
# Frontend OIDC configuration
oidc:
  issuer: https://keycloak.example.com/realms/access-tracker
  clientId: access-tracker-frontend
  redirectUri: https://access-tracker.example.com/callback
```

### Backend Token Validation

The backend should validate tokens on protected endpoints:

```go
// Middleware pseudocode
func OIDCAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        claims, err := validator.ValidateToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## API Versioning Strategy

For SaaS-like API versioning:

### Versioning Approaches

1. **URL path versioning**: `/api/v1/requests`, `/api/v2/requests`
2. **Header versioning**: `API-Version: v2`
3. **Content negotiation**: `Accept: application/vnd.api.v2+json`

### Recommendation

Use URL path versioning for clarity:

```yaml
ingress:
  annotations:
    nginx.ingress.kubernetes.io/use-regex: "true"
  rules:
    - host: api.access-tracker.example.com
      http:
        paths:
          - path: /v1/api
            pathType: Prefix
            backend:
              service:
                name: access-tracker-backend-v1
          - path: /v2/api
            pathType: Prefix
            backend:
              service:
                name: access-tracker-backend-v2
```

## Capacity Planning

### 10x Traffic Estimate

| Component | Current | 10x Growth |
|-----------|---------|------------|
| Backend pods | 2 | 5-8 |
| Frontend pods | 2 | 3-4 |
| PostgreSQL | 1 primary | 1 primary + 2 read replicas |
| Connections | ~50 concurrent | ~200 concurrent |

### Scaling Commands

```bash
# Manual scale (not normally needed with HPA)
kubectl scale deployment access-tracker-backend --replicas=5 -n access-tracker

# Check HPA status
kubectl get hpa -n access-tracker

# View HPA metrics
kubectl describe hpa access-tracker-backend -n access-tracker
```

## Monitoring Metrics

Key metrics to track for scaling decisions:

- **Request rate**: requests/second by endpoint
- **Latency**: p50, p95, p99 response times
- **Error rate**: 5xx responses as percentage of total
- **Resource utilization**: CPU and memory per pod
- **Database connections**: active/idle/waiting
- **Queue depth**: if using async processing

## Prometheus Monitoring

Access Tracker is configured for OpenShift's built-in Prometheus monitoring via ServiceMonitors.

### Enabling Monitoring

```yaml
# In values-prod.yaml
monitoring:
  enabled: true
```

### Metrics Endpoint Requirement

For full application metrics, the backend must expose a `/metrics` endpoint. Add the Prometheus client library to `backend/go.mod`:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// In your router setup
router.Handle("/metrics", promhttp.Handler())
```

### Example Custom Metrics

```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}
```

### OpenShift User Workload Monitoring

For user-defined projects to have metrics collected:

```bash
# Check if user workload monitoring is enabled
oc get pods -n openshift-user-workload-monitoring

# If not, cluster-admin must enable it:
# oc edit configs.imageregistry.operator.openshift.io/cluster
```

## Further Reading

- [Kubernetes HPA Documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [PostgreSQL Connection Pooling](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [Redis Caching Patterns](https://redis.io/docs/manual patterns/)
- [OIDC Integration Guide](https://openid.net/developers/certified/)
