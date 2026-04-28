# Implementation Plan: Optional Monitoring Stack

**Branch**: `001-monitoring-stack` | **Date**: 2026-04-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-monitoring-stack/spec.md`

## Summary

Add an optional monitoring stack (Prometheus Operator + Grafana via the `kube-prometheus-stack` Helm chart) to the kindenv environment management system. The monitoring component follows the same patterns as existing components (MySQL, RabbitMQ, KEDA) — configured via `kindenv.yaml`, deployed via Helm, exposed via NodePort, with status reporting and graceful failure handling. When enabled, it provides metrics collection, pre-built Kubernetes dashboards, and auto-discovery of application metrics endpoints — all with zero-configuration defaults.

## Technical Context

**Language/Version**: Go (module: `github.com/devhelper/devhelper-cli`, see `go.mod`)
**Primary Dependencies**: Cobra CLI framework, `gopkg.in/yaml.v3` for config parsing, `os/exec` for Helm/kubectl commands
**Storage**: N/A (Prometheus uses `emptyDir` in-cluster; no host persistence)
**Testing**: Go `testing` package with table-driven tests; manual CLI integration tests against Kind cluster
**Target Platform**: macOS (arm64/amd64), Linux — local developer machines running Docker/Podman + Kind
**Project Type**: Single Go CLI binary
**Performance Goals**: Monitoring stack operational within 5 minutes of `kindenv start` (SC-001); Grafana dashboards populated within 2 minutes of stack ready (SC-002)
**Constraints**: Combined monitoring resource budget ≤ 1 CPU / 1 GB memory (SC-003); must not break any existing component (SC-004)
**Scale/Scope**: Single developer machine, single Kind cluster, single replica of each monitoring component

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Gate

| Gate | Status | Evidence |
|------|--------|----------|
| **I. Code Quality** — Go style, `go fmt`, `go vet`, `golangci-lint`, godoc, explicit error handling | ✅ Pass | Plan follows existing patterns in `kindenv_start.go`; all new code will pass linters and include godoc comments |
| **II. TDD** — Tests before implementation, ≥80% coverage, table-driven, unit + integration | ✅ Pass | Test files identified for each changed file; table-driven tests for config validation |
| **III. Cobra CLI** — Consistent command structure, flags, descriptions, examples | ✅ Pass | New `--skip-monitoring` flag follows `--skip-temporal`, `--skip-redis` pattern |
| **IV. Command Design** — Idempotent, verbose mode, confirmation prompts, exit codes | ✅ Pass | `helm upgrade --install` is idempotent; monitoring follows warn-and-continue pattern (FR-017) |
| **V. Error Handling** — Context + solutions, structured logging, color-coded output, progress indicators | ✅ Pass | Error messages include component name and suggested next steps; uses existing color helpers (`yellow()`, `red()`, `green()`) |
| **Cobra Compliance Checklist** — Naming, help text, flag validation, output formats, tests | ✅ Pass | All items addressed in contracts |

### Post-Design Gate (re-evaluated after Phase 1)

| Gate | Status | Notes |
|------|--------|-------|
| I. Code Quality | ✅ Pass | Struct follows anonymous nested pattern; no new packages needed |
| II. TDD | ✅ Pass | Test plan covers config parsing, validation, defaults, and CLI flag handling |
| III. Cobra CLI | ✅ Pass | `--skip-monitoring` is a `bool` flag with clear description |
| IV. Command Design | ✅ Pass | Idempotent via `helm upgrade --install`; graceful failure via warn-and-continue |
| V. Error Handling | ✅ Pass | Follows KEDA/MetricsServer pattern (continue on failure) not Redis/MySQL pattern (fail-fast) |

## Project Structure

### Documentation (this feature)

```text
specs/001-monitoring-stack/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: kube-prometheus-stack research
├── data-model.md        # Phase 1: Go struct and YAML config model
├── quickstart.md        # Phase 1: Developer quickstart guide
├── contracts/
│   └── cli-contract.md  # Phase 1: CLI interface contracts
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to modify)

```text
internal/kindenv/
└── config.go            # Add Monitoring struct, defaults, validation, port mappings

cmd/
├── kindenv_start.go     # Add --skip-monitoring flag + monitoring deployment block
├── kindenv_init.go      # Add prometheus-community Helm repo registration
├── kindenv_status.go    # Add monitoring status check block
├── kindenv_start_test.go  # Tests for monitoring deployment
├── kindenv_init_test.go   # Tests for Helm repo registration
├── kindenv_status_test.go # Tests for monitoring status output

internal/kindenv/
└── config_test.go       # Tests for monitoring config defaults, validation, YAML parsing
```

**Structure Decision**: This feature modifies the existing single-project structure. No new packages or directories are created — all changes are additions to existing files following the established component pattern.

## Phase 0: Research

**Output**: [research.md](./research.md) — all unknowns resolved.

### Decisions

| Decision | Choice | Rationale | Alternatives Considered |
|----------|--------|-----------|------------------------|
| Helm chart | `kube-prometheus-stack` from `prometheus-community` | Industry-standard chart bundling Prometheus Operator, Grafana, node-exporter, kube-state-metrics, and 30+ pre-built dashboards in a single install | Standalone Prometheus + separate Grafana charts (more config burden, no CRD-based discovery); Victoria Metrics (less ecosystem support) |
| Chart version | `72.6.2` (pinned stable) | Recent stable release with Prometheus Operator v0.79.x; well-tested with Kind | Latest (`83.4.0`) available but newer versions may have untested Kind quirks; pinning reduces surprise breakage |
| Grafana auth | Anonymous access, Admin role, login form disabled | Spec clarification mandates no login screen for local dev; consistent with OpenSearch Dashboards (security disabled) | Default credentials (admin/admin) — adds unnecessary friction for localhost-only access |
| Prometheus storage | `emptyDir` (no persistence) | Local dev assumption: 24h retention with no persistence is sufficient; avoids PVC complexity in Kind | PVC with `local-path` provisioner — adds storage complexity for minimal benefit in ephemeral dev clusters |
| ServiceMonitor discovery | `serviceMonitorSelectorNilUsesHelmValues: false` with empty selectors | Enables auto-discovery of ALL ServiceMonitors across ALL namespaces (FR-009) | Default (only discover own release's monitors) — would block custom app metric scraping |
| kube-proxy scraping | Disabled | kube-proxy in Kind binds to `127.0.0.1:10249`, making it unreachable from Prometheus pods | Enabled — would cause persistent scrape errors in logs |
| Alertmanager | Disabled | Explicitly out of scope per spec clarification | Enabled but unconfigured — wastes resources for unused functionality |
| Error handling | Warn and continue (FR-017) | Monitoring is optional; should not block other components | Fail-fast — would prevent deploying critical components (MySQL, Redis) when monitoring fails |
| Grafana NodePort | `31300` (host port `3000`) | Avoids collision with all existing NodePorts; `3000` is Grafana's conventional port | `30300` — less intuitive mapping; `31000` — too close to RabbitMQ Management `31672` |
| Deployment order | After KEDA, before Redis | Monitoring benefits from being deployed early so it can observe other component deployments; placed after infrastructure components (MetricsServer, KEDA) | Last in order — would miss observing other component startups |

## Phase 1: Design & Contracts

### Data Model

**Output**: [data-model.md](./data-model.md)

New anonymous struct added to `KindEnvConfig.Components`:

```go
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
```

**Defaults**: `enabled: false`, `namespace: "monitoring"`, `chartVersion: "72.6.2"`, `grafana.nodePort: 31300`, `prometheus.retention: "24h"`, resource limits totaling 700m CPU / 768Mi memory.

**Validation**: Namespace not empty, chart version not empty, NodePort in 30000–32767, CPU/memory format regex, retention format regex. All validation skipped when `enabled: false`.

### CLI Contracts

**Output**: [contracts/cli-contract.md](./contracts/cli-contract.md)

Summary of interface changes:

| Command | Change | Details |
|---------|--------|---------|
| `kindenv init` | New Helm repo | Registers `prometheus-community` at `https://prometheus-community.github.io/helm-charts` |
| `kindenv start` | New flag `--skip-monitoring` | Bool flag to skip monitoring even if enabled in config |
| `kindenv start` | New deployment block | `helm upgrade --install monitoring prometheus-community/kube-prometheus-stack` with ~25 `--set` flags for Grafana auth, NodePort, resources, retention, discovery, and disabled sub-components |
| `kindenv status` | New status section | Shows Prometheus, Grafana, and Node Exporter pod readiness when monitoring is enabled |
| `kindenv stop` | No change | Cluster deletion already removes all namespaces including monitoring |
| Config validation | New rules | 8 new validation rules for monitoring fields (see contracts) |

### Helm Values Strategy

The monitoring deployment uses `helm upgrade --install` with `--set` flags (not a values file) to stay consistent with how all other components (Redis, MySQL, RabbitMQ, KEDA) are deployed in `kindenv_start.go`. Key Helm values:

| Category | Key Helm Values |
|----------|----------------|
| **Disabled sub-components** | `alertmanager.enabled=false`, `thanosRuler.enabled=false`, `kubeProxy.enabled=false`, `windowsMonitoring.enabled=false` |
| **Grafana no-auth** | `grafana."grafana\.ini"."auth\.anonymous".enabled=true`, `grafana."grafana\.ini".auth.disable_login_form=true` |
| **Grafana exposure** | `grafana.service.type=NodePort`, `grafana.service.nodePort=<config>` |
| **Prometheus retention** | `prometheus.prometheusSpec.retention=<config>` |
| **Auto-discovery** | `prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false` (+ podMonitor, rule variants) |
| **Resources** | Per-component CPU/memory requests and limits from config |
| **Dashboards** | `grafana.defaultDashboardsEnabled=true`, sidecar enabled |

### Port Mapping

When monitoring is enabled, `generateDefaultPortMappings` adds:

```go
{ContainerPort: "${{ components.monitoring.grafana.nodePort }}", HostPort: 3000, Protocol: "TCP"}
```

This maps Grafana's NodePort (default `31300`) to `localhost:3000` on the host.

### Quickstart

**Output**: [quickstart.md](./quickstart.md) — developer-facing guide covering enable → init → start → access Grafana → monitor custom apps → customize.

## Implementation Sequence

The following is the recommended order of implementation, designed to enable incremental testing:

### Step 1: Configuration Model (`internal/kindenv/config.go`)

1. Add `Monitoring` struct to `Components` in `KindEnvConfig`
2. Add defaults in `LoadConfig()` function
3. Add defaults in `CreateDefaultConfig()` function
4. Add validation rules in `Validate()` function
5. Add port mapping in `generateDefaultPortMappings()` function
6. Add variable substitution support in `processVariableSubstitutions()` for `${{ components.monitoring.grafana.nodePort }}`

**Tests** (`internal/kindenv/config_test.go`):
- Default values are set correctly when no config file exists
- YAML with monitoring section parses correctly
- YAML without monitoring section defaults to `enabled: false`
- Validation passes with valid monitoring config
- Validation fails for invalid NodePort (out of range)
- Validation fails for invalid CPU/memory format
- Validation fails for empty namespace/chartVersion when enabled
- Validation skipped when monitoring is disabled
- Port mapping includes monitoring port when enabled
- Port mapping excludes monitoring port when disabled

### Step 2: Helm Repo Registration (`cmd/kindenv_init.go`)

1. Add `prometheus-community` Helm repo registration block (after existing repos, before `helm repo update`)
2. Follow existing pattern: attempt `helm repo add`, handle "already exists", warn on failure

**Tests** (`cmd/kindenv_init_test.go`):
- Repo registration command is called with correct name and URL
- "Already exists" output is handled gracefully
- Failure is reported as warning (non-blocking)

### Step 3: Monitoring Deployment (`cmd/kindenv_start.go`)

1. Add `--skip-monitoring` flag in `init()` function
2. Add flag handling to disable monitoring if `--skip-monitoring` is set
3. Add monitoring deployment block after KEDA and before Redis:
   - Create namespace (idempotent)
   - Build Helm args array with all `--set` flags
   - Execute `helm upgrade --install`
   - Handle error: print warning and continue (FR-017, following KEDA/MetricsServer pattern)
   - On success: print Grafana URL with host port

**Tests** (`cmd/kindenv_start_test.go`):
- Monitoring block is skipped when `enabled: false`
- Monitoring block is skipped when `--skip-monitoring` flag is set
- Helm command is constructed with correct args
- Namespace creation is attempted before Helm install
- Failure logs warning and does not exit
- Success prints Grafana URL

### Step 4: Status Reporting (`cmd/kindenv_status.go`)

1. Add monitoring status check block (after existing component checks):
   - Check pods in monitoring namespace via `kubectl get pods -n <namespace> --no-headers`
   - Parse output for Prometheus, Grafana, and Node Exporter pod readiness
   - Display appropriate status emoji (✅ / ⚠️) with pod counts

**Tests** (`cmd/kindenv_status_test.go`):
- Status section shown when monitoring is enabled
- Status section hidden when monitoring is disabled
- Healthy status displays green checkmark
- Degraded status displays warning with details
- Missing pods displays "not running or not installed"

### Step 5: Documentation Updates

1. Update `README.md` — add Monitoring to Supported Components section and kindenv configuration examples
2. Update `KINDENV.md` — add monitoring configuration section with all options
3. Update `CUSTOM_COMPONENTS.md` — add note about ServiceMonitor CRDs for custom app metrics

## Complexity Tracking

No constitution violations to justify. This feature:
- Adds one new component following existing patterns (no new abstractions)
- Modifies 4 existing Go files + their test files
- Introduces no new dependencies (uses existing `os/exec` + `gopkg.in/yaml.v3`)
- Requires no new packages or directories in the source tree

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| kube-prometheus-stack CRDs take long to install in Kind | Medium | Medium | Use `--wait --timeout 5m` on Helm install; warn-and-continue if timeout exceeded |
| Helm `--set` with escaped dots for `grafana.ini` keys fails | Medium | High | Test exact `--set` syntax; fall back to `--values` with temp file if needed |
| NodePort 31300 conflicts with user's existing port mappings | Low | Low | Configurable via `kindenv.yaml`; document default in quickstart |
| Chart version 72.6.2 unavailable in future | Low | Low | Configurable via `chartVersion`; document how to find latest version |
| Prometheus resource usage exceeds 1 CPU / 1 GB on active clusters | Low | Medium | Default resource limits enforce ceiling; document in quickstart |