# Research: kube-prometheus-stack for Local Kind Monitoring

**Date**: 2026-04-10
**Purpose**: Technical research for implementing optional monitoring stack (Prometheus Operator + Grafana) in kindenv

## 1. Chart Overview

### Chart Identity

| Field | Value |
|-------|-------|
| **Chart Name** | `kube-prometheus-stack` |
| **Latest Stable Version** | `83.4.0` |
| **App Version** | `v0.90.1` (Prometheus Operator) |
| **Prometheus Image** | `quay.io/prometheus/prometheus:v3.11.1` |
| **Grafana Subchart Version** | `11.6.0` |
| **Minimum Kubernetes** | `>=1.25.0-0` |
| **License** | Apache-2.0 |

### Helm Repository

| Method | Value |
|--------|-------|
| **Repo Name** | `prometheus-community` |
| **Repo URL** | `https://prometheus-community.github.io/helm-charts` |
| **OCI Artifact** | `oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack` |

```bash
# Traditional repo method (consistent with existing kindenv patterns)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

### Bundled Dependencies

The chart installs these sub-charts automatically:

| Dependency | Version | Condition Flag |
|------------|---------|---------------|
| kube-state-metrics | 7.2.2 | `kubeStateMetrics.enabled` |
| prometheus-node-exporter | 4.53.1 | `nodeExporter.enabled` |
| grafana | 11.6.0 | `grafana.enabled` |
| prometheus-windows-exporter | 0.12.x | `windowsMonitoring.enabled` (default: false) |
| CRDs | 0.0.0 | `crds.enabled` |

---

## 2. Helm Values for Local Kind Deployment

### Decision: Use kube-prometheus-stack with Grafana subchart
**Rationale**: Single Helm release deploys the entire monitoring stack (Prometheus Operator, Prometheus, Grafana, kube-state-metrics, node-exporter) with pre-built dashboards and auto-configured datasources. This matches the spec's "zero manual configuration" requirement (SC-005).
**Alternatives considered**: Separate Prometheus + Grafana charts (more configuration burden, no pre-wired dashboards), raw manifests (harder to maintain and upgrade).

### Complete Values Configuration

```yaml
# --- Namespace ---
# Deployed via: helm install ... --namespace monitoring --create-namespace
# No namespaceOverride needed when using --namespace flag

# --- Alertmanager: DISABLED (out of scope per spec) ---
alertmanager:
  enabled: false

# --- ThanosRuler: DISABLED (not needed for local dev) ---
thanosRuler:
  enabled: false

# --- Grafana Configuration ---
grafana:
  enabled: true

  # Anonymous access — no login screen (FR-013, spec clarification)
  grafana.ini:
    auth.anonymous:
      enabled: true
      org_role: Admin
    auth:
      disable_login_form: true
    security:
      allow_embedding: true

  # Expose via NodePort (FR-005)
  service:
    type: NodePort
    nodePort: 31300     # Configurable in kindenv.yaml

  # Resource limits (FR-006, SC-003: combined <1 CPU, <1GB)
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi

  # Pre-built dashboards enabled by default (FR-004)
  defaultDashboardsEnabled: true
  defaultDashboardsTimezone: utc
  defaultDashboardsEditable: true

  # Sidecar discovers dashboards and datasources from ConfigMaps
  sidecar:
    dashboards:
      enabled: true
      label: grafana_dashboard
      labelValue: "1"
      searchNamespace: ALL
    datasources:
      enabled: true
      defaultDatasourceEnabled: true
      isDefaultDatasource: true

  # No persistence needed for local dev (spec assumption)
  persistence:
    enabled: false

# --- Prometheus Configuration ---
prometheus:
  enabled: true

  prometheusSpec:
    # Retention (FR-007, spec assumption: 24h default)
    retention: 24h

    # Resource limits (FR-006, SC-003)
    resources:
      requests:
        cpu: 200m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi

    # No persistent storage for local dev — use emptyDir
    storageSpec: {}

    # ServiceMonitor auto-discovery across ALL namespaces (FR-009)
    # Setting these to false means Prometheus picks up ALL ServiceMonitors
    # and PodMonitors regardless of Helm release labels
    serviceMonitorSelectorNilUsesHelmValues: false
    serviceMonitorSelector: {}
    serviceMonitorNamespaceSelector: {}

    podMonitorSelectorNilUsesHelmValues: false
    podMonitorSelector: {}
    podMonitorNamespaceSelector: {}

    # Also discover all PrometheusRules and ScrapeConfigs
    ruleSelectorNilUsesHelmValues: false
    ruleSelector: {}
    ruleNamespaceSelector: {}

  # Prometheus service (ClusterIP is fine — accessed via Grafana internally)
  service:
    type: ClusterIP

# --- Prometheus Operator ---
prometheusOperator:
  enabled: true
  resources:
    requests:
      cpu: 100m
      memory: 64Mi
    limits:
      cpu: 200m
      memory: 128Mi

# --- Node Exporter (useful even in Kind for node-level metrics) ---
nodeExporter:
  enabled: true

# --- Kube State Metrics ---
kubeStateMetrics:
  enabled: true

# --- Kubernetes component scrapers ---
kubernetesServiceMonitors:
  enabled: true
kubeApiServer:
  enabled: true
kubelet:
  enabled: true
coreDns:
  enabled: true
kubeControllerManager:
  enabled: true
kubeScheduler:
  enabled: true
kubeEtcd:
  enabled: true

# kube-proxy metrics are often inaccessible in Kind (bound to 127.0.0.1)
kubeProxy:
  enabled: false

# Windows monitoring not needed
windowsMonitoring:
  enabled: false

# Default alerting rules — keep enabled for dashboards that reference them
defaultRules:
  create: true
```

### Key Values Explained

#### Grafana Anonymous Access

The Grafana subchart passes `grafana.ini` values directly to the Grafana configuration file. The combination of `auth.anonymous.enabled: true` with `org_role: Admin` and `auth.disable_login_form: true` ensures:
- No login screen appears
- Anonymous users get full Admin access (appropriate for local dev only)
- All pre-built dashboards are immediately accessible

#### ServiceMonitor Auto-Discovery (FR-009)

By default, `serviceMonitorSelectorNilUsesHelmValues: true` causes Prometheus to only discover ServiceMonitors with labels matching the Helm release. Setting this to `false` with empty selectors enables discovery of **all** ServiceMonitors in **all** namespaces. This is critical for scraping custom application metrics from any namespace.

For application pods to be scraped, they need a `ServiceMonitor` or `PodMonitor` CRD pointing to their metrics endpoint. The chart does **not** support annotation-based discovery (`prometheus.io/scrape`). Application teams should create ServiceMonitor resources like:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-app
  namespace: default
spec:
  selector:
    matchLabels:
      app: my-app
  endpoints:
    - port: metrics
      path: /metrics
      interval: 15s
```

#### Resource Budget

| Component | CPU Request | CPU Limit | Memory Request | Memory Limit |
|-----------|-----------|-----------|---------------|-------------|
| Prometheus | 200m | 500m | 256Mi | 512Mi |
| Grafana | 100m | 200m | 128Mi | 256Mi |
| Prometheus Operator | 100m | 200m | 64Mi | 128Mi |
| kube-state-metrics | (defaults) | (defaults) | (defaults) | (defaults) |
| node-exporter | (defaults) | (defaults) | (defaults) | (defaults) |
| **Totals (explicit)** | **400m** | **900m** | **448Mi** | **896Mi** |

This fits within the SC-003 target of 1 CPU / 1 GB combined for the core components.

---

## 3. Pre-Built Grafana Dashboards

The chart bundles **30+ dashboards** as ConfigMaps (under `templates/grafana/dashboards-1.14/`), auto-loaded by the Grafana sidecar. Key dashboards relevant for local development:

### Kubernetes Cluster Dashboards
| Dashboard | Description |
|-----------|-------------|
| **k8s-resources-cluster** | Cluster-wide CPU, memory, and network resource usage |
| **k8s-resources-namespace** | Per-namespace resource consumption breakdown |
| **k8s-resources-node** | Per-node resource usage |
| **k8s-resources-pod** | Per-pod CPU, memory, and network metrics |
| **k8s-resources-workload** | Resource usage by workload type (Deployment, StatefulSet, etc.) |
| **k8s-resources-workloads-namespace** | Workload resources scoped per namespace |

### Networking Dashboards
| Dashboard | Description |
|-----------|-------------|
| **cluster-total** | Cluster-wide network bandwidth and packet rates |
| **namespace-by-pod** | Network traffic by pod within a namespace |
| **namespace-by-workload** | Network traffic by workload within a namespace |
| **pod-total** | Per-pod network I/O |
| **workload-total** | Per-workload network I/O |

### Infrastructure Dashboards
| Dashboard | Description |
|-----------|-------------|
| **nodes** | Node Exporter / Linux node overview (CPU, memory, disk, network) |
| **node-cluster-rsrc-use** | Cluster-level node resource utilization |
| **node-rsrc-use** | Individual node resource utilization |
| **kubelet** | Kubelet operation metrics |
| **persistentvolumesusage** | PVC usage metrics |

### Monitoring Stack Self-Monitoring
| Dashboard | Description |
|-----------|-------------|
| **prometheus** | Prometheus self-monitoring (scrape durations, TSDB stats, memory) |
| **grafana-overview** | Grafana performance and request metrics |
| **alertmanager-overview** | Alertmanager overview (deployed even when alertmanager disabled — dashboard just shows no data) |
| **prometheus-remote-write** | Remote write metrics (if configured) |

### Kubernetes Control Plane
| Dashboard | Description |
|-----------|-------------|
| **apiserver** | API server request rates, latencies, and error rates |
| **controller-manager** | Controller manager work queue and reconciliation metrics |
| **scheduler** | Scheduler e2e latency and queue metrics |
| **etcd** | etcd database size, leader changes, and gRPC stats |
| **proxy** | kube-proxy sync rules and connection metrics |
| **k8s-coredns** | CoreDNS request rates and cache stats |

### Windows & Other (can be ignored for Kind/Linux)
- k8s-resources-windows-cluster, k8s-resources-windows-namespace, k8s-resources-windows-pod
- k8s-windows-cluster-rsrc-use, k8s-windows-node-rsrc-use
- nodes-aix, nodes-darwin

All dashboards are controlled by `grafana.defaultDashboardsEnabled: true` (default). They are provisioned via ConfigMaps with the label `grafana_dashboard: "1"` and discovered by the Grafana sidecar.

---

## 4. Kind-Specific Considerations

### Port Mapping

Kind requires explicit `extraPortMappings` in the cluster config to expose NodePort services on the host. The Grafana NodePort (31300) must be declared:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 31300
        hostPort: 31300
        protocol: TCP
```

**Important**: Port mappings can only be set at cluster creation time. If the monitoring feature is enabled after the cluster already exists, the cluster must be recreated with the updated port mappings. This is consistent with how other kindenv components handle port exposure (FR-015).

### Resource Constraints

Kind clusters run inside Docker containers, sharing host resources. Key concerns:

- **Memory pressure**: The monitoring stack adds ~900Mi memory overhead. The spec assumes 8GB+ allocated to Docker (spec Assumptions section). On constrained machines, reduce Prometheus memory limits or retention.
- **CPU throttling**: Prometheus startup (WAL replay) can spike CPU briefly. The default limits (500m for Prometheus) are reasonable.
- **Disk I/O**: Prometheus TSDB writes can contend with other pods. Using `emptyDir` (no PVC) avoids storage provisioner issues in Kind.

### Storage

- **No default StorageClass provisioner in Kind** that supports dynamic PVC expansion. Kind uses `rancher.io/local-path` or `standard` StorageClass backed by `local-path-provisioner`.
- **Recommendation**: Skip persistent storage entirely for local dev (`storageSpec: {}`). Prometheus uses `emptyDir` by default when no storage is specified. Data is lost on pod restart, but with 24h retention this is acceptable (spec assumption).
- If persistence is desired later, `local-path` StorageClass works but PVCs are not portable across cluster recreations.

### Control Plane Component Scraping

Kind exposes control plane components differently than managed Kubernetes:

| Component | Accessible in Kind? | Notes |
|-----------|---------------------|-------|
| kube-apiserver | ✅ Yes | Scraped via kubernetes service endpoints |
| kubelet / cAdvisor | ✅ Yes | Available on each node |
| kube-controller-manager | ⚠️ Varies | May need `--bind-address=0.0.0.0` in Kind config |
| kube-scheduler | ⚠️ Varies | May need `--bind-address=0.0.0.0` in Kind config |
| etcd | ⚠️ Varies | Metrics port (2381) may not be exposed by default |
| kube-proxy | ❌ Often No | Binds to `127.0.0.1:10249` — disable `kubeProxy.enabled` |
| CoreDNS | ✅ Yes | Accessible on port 9153 |

**Recommendation**: Disable `kubeProxy` scraping. Leave others enabled — partial failures in scraping control plane components do not affect application monitoring. Some dashboards may show "No data" for certain panels; this is acceptable for local dev.

### CRD Installation

The chart installs Prometheus Operator CRDs (ServiceMonitor, PodMonitor, PrometheusRule, etc.) by default via `crds.enabled: true`. These CRDs are cluster-scoped and persist after `helm uninstall`. They must be manually deleted if a full cleanup is needed:

```bash
kubectl delete crd alertmanagerconfigs.monitoring.coreos.com
kubectl delete crd alertmanagers.monitoring.coreos.com
kubectl delete crd podmonitors.monitoring.coreos.com
kubectl delete crd probes.monitoring.coreos.com
kubectl delete crd prometheusagents.monitoring.coreos.com
kubectl delete crd prometheuses.monitoring.coreos.com
kubectl delete crd prometheusrules.monitoring.coreos.com
kubectl delete crd scrapeconfigs.monitoring.coreos.com
kubectl delete crd servicemonitors.monitoring.coreos.com
kubectl delete crd thanosrulers.monitoring.coreos.com
```

However, for `kindenv stop --delete` which deletes the entire Kind cluster, CRD cleanup is implicit.

### Webhook TLS in Kind

The Prometheus Operator uses admission webhooks for validating PrometheusRule resources. These require TLS certificates provisioned via a post-install job (`kube-webhook-certgen`). This works in Kind without issues — the chart handles it automatically via `prometheusOperator.admissionWebhooks.patch.enabled: true` (default).

---

## 5. Installation Commands

```bash
# Add Helm repo (during kindenv init — FR-012)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install (during kindenv start)
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --version 83.4.0 \
  --values monitoring-values.yaml \
  --wait \
  --timeout 5m

# Upgrade in place (re-running kindenv start with changed config — FR-016)
helm upgrade monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --version 83.4.0 \
  --values monitoring-values.yaml \
  --wait \
  --timeout 5m

# Uninstall (during kindenv stop or component disable)
helm uninstall monitoring --namespace monitoring
kubectl delete namespace monitoring
```

---

## 6. Configurable Parameters for kindenv.yaml

Based on the spec requirements, these parameters should be exposed to users:

```yaml
components:
  monitoring:
    enabled: true                          # FR-001
    namespace: monitoring                  # FR-008 (default, not typically changed)
    chartVersion: "83.4.0"                 # FR-014
    grafana:
      nodePort: 31300                      # FR-005
    prometheus:
      retention: "24h"                     # FR-007
    resources:                             # FR-006
      prometheus:
        cpu: "500m"
        memory: "512Mi"
      grafana:
        cpu: "200m"
        memory: "256Mi"
```

---

## 7. Verification Checklist

After deployment, verify the stack is operational:

```bash
# All pods running in monitoring namespace
kubectl get pods -n monitoring

# Expected pods:
# - monitoring-kube-prometheus-stack-operator-*   (1 pod)
# - prometheus-monitoring-kube-prometheus-stack-prometheus-0  (1 pod)
# - monitoring-grafana-*                          (1 pod)
# - monitoring-kube-state-metrics-*               (1 pod)
# - monitoring-prometheus-node-exporter-*         (1 daemonset pod per node)

# Grafana accessible on host
curl -s http://localhost:31300/api/health
# Expected: {"commit":"...","database":"ok","version":"..."}

# Prometheus targets are being scraped
curl -s http://localhost:31300/api/datasources/proxy/uid/prometheus/api/v1/targets | jq '.data.activeTargets | length'
# Expected: >0

# Pre-built dashboards loaded
curl -s http://localhost:31300/api/search?type=dash-db | jq 'length'
# Expected: 25+ dashboards
```
