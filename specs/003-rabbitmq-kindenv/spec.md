# Feature Specification: RabbitMQ Support for KindEnv

**Feature Branch**: `003-rabbitmq-kindenv`  
**Created**: 2026-02-05  
**Status**: Draft  
**Input**: User description: "let's add rabbitmq to our kindenv cmd. same as we added mysql"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic RabbitMQ Installation (Priority: P1)

A developer wants to add RabbitMQ to their Kind development environment to test applications that require a message broker. They should be able to enable RabbitMQ through configuration and have it automatically installed and configured when starting their kindenv.

**Why this priority**: This is the core functionality that enables developers to use RabbitMQ in their local development environment, which is essential for testing message-driven applications, event streaming, and asynchronous communication patterns.

**Independent Test**: Can be fully tested by enabling RabbitMQ in kindenv configuration, running `devhelper-cli kindenv start`, and verifying that RabbitMQ is running and accessible within the Kind cluster.

**Acceptance Scenarios**:

1. **Given** a kindenv configuration with RabbitMQ enabled, **When** running `devhelper-cli kindenv start`, **Then** RabbitMQ should be installed and running in the Kind cluster
2. **Given** RabbitMQ is enabled in configuration, **When** the kindenv starts successfully, **Then** RabbitMQ should be accessible on the configured AMQP port (5672) with proper credentials
3. **Given** RabbitMQ is enabled in configuration, **When** the kindenv starts successfully, **Then** RabbitMQ Management UI should be accessible on the configured HTTP port (15672)
4. **Given** RabbitMQ is disabled in configuration, **When** running kindenv start, **Then** RabbitMQ should not be installed or consume cluster resources

---

### User Story 2 - RabbitMQ Configuration Management (Priority: P2)

A developer wants to customize RabbitMQ settings such as username, password, virtual host, ports, and resource limits to match their application requirements and development environment constraints.

**Why this priority**: Developers need flexibility to configure RabbitMQ according to their specific application needs and local environment constraints, including custom virtual hosts and authentication.

**Independent Test**: Can be tested by modifying RabbitMQ configuration parameters in kindenv.yaml, starting the environment, and verifying that RabbitMQ runs with the specified settings.

**Acceptance Scenarios**:

1. **Given** custom RabbitMQ credentials in configuration, **When** kindenv starts, **Then** RabbitMQ should be accessible using the specified username and password
2. **Given** a custom virtual host in configuration, **When** kindenv starts, **Then** the specified virtual host should be created and available for connections
3. **Given** custom resource limits in configuration, **When** RabbitMQ is deployed, **Then** it should respect the specified CPU and memory constraints
4. **Given** custom port mappings in configuration, **When** kindenv starts, **Then** RabbitMQ should be accessible on both AMQP and Management UI ports

---

### User Story 3 - RabbitMQ Data Persistence (Priority: P3)

A developer wants their RabbitMQ data (queues, exchanges, messages) to persist across kindenv restarts so they don't lose their development messages and queue configurations when stopping and starting their environment.

**Why this priority**: Data persistence improves developer productivity by maintaining message queues and configurations across development sessions, but it's not critical for initial RabbitMQ functionality.

**Independent Test**: Can be tested by creating queues and messages in RabbitMQ, stopping kindenv, restarting it, and verifying that the queues and messages are still present.

**Acceptance Scenarios**:

1. **Given** RabbitMQ with existing queues and messages, **When** kindenv is stopped and restarted, **Then** all queues, exchanges, bindings, and durable messages should be preserved
2. **Given** persistence is disabled in configuration, **When** kindenv restarts, **Then** RabbitMQ should start with a clean state

---

### User Story 4 - RabbitMQ Health Monitoring (Priority: P3)

A developer wants to monitor RabbitMQ health and status through kindenv commands to troubleshoot issues and verify that RabbitMQ is running correctly.

**Why this priority**: Health monitoring helps with troubleshooting but is not essential for basic RabbitMQ functionality.

**Independent Test**: Can be tested by running `devhelper-cli kindenv status` and verifying that RabbitMQ status information is displayed accurately.

**Acceptance Scenarios**:

1. **Given** RabbitMQ is running in kindenv, **When** running `devhelper-cli kindenv status`, **Then** RabbitMQ status should be displayed with connection information for both AMQP and Management UI
2. **Given** RabbitMQ is not running, **When** checking kindenv status, **Then** RabbitMQ should be shown as unavailable with diagnostic information

---

### Edge Cases

- What happens when RabbitMQ fails to start due to resource constraints?
- How does the system handle RabbitMQ image pull failures?
- What occurs when RabbitMQ configuration contains invalid parameters?
- How does the system behave when RabbitMQ ports conflict with other services?
- What happens when RabbitMQ persistent volume claims cannot be created?
- How does the system handle RabbitMQ cluster plugin conflicts or incompatible configurations?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support enabling/disabling RabbitMQ through kindenv configuration
- **FR-002**: System MUST install RabbitMQ using Bitnami RabbitMQ Helm chart with bitnamilegacy image repository when ECR is enabled
- **FR-003**: System MUST create RabbitMQ with configurable username, password, and virtual host
- **FR-004**: System MUST expose RabbitMQ AMQP protocol on NodePort service with Kind cluster port mapping to default AMQP port 5672 on host
- **FR-005**: System MUST expose RabbitMQ Management UI on NodePort service with Kind cluster port mapping to default management port 15672 on host
- **FR-006**: System MUST support RabbitMQ resource limits with defaults: CPU 500m, Memory 1Gi, Storage 8Gi
- **FR-007**: System MUST create Kubernetes secrets for RabbitMQ credentials automatically
- **FR-008**: System MUST validate RabbitMQ configuration parameters before installation
- **FR-009**: System MUST provide RabbitMQ connection information (AMQP and Management UI URLs) in kindenv status output
- **FR-010**: System MUST support optional data persistence through persistent volumes (disabled by default, configurable to enable)
- **FR-011**: System MUST handle RabbitMQ startup dependencies and health checks
- **FR-012**: System MUST integrate with existing ECR image registry configuration when enabled
- **FR-013**: System MUST clean up RabbitMQ resources when kindenv is stopped
- **FR-014**: System MUST configure RabbitMQ with default virtual host "/" when not specified
- **FR-015**: System MUST enable RabbitMQ management plugin by default for UI access

### Key Entities

- **RabbitMQ Configuration**: Username, password, virtual host, AMQP port, management port, resource limits, persistence settings
- **RabbitMQ Deployment**: Kubernetes deployment, service, persistent volume claims, secrets
- **RabbitMQ Service**: NodePort services for AMQP (5672) and Management UI (15672) access, internal cluster service for pod communication
- **RabbitMQ Credentials**: Username, password, erlang cookie stored as Kubernetes secrets
- **RabbitMQ Virtual Host**: Namespace for isolating resources within RabbitMQ instance

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can enable RabbitMQ in kindenv configuration and have it running within 2 minutes of starting kindenv
- **SC-002**: RabbitMQ AMQP protocol is accessible from the host machine on the configured port (5672) with 100% success rate
- **SC-003**: RabbitMQ Management UI is accessible from the host machine on the configured port (15672) with 100% success rate
- **SC-004**: RabbitMQ configuration changes take effect within 30 seconds of kindenv restart
- **SC-005**: RabbitMQ startup success rate is 95% or higher in clean Kind environments
- **SC-006**: RabbitMQ resource usage stays within configured limits during normal operation
- **SC-007**: Data persistence works correctly in 100% of stop/restart cycles when enabled
- **SC-008**: Developers can connect to RabbitMQ and publish/consume messages within 30 seconds of startup

## Assumptions

- Developers have Docker and Kind already configured and working
- Bitnami RabbitMQ Helm chart and bitnamilegacy images are available in the configured container registry
- Sufficient cluster resources are available for RabbitMQ deployment (minimum 500m CPU, 1Gi memory)
- Developers understand basic RabbitMQ connection concepts and AMQP protocol
- Kubernetes persistent volume support is available when persistence is enabled
- Kind cluster port mapping is configured to expose RabbitMQ on standard ports 5672 (AMQP) and 15672 (Management UI)
- RabbitMQ will be deployed as a standalone instance (not clustered) for development purposes
- Default configuration includes management plugin enabled for UI access
