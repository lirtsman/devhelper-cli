# Feature Specification: Optional Monitoring Stack

**Feature Branch**: `001-monitoring-stack`  
**Created**: 2026-04-10  
**Status**: Draft  
**Input**: User description: "optional monitoring stack (prometheus-operator with grafana setup)"

## Clarifications

### Session 2026-04-10

- Q: When monitoring config changes in `kindenv.yaml` and the stack is already deployed, what should `kindenv start` do? → A: Upgrade in place — apply changed values to the existing deployment (like `helm upgrade`).
- Q: When the monitoring stack fails to deploy (insufficient resources, port conflict, Helm error), how should `kindenv start` behave? → A: Warn and continue — log a warning about the monitoring failure but proceed with deploying all other components.
- Q: How should the monitoring dashboard handle authentication for local development access? → A: No authentication — dashboard is open with no login screen.
- Q: Should log aggregation, distributed tracing, and alerting be explicitly declared out of scope? → A: Exclude all three — this feature covers metrics collection and dashboards only.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enable Monitoring in Local Environment (Priority: P1)

A developer working on a microservices application in the Kind-based local environment wants to observe resource usage (CPU, memory) and application metrics for the services running in the cluster. They enable the monitoring component in their `kindenv.yaml` configuration, run `devhelper-cli kindenv start`, and the monitoring stack is automatically deployed alongside their other components. They can then open a browser-based dashboard to view cluster and application metrics without any additional manual setup.

**Why this priority**: Monitoring is the core value proposition of this feature. Without the ability to enable and deploy the stack, nothing else matters. This delivers immediate observability value to any developer using the local environment.

**Independent Test**: Can be fully tested by enabling the monitoring component in configuration, starting the environment, and verifying that the monitoring services are running and accessible via their designated ports.

**Acceptance Scenarios**:

1. **Given** a `kindenv.yaml` with `components.monitoring.enabled: true`, **When** the developer runs `devhelper-cli kindenv start`, **Then** the monitoring stack (metrics collection and dashboards) is deployed to the cluster and all monitoring pods reach a ready state.
2. **Given** the monitoring stack is deployed, **When** the developer opens the dashboard URL in a browser, **Then** they see a pre-configured dashboard showing cluster-level metrics (node CPU, memory, pod counts).
3. **Given** a `kindenv.yaml` with `components.monitoring.enabled: false` (or omitted), **When** the developer runs `devhelper-cli kindenv start`, **Then** no monitoring components are deployed and no monitoring-related resources exist in the cluster.

---

### User Story 2 - View Application Metrics on Pre-Built Dashboards (Priority: P2)

A developer has application services running in the local Kind cluster that expose standard metrics endpoints. After enabling the monitoring stack, they want to see pre-built dashboards that visualize key Kubernetes and infrastructure metrics (node health, pod resource usage, namespace breakdowns) without manually creating dashboard configurations.

**Why this priority**: Pre-built dashboards dramatically reduce time-to-value. Without them, developers would need to manually create visualizations, which undermines the "zero-friction" philosophy of devhelper-cli.

**Independent Test**: Can be tested by deploying the monitoring stack and verifying that pre-built dashboards for Kubernetes cluster metrics are available and populated with data in the dashboard UI.

**Acceptance Scenarios**:

1. **Given** the monitoring stack is deployed and running, **When** the developer opens the dashboard UI, **Then** pre-built dashboards for Kubernetes cluster overview, node metrics, and pod resource usage are available and displaying live data.
2. **Given** the monitoring stack is deployed and an application pod is running, **When** the developer navigates to the pod-level dashboard, **Then** they see CPU, memory, and network metrics for that specific pod.

---

### User Story 3 - Configure Monitoring Stack Settings (Priority: P2)

A developer wants to customize the monitoring stack to suit their specific needs — for example, changing the dashboard access port, adjusting data retention duration, or configuring resource limits to fit their machine's capacity. They update the relevant fields in `kindenv.yaml` and restart the environment.

**Why this priority**: Customization is important for fitting the monitoring stack into varied developer machine configurations and workflows, but reasonable defaults should cover most use cases.

**Independent Test**: Can be tested by modifying monitoring configuration values in `kindenv.yaml`, restarting the environment, and verifying the changes are reflected in the deployed stack.

**Acceptance Scenarios**:

1. **Given** a `kindenv.yaml` with custom node port values for the monitoring dashboard, **When** the developer starts the environment, **Then** the dashboard is accessible on the configured port.
2. **Given** a `kindenv.yaml` with custom resource limits for the monitoring components, **When** the environment is started, **Then** the monitoring pods are created with the specified resource requests and limits.
3. **Given** a `kindenv.yaml` with a custom data retention period, **When** the environment is started, **Then** the metrics collection service is configured with the specified retention duration.
4. **Given** the monitoring stack is already deployed with default settings, **When** the developer changes monitoring configuration values in `kindenv.yaml` and re-runs `devhelper-cli kindenv start`, **Then** the existing monitoring deployment is upgraded in place with the new values without requiring a full tear-down and redeploy.

---

### User Story 4 - Monitor Custom Application Metrics (Priority: P3)

A developer has instrumented their application to expose custom metrics (e.g., request latency, queue depth, error rates) via a standard metrics endpoint. They want the monitoring stack to automatically discover and scrape these custom metrics from their application pods running in the cluster, so they can visualize them on the dashboard.

**Why this priority**: Custom application metrics provide deeper insight beyond infrastructure-level monitoring, but this is an advanced use case that builds on top of the base monitoring capability.

**Independent Test**: Can be tested by deploying a sample application that exposes custom metrics, verifying that the metrics collector discovers and scrapes the endpoint, and confirming the custom metrics appear in the dashboard query interface.

**Acceptance Scenarios**:

1. **Given** the monitoring stack is deployed and a custom component has a standard metrics endpoint annotated for scraping, **When** the metrics collector runs its discovery cycle, **Then** the custom application metrics are scraped and available for querying in the dashboard.
2. **Given** custom application metrics are being collected, **When** the developer creates a new dashboard panel using a custom metric name, **Then** the data is displayed correctly.

---

### User Story 5 - View Monitoring Stack Status (Priority: P3)

A developer wants to quickly verify whether the monitoring stack is healthy and functioning correctly. When they run the `devhelper-cli kindenv status` command, monitoring component status is included in the output.

**Why this priority**: Status visibility is important for troubleshooting but is a secondary concern once the core deployment works.

**Independent Test**: Can be tested by running `devhelper-cli kindenv status` with the monitoring stack enabled and verifying monitoring-related component statuses appear in the output.

**Acceptance Scenarios**:

1. **Given** the monitoring stack is deployed and healthy, **When** the developer runs `devhelper-cli kindenv status`, **Then** the output includes the status of monitoring components (metrics collector and dashboard) showing them as running/ready.
2. **Given** the monitoring stack is not enabled, **When** the developer runs `devhelper-cli kindenv status`, **Then** no monitoring-related status information is displayed.

---

### Out of Scope

The following observability concerns are explicitly excluded from this feature and may be considered as separate future features:

- **Log aggregation** — centralized collection, storage, and search of application and system logs.
- **Distributed tracing** — request-level tracing across microservices (e.g., trace propagation, span visualization).
- **Alerting and notifications** — alert rule definitions, threshold-based triggers, and notification channel integrations (email, Slack, PagerDuty, etc.).

This feature is scoped strictly to **metrics collection and dashboard visualization**.

### Edge Cases

- What happens when the monitoring stack is enabled but the cluster has insufficient resources (CPU/memory) to run it? → The system logs a warning and continues deploying other components.
- How does the system handle a scenario where the monitoring component's designated node ports conflict with ports already in use by other components? → The system logs a warning and continues deploying other components; the monitoring stack may be in a degraded or failed state.
- What happens when the developer disables monitoring after it was previously enabled and deployed — are all monitoring resources properly cleaned up? → When `enabled: false` and `kindenv start` is re-run, the system skips the monitoring deployment block. Previously deployed monitoring pods remain running until `kindenv stop --delete` removes the entire cluster. This is consistent with how other components behave — disabling a component prevents new deployments but does not actively uninstall existing ones.
- How does the system behave if the metrics collection service starts before any application pods are running (no targets to scrape)? → The metrics collector operates normally with zero scrape targets. It will automatically discover and begin scraping application pods as they become available. No errors or warnings are produced for an empty target list.
- What happens if the dashboard data storage fills up due to a long-running environment with high metric cardinality? → With the default 24-hour retention and `emptyDir` storage, the metrics collector automatically expires data older than the retention window. Combined with the SC-003 resource limits, disk pressure is unlikely in typical local development usage. If the node disk does fill, Kubernetes will evict the metrics collector pod, which can be restarted by re-running `kindenv start`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST support an optional monitoring component in the `kindenv.yaml` configuration that can be enabled or disabled via an `enabled` flag, consistent with all other existing components.
- **FR-002**: When monitoring is enabled, the system MUST deploy a metrics collection service capable of scraping and storing time-series metrics from cluster nodes, Kubernetes system components, and application pods.
- **FR-003**: When monitoring is enabled, the system MUST deploy a dashboard service that provides a browser-accessible UI for visualizing collected metrics.
- **FR-004**: The system MUST provide pre-built dashboards for Kubernetes cluster overview, node-level metrics, and pod-level resource usage out of the box.
- **FR-005**: The monitoring dashboard MUST be accessible from the developer's host machine via a configurable Kubernetes NodePort, consistent with how other component services are exposed.
- **FR-006**: The system MUST allow configuration of resource limits (CPU, memory) for the monitoring stack components to accommodate varied developer machine capacities.
- **FR-007**: The system MUST allow configuration of metrics data retention duration, with a sensible default appropriate for local development (e.g., 24 hours).
- **FR-008**: The system MUST deploy monitoring components into a dedicated namespace to maintain clean separation from application workloads, consistent with other components like Temporal and Redis.
- **FR-009**: The system MUST automatically discover and scrape metrics from application pods that expose a standard metrics endpoint and are annotated for scraping.
- **FR-010**: The `devhelper-cli kindenv status` command MUST include the health status of the monitoring components when monitoring is enabled.
- **FR-011**: The `devhelper-cli kindenv stop --delete` command MUST cleanly remove all monitoring-related resources when the cluster is deleted.
- **FR-016**: When monitoring is already deployed and the developer re-runs `devhelper-cli kindenv start` with changed monitoring configuration, the system MUST upgrade the existing deployment in place with the new values rather than requiring a tear-down and redeploy.
- **FR-017**: If the monitoring stack fails to deploy for any reason (insufficient resources, port conflicts, Helm chart errors), the system MUST log a clear warning message describing the failure and continue deploying all other enabled components without interruption.
- **FR-012**: The system MUST register any required Helm repositories for the monitoring stack during `devhelper-cli kindenv init`, consistent with how other component Helm repos are set up.
- **FR-013**: The monitoring dashboard MUST be configured with authentication disabled so that developers can access it without any login screen, consistent with how other local development dashboards (OpenSearch Dashboards, Temporal Web UI) are configured.
- **FR-014**: The monitoring stack configuration in `kindenv.yaml` MUST support specifying a Helm chart version to allow developers to pin to a known-good version.
- **FR-015**: The system MUST map the monitoring dashboard host port in the cluster's port mapping configuration when monitoring is enabled, similar to how other services (MySQL, OpenSearch Dashboards, RabbitMQ Management) are mapped.
- **FR-018**: When monitoring is set to `enabled: false` and `kindenv start` is re-run, the system MUST skip the monitoring deployment block without error. Previously deployed monitoring resources are NOT actively uninstalled; they persist until the cluster is deleted via `kindenv stop --delete`.

### Key Entities

- **Monitoring Component Configuration**: Represents the monitoring stack settings within `kindenv.yaml`, including enabled state, namespace, chart version, node ports, resource limits, and retention settings.
- **Metrics Collection Service**: The backend service responsible for discovering scrape targets, collecting time-series metrics, and storing them. Handles service discovery within the Kubernetes cluster.
- **Dashboard Service**: The visualization frontend that connects to the metrics collection service, hosts pre-built dashboards, and allows developers to create custom queries and panels.
- **Scrape Target**: Any pod or service within the cluster that exposes a metrics endpoint. Discovered automatically via Kubernetes annotations or service monitor definitions.
- **Pre-built Dashboard**: A pre-configured visualization template that ships with the monitoring stack, providing out-of-the-box views for common Kubernetes and infrastructure metrics.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can enable monitoring and have the full stack operational within 5 minutes of running `devhelper-cli kindenv start` (inclusive of the monitoring component deployment time).
- **SC-002**: Pre-built dashboards display live cluster metrics (CPU, memory, pod counts) within 2 minutes of the monitoring stack reaching a ready state.
- **SC-003**: The monitoring stack operates within 1 CPU core and 1 GB memory combined (default resource limits), ensuring it does not starve the developer's other workloads on a typical development machine.
- **SC-004**: 100% of existing components (Temporal, Redis, Dapr, MySQL, OpenSearch, RabbitMQ, KEDA, custom components) continue to function without degradation when monitoring is enabled alongside them.
- **SC-005**: Developers require zero manual configuration beyond setting `enabled: true` to get a functional monitoring stack with useful dashboards — all defaults provide a working out-of-the-box experience.
- **SC-006**: The monitoring configuration in `kindenv.yaml` follows the same patterns and conventions as existing components, requiring no new concepts for developers already familiar with devhelper-cli.
- **SC-007**: Custom application metrics exposed via standard endpoints are automatically discovered and available for querying within 3 minutes of the application pod reaching a ready state.

## Assumptions

- Developers using this feature have machines with at least 8 GB of RAM allocated to their container engine (Docker/Podman), as the monitoring stack will add resource overhead on top of existing components.
- The standard metrics endpoint format (e.g., `/metrics` path on a designated port) is the convention used for application metrics exposition. No proprietary or non-standard formats need to be supported.
- A default data retention of 24 hours is sufficient for local development use cases. Developers who need longer retention can configure it, understanding the storage implications.
- The monitoring dashboard does not need to persist its configuration (custom dashboards, user preferences) across cluster deletions. It is acceptable for custom dashboards to be lost when `kindenv stop --delete` is run.
- No alerting rules, notification integrations, log aggregation, or distributed tracing are needed for the local development monitoring stack. The scope is limited to metrics collection and visualization (see Out of Scope section).
- The monitoring stack does not need to support high availability. A single replica of each component is sufficient for local development.