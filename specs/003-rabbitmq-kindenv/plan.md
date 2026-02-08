# Implementation Plan: RabbitMQ Support for KindEnv

**Branch**: `003-rabbitmq-kindenv` | **Date**: 2026-02-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-rabbitmq-kindenv/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add RabbitMQ as a new component to kindenv that integrates with existing configuration patterns. Uses Bitnami RabbitMQ Helm chart with bitnamilegacy image repository, provides NodePort services with Kind port mapping to expose RabbitMQ AMQP (5672) and Management UI (15672) on host. Includes configurable resource limits, virtual host management, optional persistence, and health monitoring integration. Follows the same architecture pattern as MySQL component.

## Technical Context

**Language/Version**: Go 1.21+ (existing codebase)  
**Primary Dependencies**: Cobra CLI framework, Kubernetes client-go, Helm Go SDK, existing kindenv configuration system  
**Storage**: RabbitMQ 3.x via Bitnami Helm chart, Kubernetes Secrets, optional PersistentVolumes  
**Testing**: Go testing package, table-driven tests, integration tests with Kind cluster  
**Target Platform**: Local development environments (macOS, Linux, Windows) with Kind/Docker
**Project Type**: CLI extension - extends existing kindenv command structure  
**Performance Goals**: RabbitMQ startup <2 minutes, configuration changes <30 seconds, 95% startup success rate  
**Constraints**: bitnamilegacy image repository, Kind port mapping to 5672 (AMQP) and 15672 (Management UI), management plugin enabled by default  
**Scale/Scope**: Single developer environments, extends existing kindenv component architecture

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Code Quality Standards ✅
- Go code will follow official Go style guide and pass `go fmt`, `go vet`, and `golangci-lint`
- Clear, self-documenting code with meaningful variable and function names
- Comprehensive godoc comments for all exported functions, types, and packages
- Explicit error handling with wrapped errors using `fmt.Errorf`
- No code duplication - extract common functionality into reusable packages

### Test-Driven Development ✅
- Tests written before implementation (Red-Green-Refactor cycle)
- Minimum 80% code coverage for all new features
- Unit tests for all business logic using Go's testing package
- Table-driven tests for multiple input scenarios
- Integration tests for Helm chart deployment and Kubernetes interactions

### Cobra CLI Best Practices ✅
- Extends existing `devhelper-cli kindenv` command structure
- Reuses existing configuration patterns and flag definitions
- Consistent output formatting with existing `--output` flag support
- Progress indicators for RabbitMQ installation operations
- Follows existing command naming pattern: `cmd/kindenv_start.go`

### Command Design Standards ✅
- RabbitMQ installation is idempotent (safe to run multiple times)
- Integrates with existing `--verbose` mode for detailed logging
- Consistent exit codes following existing patterns
- Help text will include practical examples following existing patterns

### Error Handling & User Feedback ✅
- Error messages include context and suggested solutions
- Integrates with existing structured logging patterns using uber-go/zap
- Uses existing color-coded output libraries (fatih/color)
- Progress indicators for operations taking more than 2 seconds
- Clear success/failure indicators with next steps guidance

**GATE STATUS**: ✅ PASS - All constitutional requirements satisfied

### Post-Design Constitution Re-Check ✅

After completing Phase 1 design and contracts:

**Code Quality Standards**: ✅ Confirmed
- Interface contracts defined in `contracts/rabbitmq-api-interface.go` follow Go conventions
- Clear separation of concerns with dedicated interfaces for validation, management, and reporting
- Comprehensive error handling with custom error types

**Test-Driven Development**: ✅ Confirmed  
- Test structure defined in quickstart guide with table-driven test examples
- Integration test patterns specified for Kubernetes and Helm operations
- Unit test coverage planned for configuration validation and business logic

**Cobra CLI Best Practices**: ✅ Confirmed
- Extends existing `kindenv` command structure without breaking changes
- Reuses established flag and configuration patterns
- Maintains consistent output formatting and error handling

**Command Design Standards**: ✅ Confirmed
- RabbitMQ operations are idempotent (safe to run multiple times)
- Integrates with existing verbose and dry-run patterns
- Clear success/failure indicators with actionable next steps

**Error Handling & User Feedback**: ✅ Confirmed
- Structured error types defined for validation and operational failures
- Integration with existing color-coded output and progress indicators
- Comprehensive status reporting through existing `kindenv status` command

**FINAL GATE STATUS**: ✅ PASS - Design maintains full constitutional compliance

## Project Structure

### Documentation (this feature)

```text
specs/003-rabbitmq-kindenv/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── kindenv_start.go          # Extend with RabbitMQ installation logic
├── kindenv_status.go         # Extend with RabbitMQ health monitoring
├── kindenv_init.go           # Extend with RabbitMQ Helm repository setup
└── kindenv_start_test.go     # Add RabbitMQ-specific test cases

internal/kindenv/
├── config.go                 # Extend with RabbitMQ component configuration
├── config_test.go            # Add RabbitMQ configuration validation tests
└── rabbitmq.go               # NEW: RabbitMQ-specific installation and management logic

examples/
└── kindenv.yaml              # Update with RabbitMQ configuration examples

docs/
└── KINDENV.md                # Update with RabbitMQ usage documentation
```

**Structure Decision**: Extends existing CLI structure following established patterns. RabbitMQ functionality integrates into existing kindenv command files rather than creating separate commands, maintaining consistency with Redis, MySQL, and Temporal components. New `rabbitmq.go` file in `internal/kindenv/` package contains RabbitMQ-specific logic following the existing pattern established by `mysql.go`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | No constitutional violations | All requirements align with existing patterns |

## Phase 0: Research Outcomes

See [research.md](./research.md) for detailed findings.

### Key Research Decisions

1. **Bitnami RabbitMQ Helm Chart**
   - **Decision**: Use Bitnami RabbitMQ Helm chart (version 11.x+)
   - **Rationale**: Consistent with existing MySQL and Redis integration patterns, well-maintained, supports bitnamilegacy repository
   - **Alternatives**: RabbitMQ official operator (more complex, overkill for dev environment), custom deployment manifests (maintenance burden)

2. **Dual Port Exposure Strategy**
   - **Decision**: Expose both AMQP (5672) and Management UI (15672) via NodePort with Kind cluster port mapping
   - **Rationale**: AMQP required for messaging, Management UI essential for development/debugging
   - **Alternatives**: Ingress controller (adds complexity), port-forward only (manual, not persistent)

3. **Virtual Host Management**
   - **Decision**: Support configurable default virtual host with "/" as default
   - **Rationale**: Provides isolation for different applications while maintaining simplicity for single-app dev environments
   - **Alternatives**: Multiple virtual hosts via Helm values (complex), no virtual host support (inflexible)

4. **Management Plugin Configuration**
   - **Decision**: Enable management plugin by default in Helm chart configuration
   - **Rationale**: Essential for development workflow, provides visibility into queues, exchanges, and connections
   - **Alternatives**: Disabled by default (poor developer experience), optional plugin (configuration complexity)

5. **Secret Management Strategy**
   - **Decision**: Create dedicated Kubernetes secret for RabbitMQ credentials (username, password, erlang cookie)
   - **Rationale**: Follows existing MySQL secrets pattern, supports future clustering capabilities
   - **Alternatives**: Reuse existing secrets structure (name collision risk), inline credentials (insecure)

6. **Resource Defaults**
   - **Decision**: CPU 500m, Memory 1Gi, Storage 8Gi (same as MySQL)
   - **Rationale**: Sufficient for development workloads, consistent with existing component defaults
   - **Alternatives**: Lower resources (may impact performance), higher resources (wasteful for dev)

7. **Persistence Strategy**
   - **Decision**: Disabled by default, optional via configuration
   - **Rationale**: Faster startup for ephemeral dev environments, opt-in for stateful testing
   - **Alternatives**: Enabled by default (slower startup), no persistence option (inflexible)

## Phase 1: Design Artifacts

See the following files for detailed design:
- [data-model.md](./data-model.md) - Configuration and state entities
- [contracts/rabbitmq-api-interface.go](./contracts/rabbitmq-api-interface.go) - Go interfaces
- [contracts/rabbitmq-config-schema.yaml](./contracts/rabbitmq-config-schema.yaml) - YAML configuration schema
- [contracts/helm-values-template.yaml](./contracts/helm-values-template.yaml) - Bitnami RabbitMQ Helm values
- [quickstart.md](./quickstart.md) - Implementation guide

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     kindenv CLI Command                      │
│                  (cmd/kindenv_start.go)                      │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│              KindEnv Configuration System                    │
│                (internal/kindenv/config.go)                  │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Components.RabbitMQ Configuration                    │   │
│  │  - Enabled: bool                                     │   │
│  │  - Namespace: string                                 │   │
│  │  - ChartVersion: string                              │   │
│  │  - VirtualHost: string                               │   │
│  │  - NodePorts: {AMQP: int, Management: int}          │   │
│  │  - Resources: {CPU: string, Memory: string}         │   │
│  │  - Persistence: {Enabled: bool, Size: string}       │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│          RabbitMQ Manager Implementation                     │
│            (internal/kindenv/rabbitmq.go)                    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  RabbitMQManager Interface                          │    │
│  │  - Install(ctx, config) error                       │    │
│  │  - Uninstall(ctx, namespace) error                  │    │
│  │  - GetStatus(ctx, namespace) (*Status, error)       │    │
│  │  - ValidateConfig(config) error                     │    │
│  │  - WaitForReady(ctx, namespace, timeout) error      │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  RabbitMQConfigValidator Interface                  │    │
│  │  - ValidateVirtualHost(vhost) error                 │    │
│  │  - ValidateResources(cpu, memory) error             │    │
│  │  - ValidateNodePorts(amqp, mgmt) error              │    │
│  │  - ValidateChartVersion(version) error              │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  RabbitMQStatusReporter Interface                   │    │
│  │  - GetPodStatus(ctx, namespace, pod) (*Status, err) │    │
│  │  - GetServiceStatus(ctx, ns, svc) (*Status, err)    │    │
│  │  - TestConnection(ctx, connInfo) error              │    │
│  │  - GetHealthCheck(ctx, namespace) (*Health, err)    │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│              Helm & Kubernetes Integration                   │
│                                                               │
│  ┌───────────────────┐      ┌─────────────────────────┐     │
│  │ Helm Chart Deploy │      │ Kubernetes API Calls    │     │
│  │ - Install chart   │      │ - Create secrets        │     │
│  │ - Configure values│◄─────┤ - Create namespace      │     │
│  │ - Manage release  │      │ - Query pod status      │     │
│  └───────────────────┘      │ - Query service status  │     │
│                             └─────────────────────────┘     │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│                 Kind Cluster Resources                       │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  RabbitMQ Deployment (Namespace: rabbitmq)           │  │
│  │  - StatefulSet with persistent volumes (if enabled)  │  │
│  │  - ConfigMap with RabbitMQ configuration             │  │
│  │  - Secret with credentials (username/password/cookie)│  │
│  │  - Service (AMQP NodePort 5672)                      │  │
│  │  - Service (Management UI NodePort 15672)            │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Kind Cluster Port Mappings                           │  │
│  │  - Container Port 30672 → Host Port 5672 (AMQP)      │  │
│  │  - Container Port 31672 → Host Port 15672 (Mgmt UI)  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Component Interactions

1. **Configuration Loading**: `config.go` loads and validates RabbitMQ configuration from `kindenv.yaml`
2. **Installation Flow**: `kindenv_start.go` calls `RabbitMQManager.Install()` with validated configuration
3. **Helm Deployment**: Manager creates Kubernetes secret, then installs Bitnami RabbitMQ chart with custom values
4. **Health Monitoring**: `kindenv_status.go` calls `RabbitMQStatusReporter` to query pod/service status
5. **Port Mapping**: Kind cluster maps NodePort services to host ports via cluster configuration

### Data Flow

```
User Command (kindenv start)
    ↓
Load kindenv.yaml
    ↓
Validate RabbitMQ Config (if enabled)
    ↓
Create/Update Kubernetes Namespace
    ↓
Create Kubernetes Secret (credentials)
    ↓
Install Bitnami RabbitMQ Helm Chart
    ↓
Wait for RabbitMQ Pods Ready
    ↓
Verify Services Available (AMQP + Management)
    ↓
Report Connection Info to User
```

### Integration Points

1. **Helm Repository Management** (cmd/kindenv_init.go)
   - Add Bitnami chart repository with bitnamilegacy registry URL
   - Update repository index before installation

2. **Configuration System** (internal/kindenv/config.go)
   - Add `RabbitMQ` struct to `Components` section
   - Add `RabbitMQ` struct to `Secrets` section
   - Implement configuration defaults in `LoadConfig()`
   - Add validation in `Validate()` method

3. **Port Mapping** (internal/kindenv/config.go)
   - Add RabbitMQ port mapping generation in `generateDefaultPortMappings()`
   - Add RabbitMQ port variable substitution in `processVariableSubstitutions()`

4. **Status Reporting** (cmd/kindenv_status.go)
   - Query RabbitMQ deployment status
   - Display AMQP and Management UI connection information
   - Show pod health and readiness status

5. **Cleanup** (cmd/kindenv_stop.go)
   - Uninstall RabbitMQ Helm release
   - Clean up persistent volumes (if enabled)
   - Remove Kubernetes secrets

## Implementation Phases

### Phase 2: Implementation Tasks

Tasks will be generated by `/speckit.tasks` command and stored in [tasks.md](./tasks.md).

Expected task breakdown:
1. Configuration schema extension (2-3 tasks)
2. RabbitMQ manager implementation (4-5 tasks)
3. Helm chart integration (2-3 tasks)
4. Status reporting (2-3 tasks)
5. Testing and validation (3-4 tasks)
6. Documentation updates (1-2 tasks)

Total estimated: 14-20 tasks organized into parallel and sequential groups.

## Testing Strategy

### Unit Tests
- Configuration validation (virtual host format, port ranges, resource formats)
- Error handling and error types
- State transitions and status reporting
- Table-driven tests for multiple input scenarios

### Integration Tests
- Helm chart installation in Kind cluster
- Kubernetes secret creation and management
- Service discovery and port mapping verification
- Pod readiness and health checks
- AMQP connection testing
- Management UI accessibility testing

### Test Data
- Valid and invalid virtual host names
- Valid and invalid resource specifications
- Valid and invalid port numbers
- Multiple RabbitMQ configuration scenarios

### Test Coverage Goals
- Minimum 80% code coverage for `rabbitmq.go`
- 100% coverage for validation functions
- Integration tests for all public interfaces

## Performance Considerations

- RabbitMQ startup time: Target <2 minutes for pod ready state
- Configuration validation: <100ms for typical configurations
- Status queries: <500ms for health check operations
- Memory footprint: Within 1Gi default limit during normal operation
- Concurrent operations: Safe for parallel kindenv component installations

## Security Considerations

- Credentials stored in Kubernetes secrets (not in config files)
- Erlang cookie generated securely for clustering support
- NodePort services exposed only within Kind cluster network
- No privileged containers required
- Resource limits enforced to prevent resource exhaustion
- Management UI accessible only via localhost:15672 (Kind port mapping)

## Migration Path

No migration required - this is a new component addition. Existing kindenv installations will continue to work without RabbitMQ enabled.

Users can opt-in by:
1. Adding `components.rabbitmq.enabled: true` to `kindenv.yaml`
2. Running `devhelper-cli kindenv start`
3. RabbitMQ will be installed and configured automatically

## Rollback Strategy

If RabbitMQ installation fails or causes issues:
1. Set `components.rabbitmq.enabled: false` in `kindenv.yaml`
2. Run `devhelper-cli kindenv stop` (uninstalls RabbitMQ Helm release)
3. Run `devhelper-cli kindenv start` (restarts without RabbitMQ)

Complete rollback: No database migrations or persistent state changes required.

## Future Enhancements

Potential future improvements (out of scope for this feature):
- RabbitMQ clustering support (multi-pod setup)
- Federation/shovel plugin configuration
- Custom plugins installation
- Advanced exchange and queue pre-configuration
- Automated backup/restore for persistent deployments
- Integration with external monitoring systems (Prometheus metrics)

## Dependencies

External dependencies required:
- Bitnami Helm repository configured with bitnamilegacy registry
- Kind cluster with port mapping support
- Kubernetes 1.21+ (existing kindenv requirement)
- Helm 3+ (existing kindenv requirement)

Go module dependencies (already present):
- `k8s.io/client-go` - Kubernetes API client
- `helm.sh/helm/v3` - Helm SDK
- `github.com/spf13/cobra` - CLI framework
- `go.uber.org/zap` - Structured logging

## References

- [Bitnami RabbitMQ Helm Chart](https://github.com/bitnami/charts/tree/main/bitnami/rabbitmq)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Kubernetes NodePort Services](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport)
- [Kind Port Mappings](https://kind.sigs.k8s.io/docs/user/configuration/#extra-port-mappings)
- [MySQL Implementation Reference](../001-mysql8-kindenv/plan.md) - Architecture pattern
