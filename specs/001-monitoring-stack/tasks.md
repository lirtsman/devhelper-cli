# Tasks: Optional Monitoring Stack

**Input**: Design documents from `/specs/001-monitoring-stack/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: Included — constitution mandates TDD (Section II, NON-NEGOTIABLE).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go source**: `internal/kindenv/`, `cmd/` at repository root
- **Tests**: Co-located with source (`*_test.go` files)
- **Docs**: `README.md`, `KINDENV.md`, `CUSTOM_COMPONENTS.md` at repository root

---

## Phase 1: Setup

**Purpose**: No new project scaffolding needed — this feature modifies existing files only. This phase verifies prerequisites and prepares the branch.

- [X] T001 Verify branch `001-monitoring-stack` is checked out and all design artifacts exist in `specs/001-monitoring-stack/`
- [X] T002 Run `go build ./...` to confirm clean baseline compilation before any changes

---

## Phase 2: Foundational (Configuration Model)

**Purpose**: Add the Monitoring struct, defaults, validation, and port mappings to the config model. This is the shared foundation that ALL user stories depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests for Foundation

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T003 [P] Add test `TestLoadConfig_MonitoringDefaults` verifying default values (enabled=false, namespace="monitoring", chartVersion="72.6.2", grafana.nodePort=31300, prometheus.retention="24h", resource defaults) in `internal/kindenv/config_test.go`
- [X] T004 [P] Add test `TestLoadConfig_MonitoringFromYAML` verifying YAML with monitoring section parses correctly, and YAML without monitoring section defaults to enabled=false in `internal/kindenv/config_test.go`
- [X] T005 [P] Add test `TestValidate_MonitoringEnabled` with table-driven subtests covering: valid config passes, invalid NodePort (out of range) fails, invalid CPU format fails, invalid memory format fails, empty namespace fails, empty chartVersion fails, invalid retention format fails in `internal/kindenv/config_test.go`
- [X] T006 [P] Add test `TestValidate_MonitoringDisabled` verifying all validation is skipped when monitoring.enabled=false in `internal/kindenv/config_test.go`
- [X] T007 [P] Add test `TestGenerateDefaultPortMappings_Monitoring` verifying port mapping includes Grafana nodePort when monitoring enabled, and excludes it when disabled in `internal/kindenv/config_test.go`

### Implementation for Foundation

- [X] T008 Add `Monitoring` anonymous struct to `Components` inside `KindEnvConfig` struct (after `Keda`) with fields: Enabled, Namespace, ChartVersion, Grafana.NodePort, Prometheus.Retention, Resources.Prometheus.CPU/Memory, Resources.Grafana.CPU/Memory — all with yaml tags per data-model.md in `internal/kindenv/config.go`
- [X] T009 Add monitoring default values in `LoadConfig()` function (enabled=false, namespace="monitoring", chartVersion="72.6.2", grafana.nodePort=31300, prometheus.retention="24h", resources per data-model.md) in `internal/kindenv/config.go`
- [X] T010 Add identical monitoring default values in `CreateDefaultConfig()` function in `internal/kindenv/config.go`
- [X] T011 Add monitoring validation rules in `Validate()` function (guarded by `if c.Components.Monitoring.Enabled`): namespace not empty, chartVersion not empty, NodePort in 30000-32767, CPU/memory format regex, retention format regex per contracts/cli-contract.md in `internal/kindenv/config.go`
- [X] T012 Add Grafana port mapping entry in `generateDefaultPortMappings()` when monitoring is enabled: ContainerPort `${{ components.monitoring.grafana.nodePort }}`, HostPort 3000, Protocol TCP in `internal/kindenv/config.go`
- [X] T013 Add variable substitution support for `${{ components.monitoring.grafana.nodePort }}` in `processVariableSubstitutions()` function in `internal/kindenv/config.go`
- [X] T014 Run `go test ./internal/kindenv/ -run TestLoadConfig_Monitoring -v && go test ./internal/kindenv/ -run TestValidate_Monitoring -v && go test ./internal/kindenv/ -run TestGenerateDefaultPortMappings_Monitoring -v` to verify all foundation tests pass

**Checkpoint**: Config model complete — `go build ./...` passes, all config tests green. User story implementation can now begin.

---

## Phase 3: User Story 1 — Enable Monitoring in Local Environment (Priority: P1) 🎯 MVP

**Goal**: A developer can set `components.monitoring.enabled: true` in `kindenv.yaml`, run `devhelper-cli kindenv init` + `devhelper-cli kindenv start`, and the full monitoring stack (Prometheus Operator + Grafana) is deployed to the Kind cluster. When disabled, nothing is deployed.

**Independent Test**: Enable monitoring in config → `kindenv init` → `kindenv start` → verify monitoring pods are running in the `monitoring` namespace → open `http://localhost:3000` in a browser and see Grafana dashboard.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T015 [P] [US1] Add test for prometheus-community Helm repo registration in `kindenv init`: verify `helm repo add prometheus-community https://prometheus-community.github.io/helm-charts` is called, and "already exists" is handled gracefully in `cmd/kindenv_init_test.go`
- [X] T016 [P] [US1] Add test `TestStartMonitoring_Enabled` verifying that when monitoring is enabled, the deployment block creates namespace and executes `helm upgrade --install monitoring prometheus-community/kube-prometheus-stack` with correct `--namespace`, `--version`, and ALL `--set` flags in `cmd/kindenv_start_test.go`. The test MUST exhaustively assert the presence of: (1) disabled sub-components: `alertmanager.enabled=false`, `thanosRuler.enabled=false`, `kubeProxy.enabled=false`, `windowsMonitoring.enabled=false`; (2) Grafana auth: `grafana.enabled=true`, `grafana."grafana\.ini"."auth\.anonymous".enabled=true`, `grafana."grafana\.ini"."auth\.anonymous".org_role=Admin`, `grafana."grafana\.ini".auth.disable_login_form=true`, `grafana."grafana\.ini".security.allow_embedding=true`; (3) Grafana service: `grafana.service.type=NodePort`, `grafana.service.nodePort=<config>`; (4) Grafana dashboards: `grafana.defaultDashboardsEnabled=true`, `grafana.persistence.enabled=false`, `grafana.sidecar.dashboards.enabled=true`, `grafana.sidecar.dashboards.searchNamespace=ALL`, `grafana.sidecar.datasources.enabled=true`, `grafana.sidecar.datasources.defaultDatasourceEnabled=true`; (5) Prometheus: `prometheus.prometheusSpec.retention=<config>`, `prometheus.prometheusSpec.storageSpec.emptyDir.medium=""`, `prometheus.service.type=ClusterIP`; (6) resources for both Prometheus and Grafana (requests + limits from config); (7) auto-discovery: `prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false`, `serviceMonitorSelector=`, `serviceMonitorNamespaceSelector=`, `podMonitorSelectorNilUsesHelmValues=false`, `podMonitorSelector=`, `podMonitorNamespaceSelector=`, `ruleSelectorNilUsesHelmValues=false`, `ruleSelector=`, `ruleNamespaceSelector=`; (8) infrastructure: `prometheusOperator.enabled=true`, `nodeExporter.enabled=true`, `kubeStateMetrics.enabled=true`, `defaultRules.create=true`
- [X] T017 [P] [US1] Add test `TestStartMonitoring_Disabled` verifying that when monitoring is disabled (enabled=false), no monitoring-related Helm commands or namespace creation are executed in `cmd/kindenv_start_test.go`
- [X] T018 [P] [US1] Add test `TestStartMonitoring_SkipFlag` verifying that `--skip-monitoring` flag causes monitoring deployment to be skipped even when enabled=true, and prints skip message in `cmd/kindenv_start_test.go`
- [X] T019 [P] [US1] Add test `TestStartMonitoring_FailureWarnsAndContinues` verifying that when Helm install fails, error is logged as warning with ❌ prefix, "Continuing despite" message is printed, and the function does NOT call os.Exit in `cmd/kindenv_start_test.go`

### Implementation for User Story 1

- [X] T020 [US1] Add `prometheus-community` Helm repository registration block in `cmd/kindenv_init.go` — insert after existing repo blocks (kedacore), before `helm repo update`. Follow existing pattern: `helm repo add prometheus-community https://prometheus-community.github.io/helm-charts`, handle "already exists" with ✅ message, warn on failure with ⚠️
- [X] T021 [US1] Register `--skip-monitoring` bool flag in `init()` function in `cmd/kindenv_start.go` — follow the pattern of existing `--skip-temporal`, `--skip-dapr`, `--skip-redis` flags
- [X] T022 [US1] Add flag handling for `--skip-monitoring` in the `Run` function (after existing skip-flag handling blocks) to set `config.Components.Monitoring.Enabled = false` when flag is true in `cmd/kindenv_start.go`
- [X] T023 [US1] Add monitoring deployment block in `cmd/kindenv_start.go` — insert after KEDA block and before Redis block. Implementation must: (1) check `config.Components.Monitoring.Enabled`, (2) print yellow "Installing Monitoring Stack (kube-prometheus-stack)", (3) create namespace idempotently via `kubectl create namespace --dry-run=client -o yaml | kubectl apply -f -`, (4) build `helmArgs` slice with `upgrade --install monitoring prometheus-community/kube-prometheus-stack` plus all `--set` flags per contracts/cli-contract.md Helm Command Contract, (5) execute via `executeCommand("helm", helmArgs...)`, (6) on error: print ❌ error + yellow "Continuing despite Monitoring Stack installation failure...", (7) on success: print ✅ success with Grafana URL `http://localhost:3000`
- [X] T024 [US1] Run all US1 tests: `go test ./cmd/ -run TestStartMonitoring -v && go test ./cmd/ -run "TestInit.*prometheus" -v` to verify all pass

**Checkpoint**: User Story 1 complete — `devhelper-cli kindenv init` registers prometheus-community repo, `devhelper-cli kindenv start` deploys monitoring stack when enabled, warns and continues on failure, skips when disabled or `--skip-monitoring` is set.

---

## Phase 4: User Story 2 — View Application Metrics on Pre-Built Dashboards (Priority: P2)

**Goal**: When the monitoring stack is deployed, 30+ pre-built Grafana dashboards for Kubernetes cluster overview, node metrics, and pod resource usage are available and populated with live data — with zero additional configuration.

**Independent Test**: Deploy monitoring stack → open Grafana at `http://localhost:3000` → navigate to Dashboards → verify "Kubernetes / Compute Resources / Cluster", "Kubernetes / Compute Resources / Namespace (Pods)", and "Node Exporter / Nodes" dashboards exist and show live data.

### Implementation for User Story 2

> **Note**: Pre-built dashboards are enabled by the Helm `--set` flags implemented in US1. T016 now exhaustively asserts ALL `--set` flags (including dashboard sidecar, nodeExporter, kubeStateMetrics, kubeProxy, windowsMonitoring). This phase is a code-review verification pass — no new test tasks are needed because T016 already covers every flag listed below.

- [X] T025 [US2] Verify the Helm args in the monitoring deployment block in `cmd/kindenv_start.go` include: `--set grafana.defaultDashboardsEnabled=true`, `--set defaultRules.create=true`, `--set grafana.sidecar.dashboards.enabled=true`, `--set grafana.sidecar.dashboards.searchNamespace=ALL`, `--set grafana.sidecar.datasources.enabled=true`, `--set grafana.sidecar.datasources.defaultDatasourceEnabled=true`. Add any that are missing — T016 will catch omissions.
- [X] T026 [US2] Verify the Helm args include `--set nodeExporter.enabled=true` and `--set kubeStateMetrics.enabled=true` to ensure node and pod-level metrics are collected for the pre-built dashboards. Add if missing — T016 will catch omissions.
- [X] T027 [US2] Verify the Helm args include `--set kubeProxy.enabled=false` and `--set windowsMonitoring.enabled=false` per Kind-specific research findings. Add if missing — T016 will catch omissions.

**Checkpoint**: User Story 2 complete — Grafana ships with 30+ pre-built dashboards covering cluster, node, pod, and namespace metrics, all auto-populated via sidecar discovery.

---

## Phase 5: User Story 3 — Configure Monitoring Stack Settings (Priority: P2)

**Goal**: A developer can customize the monitoring stack by changing `kindenv.yaml` values (NodePort, retention, resource limits) and re-running `kindenv start` to upgrade the deployment in place.

**Independent Test**: Deploy with defaults → change `grafana.nodePort` to 31400 and `prometheus.retention` to 48h in `kindenv.yaml` → re-run `kindenv start` → verify Grafana is now on the new port and Prometheus retention is updated.

### Tests for User Story 3

- [X] T028 [P] [US3] Add test `TestStartMonitoring_CustomConfig` verifying that non-default config values (custom nodePort, retention, resource CPU/memory) are correctly passed into the Helm args array in `cmd/kindenv_start_test.go`
- [X] T029 [P] [US3] Add test `TestStartMonitoring_UpgradeInPlace` verifying that `helm upgrade --install` is used (not `helm install`), confirming idempotent upgrade behavior for changed config values in `cmd/kindenv_start_test.go`

### Implementation for User Story 3

- [X] T030 [US3] Verify the monitoring deployment block in `cmd/kindenv_start.go` reads all configurable values from the config struct (not hardcoded): `config.Components.Monitoring.Namespace`, `config.Components.Monitoring.ChartVersion`, `config.Components.Monitoring.Grafana.NodePort`, `config.Components.Monitoring.Prometheus.Retention`, `config.Components.Monitoring.Resources.Prometheus.CPU`, `config.Components.Monitoring.Resources.Prometheus.Memory`, `config.Components.Monitoring.Resources.Grafana.CPU`, `config.Components.Monitoring.Resources.Grafana.Memory`. Fix any hardcoded values.
- [X] T031 [US3] Run US3 tests: `go test ./cmd/ -run "TestStartMonitoring_Custom|TestStartMonitoring_Upgrade" -v` to verify all pass

**Checkpoint**: User Story 3 complete — all monitoring settings are configurable via `kindenv.yaml` and applied via idempotent `helm upgrade --install`.

---

## Phase 6: User Story 4 — Monitor Custom Application Metrics (Priority: P3)

**Goal**: The monitoring stack automatically discovers and scrapes metrics from application pods that expose a `/metrics` endpoint via ServiceMonitor CRDs, making custom metrics available in Grafana.

**Independent Test**: Deploy monitoring stack → deploy a sample app with a ServiceMonitor CRD → wait 3 minutes → query Prometheus via Grafana for the custom metric → verify data is returned.

### Implementation for User Story 4

> **Note**: Auto-discovery is enabled by Helm `--set` flags implemented in US1. T016 now exhaustively asserts all discovery-related flags. This phase is a code-review verification pass — no new test tasks are needed because T016 already covers every flag listed below.

- [X] T032 [US4] Verify the Helm args in `cmd/kindenv_start.go` include: `--set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false`, `--set prometheus.prometheusSpec.serviceMonitorSelector=`, `--set prometheus.prometheusSpec.serviceMonitorNamespaceSelector=`, `--set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false`, `--set prometheus.prometheusSpec.podMonitorSelector=`, `--set prometheus.prometheusSpec.podMonitorNamespaceSelector=`. Add any that are missing — T016 will catch omissions.
- [X] T033 [US4] Add a note in `CUSTOM_COMPONENTS.md` under a new subsection "### Monitoring Custom Application Metrics" explaining how to create a ServiceMonitor CRD for custom components that expose a `/metrics` endpoint, with a YAML example per quickstart.md

**Checkpoint**: User Story 4 complete — ServiceMonitor auto-discovery works across all namespaces, documented for custom component users.

---

## Phase 7: User Story 5 — View Monitoring Stack Status (Priority: P3)

**Goal**: The `devhelper-cli kindenv status` command shows the health of Prometheus, Grafana, and Node Exporter pods when monitoring is enabled. No monitoring output when disabled.

**Independent Test**: Enable monitoring → `kindenv start` → `kindenv status` → verify output includes monitoring section with pod readiness counts and Grafana URL.

### Tests for User Story 5

- [X] T034 [P] [US5] Add test `TestStatus_MonitoringEnabled` verifying that when monitoring is enabled and pods are running, status output includes "Monitoring Stack is installed and running" with Prometheus, Grafana, and Node Exporter pod counts in `cmd/kindenv_status_test.go`
- [X] T035 [P] [US5] Add test `TestStatus_MonitoringDisabled` verifying that when monitoring is disabled, no monitoring-related output is produced in `cmd/kindenv_status_test.go`
- [X] T036 [P] [US5] Add test `TestStatus_MonitoringDegraded` verifying that when some monitoring pods are not ready, status shows ⚠️ warning with specific pod status details in `cmd/kindenv_status_test.go`

### Implementation for User Story 5

- [X] T037 [US5] Add monitoring status check block in `cmd/kindenv_status.go` — insert after existing component status checks. Implementation must: (1) check `config.Components.Monitoring.Enabled`, (2) run `kubectl get pods -n <namespace> --no-headers -l "app.kubernetes.io/instance=monitoring"`, (3) parse output to count ready/total pods for Prometheus, Grafana, and Node Exporter, (4) display ✅ with pod counts and Grafana URL when all healthy, ⚠️ with details when degraded, or ⚠️ "not running or not installed" when no pods found — per contracts/cli-contract.md Section 4
- [X] T038 [US5] Run US5 tests: `go test ./cmd/ -run TestStatus_Monitoring -v` to verify all pass

**Checkpoint**: User Story 5 complete — `kindenv status` reports monitoring health with color-coded output.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates, final validation, and cleanup across all user stories.

- [X] T039 [P] Add "### Monitoring" subsection to the "## Supported Components" section in `README.md` — describe Prometheus Operator + Grafana monitoring stack with metrics collection, pre-built dashboards, and auto-discovery of application metrics
- [X] T040 [P] Add monitoring component configuration section to `KINDENV.md` — document all configuration options (enabled, namespace, chartVersion, grafana.nodePort, prometheus.retention, resources) with YAML example per quickstart.md
- [X] T041 [P] Add monitoring configuration example to the `kindenv init` default config section and `kindenv start` components section in `README.md` — show the monitoring YAML block alongside existing component examples
- [X] T042 Run full test suite: `go test ./... -v` to verify no regressions across all packages
- [ ] T046 Verify FR-011: with monitoring enabled, run `devhelper-cli kindenv start` then `devhelper-cli kindenv stop --delete`. Confirm the `monitoring` namespace and all monitoring CRDs (prometheuses.monitoring.coreos.com, servicemonitors.monitoring.coreos.com, etc.) are fully removed by the cluster deletion
- [ ] T047 Verify FR-018: with monitoring previously deployed, set `components.monitoring.enabled: false` in `kindenv.yaml` and re-run `devhelper-cli kindenv start`. Confirm no monitoring-related Helm commands are executed, no errors are produced, and previously deployed monitoring pods remain running until `kindenv stop --delete`
- [X] T043 Run `go vet ./...` and `golangci-lint run` to verify code quality compliance with constitution
- [X] T044 Run `go build -o devhelper-cli .` to verify clean binary build
- [ ] T045 Validate quickstart.md flow end-to-end: add monitoring config to `kindenv.yaml`, run `devhelper-cli kindenv init`, run `devhelper-cli kindenv start`, verify Grafana accessible at `http://localhost:3000`, verify `devhelper-cli kindenv status` shows monitoring health

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — MVP, BLOCKS US2/US3/US4 (they verify/extend US1's Helm args)
- **US2 (Phase 4)**: Depends on US1 (verifies Helm values set in US1)
- **US3 (Phase 5)**: Depends on US1 (extends config handling in US1's deployment block)
- **US4 (Phase 6)**: Depends on US1 (verifies Helm values set in US1)
- **US5 (Phase 7)**: Depends on Phase 2 only — can run in parallel with US1-US4 (different file: `kindenv_status.go`)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational only — core MVP
- **US2 (P2)**: Depends on US1 — verifies/extends Helm dashboard values in `kindenv_start.go`
- **US3 (P2)**: Depends on US1 — verifies/extends config handling in `kindenv_start.go`
- **US4 (P3)**: Depends on US1 — verifies Helm discovery values in `kindenv_start.go`, adds docs in `CUSTOM_COMPONENTS.md`
- **US5 (P3)**: Independent of US1-US4 — works in `kindenv_status.go` (different file)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Config model before CLI commands
- Core implementation before verification/documentation
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 2**: T003–T007 (all config tests) can run in parallel
- **Phase 3**: T015–T019 (all US1 tests) can run in parallel
- **Phase 5 + Phase 7**: US3 and US5 can run in parallel (different files: `kindenv_start.go` vs `kindenv_status.go`)
- **Phase 6 + Phase 7**: US4 and US5 can run in parallel (different files: `CUSTOM_COMPONENTS.md` vs `kindenv_status.go`)
- **Phase 8**: T039–T041 (all documentation tasks) can run in parallel; T046 and T047 (FR-011/FR-018 verification) can run in parallel with documentation tasks

---

## Parallel Example: Foundation Phase

```
# Launch all foundation config tests in parallel (different test functions, same file):
Task T003: "TestLoadConfig_MonitoringDefaults in internal/kindenv/config_test.go"
Task T004: "TestLoadConfig_MonitoringFromYAML in internal/kindenv/config_test.go"
Task T005: "TestValidate_MonitoringEnabled in internal/kindenv/config_test.go"
Task T006: "TestValidate_MonitoringDisabled in internal/kindenv/config_test.go"
Task T007: "TestGenerateDefaultPortMappings_Monitoring in internal/kindenv/config_test.go"
```

## Parallel Example: User Story 1

```
# Launch all US1 tests in parallel (different test functions, same file):
Task T015: "Test Helm repo registration in cmd/kindenv_init_test.go"
Task T016: "TestStartMonitoring_Enabled in cmd/kindenv_start_test.go"
Task T017: "TestStartMonitoring_Disabled in cmd/kindenv_start_test.go"
Task T018: "TestStartMonitoring_SkipFlag in cmd/kindenv_start_test.go"
Task T019: "TestStartMonitoring_FailureWarnsAndContinues in cmd/kindenv_start_test.go"
```

## Parallel Example: US5 alongside US3

```
# These touch different files and can run simultaneously:
Developer A (kindenv_start.go):
  Task T028-T031: US3 — config customization tests and verification

Developer B (kindenv_status.go):
  Task T034-T038: US5 — status reporting tests and implementation
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T002)
2. Complete Phase 2: Foundational config model (T003–T014)
3. Complete Phase 3: User Story 1 — deploy monitoring (T015–T024)
4. **STOP and VALIDATE**: `kindenv init` + `kindenv start` with monitoring enabled → Grafana at `localhost:3000`
5. This is a fully functional monitoring stack — demo/deploy ready

### Incremental Delivery

1. Setup + Foundational → Config model ready, `go build` passes
2. Add US1 → Core deployment works → **MVP!**
3. Add US2 → Pre-built dashboards verified → Enhanced value
4. Add US3 → Full configurability → Power users supported
5. Add US4 → Custom metrics discovery → Advanced use case
6. Add US5 → Status reporting → Complete observability
7. Polish → Docs, lint, full validation → Release ready

### Single Developer Strategy (Recommended)

Sequential phases in priority order:

1. Phase 2: Foundation (T003–T014) — ~2 hours
2. Phase 3: US1 MVP (T015–T024) — ~3 hours
3. Phase 4: US2 dashboards (T025–T027) — ~30 minutes
4. Phase 5: US3 config (T028–T031) — ~1 hour
5. Phase 7: US5 status (T034–T038) — ~1.5 hours
6. Phase 6: US4 custom metrics (T032–T033) — ~30 minutes
7. Phase 8: Polish (T039–T045) — ~1.5 hours

**Estimated total**: ~10 hours

---

## Notes

- [P] tasks = different files or independent test functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable at its checkpoint
- Constitution requires TDD: write tests first, verify they fail, then implement
- Commit after each phase checkpoint for clean git history
- All Helm `--set` flag values reference contracts/cli-contract.md as the source of truth
- T016 exhaustively asserts ALL `--set` flags, providing test coverage for US2 (Phase 4) and US4 (Phase 6) verification tasks without requiring separate test tasks in those phases
- The `grafana.ini` nested keys may require `--values` file approach if `--set` escaping proves problematic (see Risk Assessment in plan.md)
- Total tasks: 47 (T001–T047)