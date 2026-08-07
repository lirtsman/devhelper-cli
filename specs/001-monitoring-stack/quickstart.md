# Quick Start: Monitoring Stack (Prometheus + Grafana)

**Target Audience**: Developers using devhelper-cli Kind-based local environments
**Prerequisites**: devhelper-cli installed, Kind cluster not yet created (or ready to recreate)
**Estimated Time**: ~5 minutes to a running monitoring stack

## Overview

The monitoring stack deploys [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) into your local Kind cluster, giving you:

- **Prometheus Operator** — automatic metrics collection from all cluster components
- **Grafana** — pre-built dashboards for Kubernetes nodes, pods, and namespaces
- **No login required** — Grafana is open-access for local development
- **Auto-discovery** — application pods exposing `/metrics` are scraped automatically

## Quick Start

### 1. Enable Monitoring in Configuration

Add the `monitoring` block to your `kindenv.yaml` under `components`:

```yaml
components:
  monitoring:
    enabled: true
    namespace: monitoring
    chartVersion: "72.6.2"
    grafana:
      nodePort: 31300
    prometheus:
      retention: "24h"
    resources:
      prometheus:
        cpu: "500m"
        memory: "512Mi"
      grafana:
        cpu: "200m"
        memory: "256Mi"
```

All fields above are defaults — setting `enabled: true` with no other options works out of the box.

### 2. Add Port Mapping

Ensure the Grafana NodePort is mapped from the Kind container to your host. Add this to the `cluster` section of `kindenv.yaml` if it is not already auto-generated:

```yaml
cluster:
  mapPorts:
    - containerPort: 31300
      hostPort: 3000
      protocol: TCP
```

This makes Grafana available at `localhost:3000` on your machine.

### 3. Initialize the Environment

Run `init` to register the required Helm repository (`prometheus-community`):

```bash
devhelper-cli kindenv init
```

This is idempotent — safe to run again if you've already initialized other components.

### 4. Start the Environment

Deploy all enabled components, including the monitoring stack:

```bash
devhelper-cli kindenv start
```

The monitoring stack typically reaches a ready state within 2–3 minutes. If it fails to deploy (insufficient resources, port conflict), a warning is logged and all other components continue deploying normally.

### 5. Access Grafana

Open your browser to:

```
http://localhost:3000
```

No login is required. You'll land on the Grafana home page with pre-built dashboards already available.

**Pre-built dashboards to explore** (navigate to Dashboards → Browse):

| Dashboard | What It Shows |
|---|---|
| Kubernetes / Compute Resources / Cluster | Cluster-wide CPU and memory usage |
| Kubernetes / Compute Resources / Namespace (Pods) | Per-pod resource usage within a namespace |
| Node Exporter / Nodes | Host-level CPU, memory, disk, and network |
| Kubernetes / Networking / Namespace (Pods) | Network traffic per pod |

### 6. Verify Monitoring Status

Check that all monitoring components are healthy:

```bash
devhelper-cli kindenv status
```

Expected output includes:

```
  Monitoring:
    Prometheus:   ✅ Running
    Grafana:      ✅ Running (http://localhost:3000)
```

You can also inspect pods directly:

```bash
kubectl get pods -n monitoring
```

## Monitor Custom Application Metrics

The monitoring stack automatically discovers scrape targets via **ServiceMonitor** CRDs (installed by the Prometheus Operator). To expose metrics from your own application:

1. Ensure your app serves a Prometheus-compatible `/metrics` endpoint.
2. Create a `ServiceMonitor` pointing to your app's `Service`:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-app-metrics
  namespace: default
  labels:
    release: monitoring   # Must match the Helm release label selector
spec:
  selector:
    matchLabels:
      app: my-app          # Must match your Service's labels
  endpoints:
    - port: http           # Named port on your Service
      path: /metrics
      interval: 15s
```

```bash
kubectl apply -f my-app-servicemonitor.yaml
```

Within 1–2 minutes, your custom metrics will appear in Grafana's **Explore** view (data source: Prometheus). Use PromQL to query them:

```
rate(http_requests_total{namespace="default", app="my-app"}[5m])
```

## Customize the Monitoring Stack

All monitoring settings live under `components.monitoring` in `kindenv.yaml`. Change values and re-run `devhelper-cli kindenv start` — the existing deployment is upgraded in place (no tear-down required).

### Change Data Retention

Keep metrics for 7 days instead of the default 24 hours:

```yaml
components:
  monitoring:
    enabled: true
    prometheus:
      retention: "7d"
```

> **Note**: Longer retention increases disk usage inside the Kind cluster. For local development, 24h–48h is usually sufficient.

### Adjust Resource Limits

Give Prometheus more headroom on a well-provisioned machine:

```yaml
components:
  monitoring:
    enabled: true
    resources:
      prometheus:
        cpu: "1000m"
        memory: "1Gi"
      grafana:
        cpu: "200m"
        memory: "256Mi"
```

### Change the Grafana Port

Serve Grafana on a different host port (e.g., `localhost:3001`):

```yaml
components:
  monitoring:
    enabled: true
    grafana:
      nodePort: 31301

cluster:
  mapPorts:
    - containerPort: 31301
      hostPort: 3001
      protocol: TCP
```

### Full Configuration Reference

```yaml
components:
  monitoring:
    enabled: true            # Enable/disable the entire stack (default: false)
    namespace: monitoring    # Kubernetes namespace (default: monitoring)
    chartVersion: "72.6.2"   # kube-prometheus-stack Helm chart version
    grafana:
      nodePort: 31300        # NodePort for Grafana (default: 31300)
    prometheus:
      retention: "24h"       # How long to keep metrics (default: 24h)
    resources:
      prometheus:
        cpu: "500m"          # Prometheus CPU request/limit
        memory: "512Mi"      # Prometheus memory request/limit
      grafana:
        cpu: "200m"          # Grafana CPU request/limit
        memory: "256Mi"      # Grafana memory request/limit
```

## Troubleshooting

### Grafana not reachable at localhost:3000

1. Verify the port mapping in `kindenv.yaml` matches the `grafana.nodePort` value.
2. Check that Grafana is running: `kubectl get pods -n monitoring | grep grafana`.
3. Check pod logs: `kubectl logs -n monitoring -l app.kubernetes.io/name=grafana`.

### Prometheus pods in CrashLoopBackOff

Usually caused by insufficient memory. Increase the Prometheus memory limit:

```yaml
resources:
  prometheus:
    memory: "1Gi"
```

Then re-run `devhelper-cli kindenv start`.

### No data in dashboards

Prometheus may still be completing its first scrape cycle. Wait 2 minutes after all pods are ready, then refresh. Verify targets are being scraped:

```bash
kubectl port-forward -n monitoring svc/monitoring-kube-prometheus-prometheus 9090:9090
```

Then open `http://localhost:9090/targets` to inspect active scrape targets.

### ServiceMonitor not picked up

Ensure the `release: monitoring` label is present on your `ServiceMonitor` metadata. The Prometheus Operator uses this label selector to discover monitors belonging to its stack.