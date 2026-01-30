# Implementation Plan: Custom Components for KindEnv

**Branch**: `002-custom-components` | **Date**: 2026-01-30 | **Spec**: [spec.md](spec.md)

**Note**: This plan defines the technical approach for enabling developers to deploy custom services with configurable images, commands, environment variables, port mappings, resource limits, and configuration file mounting in kindenv clusters.

## Summary

This feature extends the kindenv CLI to support user-defined custom components alongside the existing built-in components (MySQL, OpenSearch, Temporal, etc.). Developers will be able to configure custom services in kindenv.yaml with full Kubernetes deployment capabilities including:

1. **Container specification**: Image, command, args override
2. **Configuration**: Environment variables (direct values and secret references), custom config file mounting
3. **Networking**: Port mappings from container to host with NodePort services
4. **Resources**: CPU and memory requests and limits
5. **Metadata**: Labels, annotations, replicas
6. **Lifecycle**: Deployment after infrastructure, status reporting, clean removal

The implementation follows existing kindenv patterns for component management, configuration parsing, and Kubernetes deployment orchestration. Configuration files are specified inline in kindenv.yaml and automatically converted to Kubernetes ConfigMaps that are mounted as read-only volumes.

## Technical Context

**Language/Version**: Go 1.23+ (current toolchain: 1.24.1)  

**Primary Dependencies**: 
- Cobra v1.9.1 (CLI framework)
- Viper v1.20.1 (configuration management)
- gopkg.in/yaml.v3 (YAML parsing)
- fatih/color v1.18.0 (terminal output formatting)
- testify v1.10.0 (testing framework)

**Storage**: YAML configuration files (kindenv.yaml), Kubernetes cluster state (ConfigMaps for config files)

**Testing**: Go testing package with testify assertions, table-driven tests, integration tests with kubectl/kind

**Target Platform**: macOS/Linux developer workstations with Docker/Podman + Kind + kubectl

**Project Type**: Single project (CLI application with Cobra commands)

**Performance Goals**: 
- Configuration validation: <100ms for typical config files
- Deployment orchestration: <30 seconds for custom components (parallel deployment)
- ConfigMap creation: <5 seconds per component
- Status checking: <2 seconds for component status queries

**Constraints**: 
- Must not break existing kindenv.yaml configurations (backward compatibility)
- Must follow existing component deployment patterns (consistency)
- Configuration changes require cluster restart (existing behavior)
- Maximum 32 total port mappings per cluster (Kind limitation)
- ConfigMap size limit: 1MB per component (Kubernetes limitation)

**Scale/Scope**: 
- Support 10+ concurrent custom components
- Support 10+ config files per component
- Configuration files up to 1000 lines total
- Single kindenv.yaml per project

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Code Quality Standards ✅
- [ ] All Go code will follow official style guide (go fmt, go vet, golangci-lint)
- [ ] Clear, self-documenting code with meaningful names
- [ ] Comprehensive godoc comments for all exported types/functions
- [ ] Explicit error handling with wrapped errors using fmt.Errorf
- [ ] No code duplication - extract common functionality to internal/kindenv

### Test-Driven Development ✅
- [ ] Tests written before implementation (Red-Green-Refactor)
- [ ] Minimum 80% code coverage for new features
- [ ] Unit tests for all configuration parsing and validation logic
- [ ] Table-driven tests for multiple component configuration scenarios
- [ ] Integration tests for Kubernetes deployment orchestration (ConfigMap creation, volume mounting)
- [ ] Benchmark tests for configuration validation performance

### Cobra CLI Best Practices ✅
- [ ] Consistent command structure (extends existing `kindenv start/stop/status`)
- [ ] Clear error messages for configuration validation failures
- [ ] No new commands required (extends existing commands)
- [ ] Consistent output formatting with existing kindenv commands
- [ ] Progress indicators for custom component deployment

### Command Design Standards ✅
- [ ] Idempotent operations (safe to run `kindenv start` multiple times)
- [ ] Verbose mode support (already exists in kindenv commands)
- [ ] Consistent exit codes (0=success, 1=user error, 2=system error)
- [ ] Help text includes custom component examples in kindenv.yaml

### Error Handling & User Feedback ✅
- [ ] Error messages include context and suggested solutions
- [ ] Structured logging with appropriate levels
- [ ] Color-coded output for readability (using fatih/color)
- [ ] Progress indicators for deployment operations >2 seconds
- [ ] Clear success/failure indicators with next steps

### Compliance Summary
**Status**: ✅ **PASSES** - All constitutional requirements satisfied

**Rationale**:
- Extends existing kindenv command structure (no new commands required)
- Follows established patterns for component configuration and deployment
- TDD approach with comprehensive test coverage planned
- Consistent with existing code quality and CLI standards
- No constitutional violations requiring justification

## Project Structure

### Documentation (this feature)

```text
specs/002-custom-components/
├── plan.md                          # This file (implementation plan)
├── research.md                      # Phase 0: Research on K8s patterns, ConfigMaps
├── data-model.md                    # Phase 1: CustomComponent, ConfigFile data structures
├── quickstart.md                    # Phase 1: User quick start guide
├── contracts/                       # Phase 1: YAML schema and Go interfaces
│   ├── custom-component-schema.yaml
│   ├── custom-component-api-interface.go
│   └── deployment-template.yaml
├── checklists/
│   └── requirements.md              # Spec quality validation
├── PLAN_UPDATE_CONFIG_FILES.md      # Config file mounting update summary
└── tasks.md                         # Phase 2: Implementation tasks (from /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
├── kindenv_start.go         # MODIFY: Add custom component deployment logic
├── kindenv_start_test.go    # MODIFY: Add custom component deployment tests
├── kindenv_stop.go          # MODIFY: Add custom component cleanup logic
├── kindenv_stop_test.go     # MODIFY: Add custom component cleanup tests
├── kindenv_status.go        # MODIFY: Add custom component status reporting
└── kindenv_status_test.go   # MODIFY: Add custom component status tests

internal/kindenv/
├── config.go                # MODIFY: Add CustomComponent, ConfigFile structs and parsing
├── config_test.go           # MODIFY: Add CustomComponent configuration tests
├── customcomponent.go       # NEW: Custom component deployment logic
├── customcomponent_test.go  # NEW: Custom component unit tests
├── configmap.go             # NEW: ConfigMap creation and management
├── configmap_test.go        # NEW: ConfigMap unit tests
├── volume.go                # NEW: Volume and VolumeMount generation
├── volume_test.go           # NEW: Volume generation tests
└── validation.go            # NEW: Custom component validation logic
    └── validation_test.go   # NEW: Validation tests

examples/
└── custom-component-app/    # NEW: Example Spring Boot application
    ├── Dockerfile
    ├── config/
    │   ├── application.yaml
    │   └── logback.xml
    ├── main.go or app.jar
    └── README.md

docs/ (or root-level documentation)
└── CUSTOM_COMPONENTS.md     # NEW: User guide for custom components

kindenv.yaml                 # MODIFY: Add example custom component configuration
```

**Structure Decision**: Single project structure - The devhelper-cli is a standalone CLI application using Cobra. All code lives in `cmd/` for command handlers and `internal/` for shared business logic. This feature extends the existing `kindenv` subcommand family with custom component support, following the established pattern of component management used for MySQL, OpenSearch, Temporal, etc.

## Complexity Tracking

> **No constitutional violations** - This section is not applicable.

All implementation follows established patterns and constitutional requirements. No complexity justifications needed.

---

## Phase 0: Research ✅ COMPLETE

**Status**: All research tasks completed successfully.

**Deliverables**:
- ✅ `research.md`: Comprehensive technical decisions and rationale
- ✅ All unknowns from Technical Context resolved
- ✅ Technology choices documented and justified
- ✅ Best practices identified for Kubernetes deployments, secret management, configuration validation, and ConfigMap mounting

**Key Decisions**:
1. **Deployment Pattern**: Kubernetes Deployments for component orchestration
2. **Service Exposure**: NodePort Services for external access
3. **Configuration Schema**: Extended KindEnvConfig struct for YAML parsing
4. **Environment Variables**: Support both direct values and SecretKeyRef
5. **Validation**: Multi-phase (config → pre-deploy → runtime)
6. **Resource Limits**: Default limits with override support
7. **Secret Management**: Kubernetes SecretKeyRef (native)
8. **Deployment Order**: After infrastructure components (MySQL, OpenSearch)
9. **Status Reporting**: Extended kindenv status command
10. **Registry Support**: Leverage existing ECR/Harbor integration
11. **Config File Mounting**: ConfigMaps with volume mounts (inline contents)

---

## Phase 1: Design & Contracts ✅ COMPLETE

**Status**: All design artifacts generated successfully.

**Deliverables**:
- ✅ `data-model.md`: Complete Go struct definitions with validation rules
- ✅ `contracts/custom-component-schema.yaml`: YAML configuration schema
- ✅ `contracts/custom-component-api-interface.go`: Go interface definitions
- ✅ `contracts/deployment-template.yaml`: Kubernetes manifest templates
- ✅ `quickstart.md`: User-facing quick start guide
- ✅ Agent context updated with new patterns and dependencies

**Data Model Summary**:

1. **CustomComponent**: Main entity with 14 configuration fields
   - Identity: name, enabled, namespace
   - Container: image, command, args
   - Configuration: env, configFiles
   - Networking: ports
   - Resources: resources, replicas
   - Metadata: labels, annotations

2. **EnvVar**: Environment variable with direct value or secret reference support
   - Supports: value (string) or valueFrom.secretKeyRef

3. **ConfigFile**: Custom configuration file for mounting
   - Attributes: name, path, contents
   - Generates: Kubernetes ConfigMap
   - Mounted as: Read-only volume with subPath

4. **PortMapping**: Port exposure with NodePort/HostPort configuration
   - Fields: containerPort, hostPort, protocol, nodePort

5. **ResourceRequirements**: CPU and memory limits with validation
   - Requests and Limits for CPU and Memory

6. **Probe**: Health check support (future enhancement)

**Interface Design**:
- `CustomComponentManager`: Core deployment orchestration
- `KubernetesClient`: Abstraction for kubectl operations (testable)
- `ConfigMapManager`: ConfigMap creation and management
- `Validator`: Multi-phase validation logic
- `PortManager`: Port allocation and conflict detection
- `StatusReporter`: Formatted status output
- `TemplateGenerator`: Kubernetes manifest generation

---

## Phase 1 Detailed: Configuration Schema

### YAML Structure

```yaml
customComponents:
  - name: my-spring-app              # Required: DNS-compatible name
    enabled: true                    # Optional: Enable/disable (default: true)
    namespace: default               # Optional: K8s namespace (default: "default")
    image: registry/repo:tag         # Required: Container image
    command: ["java"]                # Optional: Override ENTRYPOINT
    args: ["-jar", "app.jar"]        # Optional: Container args
    replicas: 1                      # Optional: Pod count (default: 1)
    
    # Environment variables
    env:
      - name: APP_ENV               # Direct value
        value: "development"
      - name: DB_PASSWORD           # Secret reference
        valueFrom:
          secretKeyRef:
            name: mysql-secret
            key: password
    
    # Configuration files (NEW)
    configFiles:
      - name: application.yaml
        path: /config/application.yaml
        contents: |
          server:
            port: 8080
          database:
            host: mysql
    
    # Port mappings
    ports:
      - containerPort: 8080
        hostPort: 8080
        protocol: TCP
        nodePort: 30800
    
    # Resource limits
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
    
    # Metadata
    labels:
      tier: backend
    annotations:
      description: "My app"
```

### Generated Kubernetes Resources

For each custom component, the system generates:

1. **ConfigMap** (if configFiles specified):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-spring-app-config
  namespace: default
data:
  application.yaml: |
    server:
      port: 8080
```

2. **Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-spring-app
  namespace: default
spec:
  replicas: 1
  template:
    spec:
      volumes:
        - name: my-spring-app-config-volume
          configMap:
            name: my-spring-app-config
            defaultMode: 0644
      containers:
        - name: my-spring-app
          image: registry/repo:tag
          volumeMounts:
            - name: my-spring-app-config-volume
              mountPath: /config/application.yaml
              subPath: application.yaml
              readOnly: true
```

3. **Service** (if ports specified):
```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-spring-app-service
  namespace: default
spec:
  type: NodePort
  ports:
    - port: 8080
      targetPort: 8080
      nodePort: 30800
```

---

## Constitution Re-Check (Post-Design)

**Status**: ✅ **STILL COMPLIANT** - No violations introduced during design phase.

### Code Quality Standards ✅
- Data structures follow Go conventions and best practices
- Clear interfaces for testing and extensibility
- Comprehensive validation with detailed error messages
- No code duplication in design (DRY principle)

### Test-Driven Development ✅
- Unit test examples provided in data-model.md
- Table-driven test structure defined
- Integration test strategy documented (ConfigMap creation, volume mounting)
- Benchmark test requirements identified
- Target: 80%+ code coverage

### Cobra CLI Best Practices ✅
- No new commands (extends existing kindenv start/stop/status)
- Configuration-driven approach (YAML-based)
- Clear error messages designed in validation logic
- Consistent with existing kindenv patterns
- Progress indicators planned for deployment

### Command Design Standards ✅
- Idempotent operations (safe to run multiple times)
- Validation before execution (fail fast)
- Backward compatible (existing configs work unchanged)
- Clear success/failure reporting

### Error Handling & User Feedback ✅
- Structured error types defined (ValidationError, DeploymentError, ConflictError)
- Context-rich error messages with suggestions
- Color-coded output using existing fatih/color library
- Progressive status reporting during deployment

**Final Verdict**: Implementation plan fully complies with DevHelper CLI Constitution. Ready for Phase 2 (Task Breakdown).

---

## Next Steps

**Phase 1 Complete** ✅

**Ready for Phase 2**: Task Breakdown

Run `/speckit.tasks` to generate the implementation task list based on this plan.

**Recommended Execution Order**:
1. Create `tasks.md` with `/speckit.tasks` command
2. Implement core data structures (CustomComponent, ConfigFile, EnvVar, PortMapping, ResourceRequirements)
3. Add configuration parsing and validation
4. Implement ConfigMap creation and management
5. Implement volume/volumeMount generation
6. Implement deployment orchestration
7. Extend kindenv start command (deploy custom components)
8. Extend kindenv stop command (cleanup custom components + ConfigMaps)
9. Extend kindenv status command (display custom component status)
10. Write comprehensive tests (unit → integration)
11. Create example application
12. Update documentation

**Estimated Scope**:
- **New files**: ~10-12 Go files (~2500-3000 LOC)
  - customcomponent.go, configmap.go, volume.go, validation.go
  - Corresponding test files
- **Modified files**: ~6-8 existing files (~500-800 LOC changes)
  - config.go, kindenv_start.go, kindenv_stop.go, kindenv_status.go
- **Test files**: ~12-15 test files (~2000-2500 LOC)
  - Unit tests, integration tests, table-driven tests
- **Documentation**: ~5-6 files
  - README updates, CUSTOM_COMPONENTS.md, example app README
- **Total implementation effort**: **Large** feature

**Feature Breakdown by Priority**:
- P1: Basic deployment (image, env vars) - ~40% of effort
- P2: Secret references - ~10% of effort
- P3: Command/args override - ~5% of effort
- P4: Port exposure and mapping - ~15% of effort
- P5: Resource limits - ~10% of effort
- P6: Config file mounting - ~20% of effort

---

**Plan Status**: ✅ **COMPLETE**  
**Branch**: `002-custom-components`  
**Plan Path**: `/Users/ilya/repo/devhelper-cli/specs/002-custom-components/plan.md`

**Generated Artifacts**:
1. ✅ plan.md (this file)
2. ✅ research.md
3. ✅ data-model.md
4. ✅ contracts/custom-component-schema.yaml
5. ✅ contracts/custom-component-api-interface.go
6. ✅ contracts/deployment-template.yaml
7. ✅ quickstart.md
8. ✅ PLAN_UPDATE_CONFIG_FILES.md
9. ✅ Updated agent context

**Ready for**: `/speckit.tasks` to break down implementation into actionable tasks covering all custom component features (deployment, environment variables, secrets, ports, resources, and configuration file mounting).
