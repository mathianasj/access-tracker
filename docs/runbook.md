# Troubleshooting Runbook

Common deployment issues and how to diagnose them.

## Quick Health Checks

```bash
# Check overall pod status
kubectl get pods -n access-tracker

# Check pod details
kubectl describe pod <pod-name> -n access-tracker

# View pod logs
kubectl logs <pod-name> -n access-tracker

# Follow logs in real-time
kubectl logs -f <pod-name> -n access-tracker
```

## Pods Not Starting

### Symptoms
Pods are in `Pending`, `CrashLoopBackOff`, or `Error` state.

### Diagnosis

```bash
# Get detailed pod status
kubectl get pod <pod-name> -n access-tracker -o wide

# Check events for the pod
kubectl describe pod <pod-name> -n access-tracker | grep -A 10 "Events:"

# Check if there are resource constraints
kubectl describe nodes | grep -A 5 "Allocated resources"
```

### Common Causes

1. **Pending due to resource constraints**: Scale down other pods or add more cluster resources
2. **Image pull issues**: Check Quay.io credentials and image tags
3. **OOMKilled**: Increase memory limits in values file
4. **CrashLoopBackOff**: Check logs for application errors

## Image Pull Failures

### Symptoms
Pods stuck in `ContainerCreating` or `ImagePullBackOff` state.

### Diagnosis

```bash
# Check pod events
kubectl describe pod <pod-name> -n access-tracker | grep -A 5 "Events:"

# Verify image exists in Quay.io
# Check image pull secret if private registry
kubectl get secret -n access-tracker
```

### Solutions

```bash
# If using private registry, create image pull secret
kubectl create secret docker-registry quay-secret \
  --docker-server=quay.io \
  --docker-username=<username> \
  --docker-password=<token> \
  -n access-tracker

# Update service account to use the secret
kubectl patch serviceaccount default -p '{"imagePullSecrets":[{"name":"quay-secret"}]}' -n access-tracker
```

## Database Connection Issues

### Symptoms
Backend pods crash or cannot serve requests; connection errors in logs.

### Diagnosis

```bash
# Check PostgreSQL pod status
kubectl get pods -n access-tracker -l app.kubernetes.io/name=postgresql

# Check database credentials in configmap
kubectl get configmap access-tracker -n access-tracker -o yaml

# Test database connectivity from backend pod
kubectl exec -it <backend-pod> -n access-tracker -- nc -zv postgresql 5432
```

### Solutions

```bash
# Verify PostgreSQL is running
kubectl rollout status deployment/access-tracker-postgresql -n access-tracker

# Check PVC status if using persistence
kubectl get pvc -n access-tracker

# Restart backend after database is ready
kubectl rollout restart deployment/access-tracker-backend -n access-tracker
```

## Health Check Failures

### Symptoms
Pods pass `kubectl get pods` but are marked unhealthy by Kubernetes probes.

### Diagnosis

```bash
# Check health endpoint directly
kubectl exec -it <backend-pod> -n access-tracker -- wget -O- http://localhost:8080/healthz

# Check probe configuration
kubectl get pod <pod-name> -n access-tracker -o jsonpath='{.spec.containers[*].livenessProbe}'
kubectl get pod <pod-name> -n access-tracker -o jsonpath='{.spec.containers[*].readinessProbe}'
```

### Solutions

```bash
# Adjust probe timings if database takes time to initialize
# Edit the deployment to increase initialDelaySeconds

# For backend startup issues, check environment variables
kubectl exec -it <backend-pod> -n access-tracker -- env | grep -i db
```

## Ingress Not Routing

### Symptoms
External requests return 404 or timeout; Ingress shows no backend endpoints.

### Diagnosis

```bash
# Check Ingress status
kubectl get ingress -n access-tracker

# Describe Ingress for details
kubectl describe ingress access-tracker -n access-tracker

# Check nginx-ingress controller pods
kubectl get pods -n ingress-nginx

# Check endpoint existence
kubectl get endpoints -n access-tracker
```

### Solutions

```bash
# Check if ingress controller is running
kubectl rollout status deployment/ingress-nginx-controller -n ingress-nginx

# Verify backend/frontend services have endpoints
kubectl get endpoints access-tracker-backend -n access-tracker
kubectl get endpoints access-tracker-frontend -n access-tracker

# Check DNS resolution
nslookup <your-hostname>
```

## Helm Deployment Issues

### Symptoms
`helm upgrade` fails or results in unexpected configuration.

### Diagnosis

```bash
# List current helm releases
helm list -n access-tracker

# Check revision history
helm history access-tracker -n access-tracker

# View rendered templates
helm template ./k8s/access-tracker -f ./k8s/access-tracker/values-prod.yaml
```

### Solutions

```bash
# Rollback to previous revision
helm rollback access-tracker -n access-tracker

# Dry run before applying
helm upgrade --dry-run --debug access-tracker ./k8s/access-tracker -f ./k8s/access-tracker/values-prod.yaml -n access-tracker

# Clean up and reinstall if needed
helm uninstall access-tracker -n access-tracker
helm install access-tracker ./k8s/access-tracker -f ./k8s/access-tracker/values-prod.yaml -n access-tracker
```

## TLS/SSL Certificate Issues

### Symptoms
HTTPS requests fail; certificate errors in browser.

### Diagnosis

```bash
# Check cert-manager status
kubectl get certificate -n access-tracker

# Describe certificate for details
kubectl describe certificate access-tracker-tls -n access-tracker

# Check cert-manager pods
kubectl get pods -n cert-manager
```

### Solutions

```bash
# Check if Let's Encrypt is responding
kubectl logs -n cert-manager -l app=cert-manager

# Force certificate recreation
kubectl delete certificate access-tracker-tls -n access-tracker
kubectl delete secret access-tracker-tls -n access-tracker
# cert-manager will automatically recreate these
```

## Monitoring and Logs

### OpenShift Monitoring Stack

Access Tracker uses OpenShift's built-in monitoring stack:

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Dashboards for visualization
- **Alertmanager**: Alert routing

#### Accessing the Monitoring Dashboards

```bash
# Access Prometheus console (requires cluster-admin or user-workload-monitoring role)
oc get routes -n openshift-monitoring prometheus-k8s

# Access Grafana
oc get routes -n openshift-gitops grafana

# Or use the OpenShift Console -> Observe
```

#### Viewing Metrics in Prometheus

```bash
# Create a port-forward to Prometheus
oc port-forward -n openshift-monitoring prometheus-k8s 9090:9090

# Then open http://localhost:9090 in your browser
```

#### Useful PromQL Queries

```promql
# Backend request rate
rate(http_requests_total{service="access-tracker-backend"}[5m])

# Error rate
rate(http_requests_total{service="access-tracker-backend", status=~"5.."}[5m])

# Pod CPU usage
pod:container_cpu_usage_seconds_total:sum{namespace="access-tracker"}

# Pod memory usage
pod:container_memory_working_set_bytes:sum{namespace="access-tracker"}

# Backend health status
up{service="access-tracker-backend"}
```

#### ServiceMonitor Status

```bash
# Check if ServiceMonitors are discovered
oc get servicemonitor -n access-tracker

# Check Prometheus operator logs
oc logs -n openshift-monitoring prometheus-operator-0 -c prometheus-operator
```

### Viewing Application Logs

#### Using OpenShift Aggregated Logging

```bash
# Access logs via OpenShift Console -> Observe -> Logging
# Or via CLI with journalctl/oc logs

# All backend logs
kubectl logs -l app.kubernetes.io/component=backend -n access-tracker -f

# All frontend logs
kubectl logs -l app.kubernetes.io/component=frontend -n access-tracker -f

# With timestamps
kubectl logs -l app.kubernetes.io/component=backend -n access-tracker -f --timestamps

# Query logs using Elasticsearch (if configured)
oc exec -it <pod> -n access-tracker -- curl -sk https://elasticsearch.<namespace>.svc:9200/_search
```

### Common Log Patterns to Watch

```
# Database connection errors
ERROR.*database.*connection

# Health check failures
Health check failed

# Out of memory
OOMKilled

# Image pull errors
ImagePullBackOff

# Readiness probe failures
Readiness probe failed
```

## Getting Help

1. Check pod events: `kubectl describe pod <pod-name>`
2. Check application logs: `kubectl logs <pod-name>`
3. Verify configuration: `kubectl get configmap <configmap-name> -o yaml`
4. Check Helm values: `helm get values access-tracker -n access-tracker`
