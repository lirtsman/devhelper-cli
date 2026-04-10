# Data Model: Monitoring Stack Configuration

**Feature**: 001-monitoring-stack
**Date**: 2026-04-10
**Status**: Design Phase

## Overview

This document defines the data structures and configuration model for the optional monitoring stack (kube-prometheus-stack Helm chart — Prometheus Operator + Grafana) in the kindenv environment. The model follows the established anonymous nested struct pattern used by all other components inside `KindEnvConfig.Components`.

---

## Configuration Entity

### Monitoring Component Configuration

**Entity**: `MonitoringConfig`
**Location**: `internal/kindenv/config.go` (anonymous struct inside `Components`)
**Purpose**: Controls monitoring stack installation and configuration

#### Go Struct Definition

```go
type KindEnvConfig struct {
	// ... other fields ...

	Components struct {
		// ... existing components (Temporal, Redis, Dapr, MySQL, etc.) ...

		Monitoring struct {
			Enabled      bool   `yaml:"enabled"`
			Namespace    string `yaml:"namespace"`
			ChartVersion string `yaml:"chartVersion"`
			Grafana      struct {
				NodePort int `yaml:"nodePort"`
			} `yaml:"grafana"`
			Prometheus struct {
				Retention string `yaml:"retention"`
			} `yaml:"prometheus"`
			Resources struct {
				Prometheus struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"prometheus"`
				Grafana struct {
					CPU    string `yaml:"cpu"`
					Memory string `yaml:"memory"`
				} `yaml:"grafana"`
			} `yaml:"resources"`
		} `yaml:"monitoring"`
	} `yaml:"components"`
}
```

#### YAML Representation

```yaml
components:
  # ... other components ...

  monitoring:
    enabled: false
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

---

## Default Values

### Table

| Field | Type | Default | FR | Description |
|-------|------|---------|----|-------------|
| `enabled` | `bool` | `false` | FR-001 | Enables monitoring stack installation |
| `namespace` | `string` | `"monitoring"` | FR-008 | Dedicated namespace for monitoring resources |
| `chartVersion` | `string` | `"72.6.2"` | FR-014 | kube-prometheus-stack Helm chart version |
| `grafana.nodePort` | `int` | `31300` | FR-005 | Host-accessible NodePort for Grafana dashboard |
| `prometheus.retention` | `string` | `"24h"` | FR-007 | Prometheus time-series data retention duration |
| `resources.prometheus.cpu` | `string` | `"500m"` | FR-006 | CPU limit for Prometheus server |
| `resources.prometheus.memory` | `string` | `"512Mi"` | FR-006 | Memory limit for Prometheus server |
| `resources.grafana.cpu` | `string` | `"200m"` | FR-006 | CPU limit for Grafana |
| `resources.grafana.memory` | `string` | `"256Mi"` | FR-006 | Memory limit for Grafana |

### Code Defaults

```go
// In CreateDefaultConfig function (config.go)
config.Components.Monitoring.Enabled = false
config.Components.Monitoring.Namespace = "monitoring"
config.Components.Monitoring.ChartVersion = "72.6.2"
config.Components.Monitoring.Grafana.NodePort = 31300
config.Components.Monitoring.Prometheus.Retention = "24h"
config.Components.Monitoring.Resources.Prometheus.CPU = "500m"
config.Components.Monitoring.Resources.Prometheus.Memory = "512Mi"
config.Components.Monitoring.Resources.Grafana.CPU = "200m"
config.Components.Monitoring.Resources.Grafana.Memory = "256Mi"
```

**Rationale for Defaults**:
- `enabled: false` — Opt-in model consistent with all other components.
- `namespace: "monitoring"` — Dedicated namespace keeps monitoring resources isolated (FR-008).
- `chartVersion: "72.6.2"` — Pinned stable release of kube-prometheus-stack.
- `grafana.nodePort: 31300` — Avoids collision with existing NodePorts (MySQL 30306, RabbitMQ 31672/31673, OpenSearch 30920, etc.).
- `prometheus.retention: "24h"` — Sufficient for local dev; keeps storage footprint small.
- Combined resource defaults (700m CPU, 768Mi memory) stay within SC-003 budget of 1 CPU / 1 GB.

---

## Validation Rules

```go
// validateMonitoring checks the monitoring component configuration.
func (c *KindEnvConfig) validateMonitoring() error {
	m := c.Components.Monitoring
	if !m.Enabled {
		return nil
	}

	if strings.TrimSpace(m.Namespace) == "" {
		return fmt.Errorf("monitoring namespace cannot be empty")
	}

	if strings.TrimSpace(m.ChartVersion) == "" {
		return fmt.Errorf("monitoring chartVersion cannot be empty")
	}

	if m.Grafana.NodePort < 30000 || m.Grafana.NodePort > 32767 {
		return fmt.Errorf("monitoring grafana.nodePort must be in range 30000-32767, got %d", m.Grafana.NodePort)
	}

	if err := validateCPUFormat(m.Resources.Prometheus.CPU); err != nil {
		return fmt.Errorf("monitoring resources.prometheus.cpu: %w", err)
	}
	if err := validateMemoryFormat(m.Resources.Prometheus.Memory); err != nil {
		return fmt.Errorf("monitoring resources.prometheus.memory: %w", err)
	}
	if err := validateCPUFormat(m.Resources.Grafana.CPU); err != nil {
		return fmt.Errorf("monitoring resources.grafana.cpu: %w", err)
	}
	if err := validateMemoryFormat(m.Resources.Grafana.Memory); err != nil {
		return fmt.Errorf("monitoring resources.grafana.memory: %w", err)
	}

	return nil
}
```

### Validation Summary

| Field | Rule | Example Valid | Example Invalid |
|-------|------|---------------|-----------------|
| `namespace` | Non-empty after trimming whitespace | `"monitoring"` | `""`, `"  "` |
| `chartVersion` | Non-empty after trimming whitespace | `"72.6.2"` | `""` |
| `grafana.nodePort` | Integer in 30000–32767 | `31300` | `8080`, `0`, `40000` |
| `resources.*.cpu` | Kubernetes CPU format (`/^\d+m?$/`) | `"500m"`, `"1"` | `"lots"`, `""` |
| `resources.*.memory` | Kubernetes memory format (`/^\d+(Mi\|Gi)$/`) | `"512Mi"`, `"1Gi"` | `"big"`, `""` |
| `prometheus.retention` | Accepted by Prometheus (validated at deploy time) | `"24h"`, `"7d"` | — |

Validation runs only when `enabled` is `true`. When `enabled` is `false`, the component is skipped entirely and no validation is necessary.

---

## Relationship to Parent Struct

The monitoring struct is added as an anonymous nested struct inside `KindEnvConfig.Components`, identical in style to every other component:

```
KindEnvConfig
 └── Components                      (struct, yaml:"components")
      ├── Temporal     {...}          yaml:"temporal"
      ├── Redis        {...}          yaml:"redis"
      ├── Dapr         {...}          yaml:"dapr"
      ├── OpenSearch   {...}          yaml:"openSearch"
      ├── MetricsServer{...}          yaml:"metricsServer"
      ├── Keda         {...}          yaml:"keda"
      ├── MySQL        {...}          yaml:"mysql"
      ├── RabbitMQ     {...}          yaml:"rabbitmq"
      └── Monitoring   {...}          yaml:"monitoring"   ← NEW
           ├── Enabled      bool
           ├── Namespace    string
           ├── ChartVersion string
           ├── Grafana
           │    └── NodePort int
           ├── Prometheus
           │    └── Retention string
           └── Resources
                ├── Prometheus
                │    ├── CPU    string
                │    └── Memory string
                └── Grafana
                     ├── CPU    string
                     └── Memory string
```

### Structural Comparison

| Aspect | MetricsServer | Keda | MySQL | **Monitoring** |
|--------|---------------|------|-------|----------------|
| Config fields | 2 | 3 | 8+ | 9 |
| Namespace | Fixed (`kube-system`) | Configurable | Configurable | Configurable |
| NodePorts | None | None | 1 | 1 (Grafana) |
| Resources | None | None | Single pair | Two pairs (Prometheus, Grafana) |
| Secrets | None | None | Separate `secrets.mysql` | None |
| Persistence | None | None | Optional | None |
| Nesting depth | 1 | 1 | 3 | 3 |

The monitoring struct is moderately complex — more than Keda (which has three flat fields) but simpler than MySQL (which has persistence, init scripts, and a companion secrets block). It introduces no new patterns; the two-level `resources` nesting with separate sub-components is the only aspect that differs from existing components, but it follows the same anonymous struct + yaml tag convention used everywhere else.

### Dependencies

- **No dependency on other components.** Monitoring is self-contained.
- **Other components do not depend on monitoring.** Enabling or disabling monitoring has no effect on any other component.
- **Helm repository dependency:** `https://prometheus-community.github.io/helm-charts` must be registered during `kindenv init` (FR-012).
- **Port mapping dependency:** `grafana.nodePort` must be included in Kind cluster `extraPortMappings` (FR-015).

---

## Backward Compatibility

Adding the `Monitoring` struct to `Components` is a purely additive change:

- **Existing configs without a `monitoring:` key** — Go's YAML unmarshaler applies zero values (`enabled: false`), so the component is silently disabled. No behavior change.
- **Existing configs with unrelated fields** — Unaffected; the new struct occupies a new YAML key.
- **Default `enabled: false`** — Ensures no existing environment gains monitoring unless the developer explicitly opts in.