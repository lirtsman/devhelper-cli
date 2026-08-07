# CLI Contracts: Monitoring Stack

**Feature**: 001-monitoring-stack
**Date**: 2026-04-10

## Overview

This document defines the CLI interface contracts for the monitoring stack feature — new flags, configuration schema additions, and command output changes across `kindenv init`, `kindenv start`, `kindenv status`, and `kindenv stop`.

---

## 1. Configuration Schema Contract

### New `kindenv.yaml` Section

Added under `components`:

```yaml
components:
  monitoring:
    enabled: false                    # bool — opt-in, default false
    namespace: "monitoring"           # string — dedicated namespace
    chartVersion: "72.6.2"            # string — kube-prometheus-stack chart version
    grafana:
      nodePort: 31300                 # int — Grafana NodePort (30000-32767)
    prometheus:
      retention: "24h"               # string — Prometheus data retention duration
    resources:
      prometheus:
        cpu: "500m"                  # string — Prometheus CPU limit
        memory: "512Mi"              # string — Prometheus memory limit
      grafana:
        cpu: "200m"                  # string — Grafana CPU limit
        memory: "256Mi"              # string — Grafana memory limit
```

### Backward Compatibility

- Existing `kindenv.yaml` files without a `monitoring` key remain valid.
- When omitted, `monitoring.enabled` defaults to `false` (zero-value).
- No existing fields are modified or removed.

---

## 2. Command: `kindenv init`

### New Behavior

Registers the `prometheus-community` Helm repository alongside existing repos.

### Helm Repo Registration

```
Repo Name:  prometheus-community
Repo URL:   https://prometheus-community.github.io/helm-charts
```

### Output Contract

**Success:**
```
Adding Prometheus Community Helm repository...
✅ Prometheus Community Helm repository configured
```

**Already exists:**
```
Adding Prometheus Community Helm repository...
✅ Prometheus Community Helm repository already configured
```

**Failure (non-blocking):**
```
Adding Prometheus Community Helm repository...
⚠️  Warning: Failed to add Prometheus Community Helm repository: <error>
```

### Default Config Generation

When `kindenv init` generates a default `kindenv.yaml`, the monitoring section is included with `enabled: false` and all defaults populated.

---

## 3. Command: `kindenv start`

### New Flag

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--skip-monitoring` | (none) | `bool` | `false` | Skip installing the monitoring stack even if enabled in config |

### Deployment Behavior

**Preconditions:**
- `config.Components.Monitoring.Enabled == true`
- `--skip-monitoring` flag is not set

**Deployment sequence** (inserted after KEDA, before Redis in the component order):

1. Print: `Installing Monitoring Stack (kube-prometheus-stack)`
2. Create namespace (idempotent via `kubectl create namespace <ns> --dry-run=client -o yaml | kubectl apply -f -`)
3. Build Helm args array
4. Execute: `helm upgrade --install monitoring prometheus-community/kube-prometheus-stack ...`
5. Wait for key deployments to become ready
6. Print success or warning

### Helm Command Contract

```bash
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  --namespace <namespace> \
  --version <chartVersion> \
  --set alertmanager.enabled=false \
  --set thanosRuler.enabled=false \
  --set grafana.enabled=true \
  --set grafana."grafana\.ini"."auth\.anonymous".enabled=true \
  --set grafana."grafana\.ini"."auth\.anonymous".org_role=Admin \
  --set grafana."grafana\.ini".auth.disable_login_form=true \
  --set grafana."grafana\.ini".security.allow_embedding=true \
  --set grafana.service.type=NodePort \
  --set grafana.service.nodePort=<grafana.nodePort> \
  --set grafana.resources.requests.cpu=<resources.grafana.cpu> \
  --set grafana.resources.requests.memory=<resources.grafana.memory> \
  --set grafana.resources.limits.cpu=<resources.grafana.cpu> \
  --set grafana.resources.limits.memory=<resources.grafana.memory> \
  --set grafana.defaultDashboardsEnabled=true \
  --set grafana.persistence.enabled=false \
  --set prometheus.prometheusSpec.retention=<prometheus.retention> \
  --set prometheus.prometheusSpec.resources.requests.cpu=<resources.prometheus.cpu> \
  --set prometheus.prometheusSpec.resources.requests.memory=<resources.prometheus.memory> \
  --set prometheus.prometheusSpec.resources.limits.cpu=<resources.prometheus.cpu> \
  --set prometheus.prometheusSpec.resources.limits.memory=<resources.prometheus.memory> \
  --set prometheus.prometheusSpec.storageSpec.emptyDir.medium="" \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
  --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
  --set prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues=false \
  --set prometheus.service.type=ClusterIP \
  --set prometheusOperator.enabled=true \
  --set nodeExporter.enabled=true \
  --set kubeStateMetrics.enabled=true \
  --set kubeProxy.enabled=false \
  --set windowsMonitoring.enabled=false \
  --set defaultRules.create=true \
  --wait \
  --timeout 5m
```

> **Note**: The `--set` values for `grafana.ini` keys use escaped dots in Helm's `--set` syntax. The implementation may alternatively use a values file (`--values`) for cleaner handling of nested keys.

### Output Contract

**Success:**
```
Installing Monitoring Stack (kube-prometheus-stack)
Creating namespace: monitoring
✅ Monitoring stack installed successfully
   Grafana dashboard: http://localhost:<hostPort> (no login required)
```

**Failure (warn and continue — FR-017):**
```
Installing Monitoring Stack (kube-prometheus-stack)
Creating namespace: monitoring
❌ Error installing Monitoring Stack: <error details>
⚠️  Continuing despite Monitoring Stack installation failure...
```

**Skipped (disabled):**
No output — consistent with other disabled components.

**Skipped (flag):**
```
⏭️  Skipping Monitoring Stack installation (--skip-monitoring)
```

### Port Mapping Contract

When monitoring is enabled and `generateDefaultPortMappings` is called, append:

```go
{
    ContainerPort: "${{ components.monitoring.grafana.nodePort }}",
    HostPort:      3000,
    Protocol:      "TCP",
}
```

Default host port `3000` maps Grafana's NodePort `31300` to `localhost:3000`.

---

## 4. Command: `kindenv status`

### New Output Section

When `components.monitoring.enabled == true`:

**Healthy:**
```
- ✅ Monitoring Stack is installed and running
     Prometheus: 1/1 pods ready
     Grafana: 1/1 pods ready (http://localhost:<hostPort>)
     Node Exporter: 1/1 pods ready
```

**Degraded:**
```
- ⚠️  Monitoring Stack is partially running
     Prometheus: 1/1 pods ready
     Grafana: 0/1 pods ready
     Node Exporter: 1/1 pods ready
```

**Not running:**
```
- ⚠️  Monitoring Stack is not running or not installed
```

**Disabled:**
No monitoring-related output — consistent with other disabled components.

### Status Check Implementation

```bash
kubectl get pods -n <namespace> --no-headers -l "app.kubernetes.io/instance=monitoring"
```

Parse output to check readiness of:
- `monitoring-grafana-*`
- `prometheus-monitoring-kube-prometheus-stack-prometheus-*`
- `monitoring-prometheus-node-exporter-*`

---

## 5. Command: `kindenv stop`

### Behavior

When `--delete` is used, the Kind cluster deletion removes all resources including the monitoring namespace. No monitoring-specific cleanup is required — consistent with how all other Helm-based components are handled (FR-011).

No new flags or output changes for `kindenv stop`.

---

## 6. Validation Contract

### Config Validation Rules

Added to `KindEnvConfig.Validate()` when `Components.Monitoring.Enabled == true`:

| Field | Rule | Error Message |
|-------|------|---------------|
| `Namespace` | Not empty | `"monitoring namespace cannot be empty when monitoring is enabled"` |
| `ChartVersion` | Not empty | `"monitoring chartVersion cannot be empty when monitoring is enabled"` |
| `Grafana.NodePort` | In range 30000–32767 | `"monitoring grafana nodePort must be in range 30000-32767"` |
| `Resources.Prometheus.CPU` | Matches `^[0-9]+m?$` | `"monitoring prometheus cpu resource must be in valid format (e.g., 500m, 1)"` |
| `Resources.Prometheus.Memory` | Matches `^[0-9]+[KMGT]i$` | `"monitoring prometheus memory resource must be in valid format (e.g., 512Mi, 1Gi)"` |
| `Resources.Grafana.CPU` | Matches `^[0-9]+m?$` | `"monitoring grafana cpu resource must be in valid format (e.g., 200m, 1)"` |
| `Resources.Grafana.Memory` | Matches `^[0-9]+[KMGT]i$` | `"monitoring grafana memory resource must be in valid format (e.g., 256Mi, 1Gi)"` |
| `Prometheus.Retention` | Matches `^[0-9]+[hdwmy]$` | `"monitoring prometheus retention must be in valid format (e.g., 24h, 7d, 4w)"` |

---

## 7. Files Changed Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/kindenv/config.go` | Modify | Add `Monitoring` struct to `Components`, defaults in `LoadConfig` and `CreateDefaultConfig`, validation in `Validate`, port mapping in `generateDefaultPortMappings` |
| `cmd/kindenv_init.go` | Modify | Add `prometheus-community` Helm repo registration block |
| `cmd/kindenv_start.go` | Modify | Add `--skip-monitoring` flag, add monitoring deployment block |
| `cmd/kindenv_status.go` | Modify | Add monitoring status check block |
| `cmd/kindenv_start_test.go` | Modify | Add tests for monitoring deployment and skip flag |
| `cmd/kindenv_init_test.go` | Modify | Add test for Helm repo registration |
| `cmd/kindenv_status_test.go` | Modify | Add test for monitoring status output |
| `internal/kindenv/config_test.go` | Modify | Add tests for monitoring config defaults, validation, and YAML parsing |