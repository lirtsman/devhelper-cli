# Research: Custom Components for KindEnv

**Date**: 2026-01-30  
**Feature**: Custom Components for KindEnv  
**Phase**: 0 - Technical Research and Decision Making

## Overview

This document consolidates research findings for implementing custom component support in kindenv, including Kubernetes deployment patterns, configuration management approaches, and integration with existing component infrastructure.

## Research Areas

### 1. Kubernetes Deployment Patterns for Custom Workloads

#### Decision: Use Kubernetes Deployment Resources

**Rationale**:
- Deployments provide declarative updates for Pods and ReplicaSets
- Built-in support for rolling updates and rollbacks
- Automatic pod restart and rescheduling on node failure
- Consistent with existing kindenv components (MySQL, OpenSearch, etc.)
- Native support for replicas, resource limits, and health checks

**Alternatives Considered**:

| Alternative | Pros | Cons | Why Rejected |
|------------|------|------|--------------|
| StatefulSets | Stable network identity, ordered deployment | Overkill for stateless dev services | Custom components are typically stateless in dev environments |
| DaemonSets | Runs on every node | Not applicable for app deployment | Need selective deployment, not node-level |
| Bare Pods | Simplest approach | No auto-restart, no rolling updates | Production-like patterns preferred |
| Jobs/CronJobs | Good for batch work | Not for long-running services | Custom components are typically services, not jobs |

**Implementation Approach**:
- Generate Deployment YAML programmatically in Go
- Use kubectl apply for declarative deployment
- Support standard Deployment features: replicas, selectors, pod templates
- Follow existing pattern in `cmd/kindenv_start.go` for component deployment

#### Decision: Use NodePort Services for External Access

**Rationale**:
- Kind cluster is local development environment
- NodePort allows host-to-cluster access without complex networking
- Consistent with existing components (MySQL port 30306, OpenSearch 30920)
- Simple port mapping configuration for developers
- NodePort range 30000-32767 sufficient for dev use cases

**Alternatives Considered**:

| Alternative | Pros | Cons | Why Rejected |
|------------|------|------|--------------|
| LoadBalancer | Production-like | Requires external LB (complex in local dev) | Overkill for local development |
| Ingress | HTTP routing, path-based routing | Requires ingress controller setup | NodePort simpler for direct service access |
| Port Forwarding | No config needed | Manual kubectl command required | Not persistent, requires manual intervention |
| ClusterIP only | Simplest | No external access | Defeats purpose of exposing services to host |

**Implementation Approach**:
- Create Service resources with type: NodePort
- Map container ports to NodePorts in 30000-32767 range
- Validate port conflicts with existing components
- Update Kind cluster port mappings automatically

### 2. Environment Variable Management and Secret References

#### Decision: Support Both Direct Values and SecretKeyRef

**Rationale**:
- Direct values: Simple, clear for non-sensitive configuration
- SecretKeyRef: Secure credential management, mirrors production patterns
- Kubernetes-native approach (no custom secret management)
- Allows gradual security improvement (start with direct, move to secrets)
- Compatible with existing MySQL/OpenSearch secret patterns

**Implementation Details**:

```yaml
env:
  # Direct value (simple configuration)
  - name: APP_ENV
    value: "development"
  
  # Secret reference (credentials)
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: mysql-secret
        key: password
```

**Secret Management Strategy**:
- Reuse existing secrets (mysql-secret, opensearch-secret)
- Support user-defined secrets (created manually or via additional config)
- Validate secret existence before deployment
- Clear error messages for missing secrets
- Document secret creation in quickstart guide

#### Decision: Support ConfigMap References (Future Enhancement)

**Current Scope**: Not included in initial implementation (P6 priority)

**Rationale for Deferral**:
- Direct environment values sufficient for MVP
- Secrets cover credential use cases
- ConfigMaps add complexity without immediate value
- Can be added in future iteration without breaking changes

### 3. Configuration Schema Design

#### Decision: Extend Existing KindEnvConfig Struct

**Rationale**:
- Consistent with existing component configuration pattern
- Leverage existing YAML parsing infrastructure (gopkg.in/yaml.v3)
- Maintains single configuration file (kindenv.yaml)
- Familiar structure for users already using kindenv
- Supports variable substitution like existing components

**Configuration Structure**:

```yaml
customComponents:
  - name: my-spring-app
    enabled: true
    namespace: default
    image: registry.example.com/my-app:latest
    replicas: 1
    command: ["/bin/sh"]  # Optional: override container command
    args: ["-c", "java -jar app.jar"]  # Optional: container args
    env:
      - name: SPRING_PROFILES_ACTIVE
        value: "local"
      - name: DB_PASSWORD
        valueFrom:
          secretKeyRef:
            name: mysql-secret
            key: password
    ports:
      - containerPort: 8080
        hostPort: 8080
        protocol: TCP
    resources:
      requests:
        cpu: "200m"
        memory: "512Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
    labels:
      app: my-spring-app
      tier: backend
    annotations:
      description: "Custom Spring Boot application"
```

**Go Struct Design**:

```go
type CustomComponent struct {
    Name        string                 `yaml:"name"`
    Enabled     bool                   `yaml:"enabled"`
    Namespace   string                 `yaml:"namespace"`
    Image       string                 `yaml:"image"`
    Replicas    int                    `yaml:"replicas"`
    Command     []string               `yaml:"command,omitempty"`
    Args        []string               `yaml:"args,omitempty"`
    Env         []EnvVar               `yaml:"env,omitempty"`
    Ports       []PortMapping          `yaml:"ports,omitempty"`
    Resources   ResourceRequirements   `yaml:"resources,omitempty"`
    Labels      map[string]string      `yaml:"labels,omitempty"`
    Annotations map[string]string      `yaml:"annotations,omitempty"`
}

type EnvVar struct {
    Name      string         `yaml:"name"`
    Value     string         `yaml:"value,omitempty"`
    ValueFrom *EnvVarSource  `yaml:"valueFrom,omitempty"`
}

type EnvVarSource struct {
    SecretKeyRef *SecretKeySelector `yaml:"secretKeyRef,omitempty"`
}

type SecretKeySelector struct {
    Name string `yaml:"name"`
    Key  string `yaml:"key"`
}

type PortMapping struct {
    ContainerPort int    `yaml:"containerPort"`
    HostPort      int    `yaml:"hostPort,omitempty"`
    Protocol      string `yaml:"protocol"`
    NodePort      int    `yaml:"nodePort,omitempty"`
}

type ResourceRequirements struct {
    Requests ResourceList `yaml:"requests,omitempty"`
    Limits   ResourceList `yaml:"limits,omitempty"`
}

type ResourceList struct {
    CPU    string `yaml:"cpu,omitempty"`
    Memory string `yaml:"memory,omitempty"`
}
```

### 4. Validation Strategy

#### Decision: Multi-Phase Validation (Configuration → Pre-Deployment → Runtime)

**Phase 1: Configuration Parsing Validation**
- YAML syntax validation (handled by yaml.v3)
- Required field validation (name, image)
- Type validation (replicas as int, ports as int, etc.)
- Enum validation (protocol: TCP|UDP)

**Phase 2: Pre-Deployment Validation**
- Image format validation (registry/repo:tag)
- Port conflict detection (with existing components)
- Namespace existence check
- Secret reference validation (secret exists in cluster)
- Resource format validation (CPU: "500m", Memory: "1Gi")
- Replica count > 0

**Phase 3: Runtime Validation**
- Kubernetes API validation (deployment creation)
- Image pull validation (registry access, image exists)
- Pod startup validation (container starts successfully)
- Readiness probe validation (if configured)

**Error Reporting Strategy**:
- Configuration errors: Fail fast before cluster operations
- Pre-deployment errors: Detailed error with suggested fixes
- Runtime errors: Progressive reporting with kubectl-style output
- All errors include context (component name, field, value)

**Example Error Messages**:

```
❌ Configuration Error in custom component 'my-app':
   - Field 'image' is required but not specified
   - Port 3306 conflicts with MySQL component (hostPort must be unique)
   - Secret 'app-secret' referenced in env var 'API_KEY' does not exist in namespace 'default'
   
   Suggestions:
   - Add 'image' field with format 'registry/repository:tag'
   - Change hostPort to an available port (e.g., 8080)
   - Create secret: kubectl create secret generic app-secret --from-literal=API_KEY=your-key
```

### 5. Deployment Orchestration and Lifecycle

#### Decision: Deploy Custom Components After Infrastructure Components

**Deployment Order**:
1. Infrastructure namespace creation
2. Secret creation (MySQL, OpenSearch, custom)
3. Infrastructure components (MySQL, OpenSearch, etc.)
4. Wait for infrastructure readiness
5. **Custom components deployment** ← New phase
6. Port mapping updates (if needed)

**Rationale**:
- Custom apps often depend on databases/search engines
- Ensures dependencies are ready before app starts
- Mirrors production deployment patterns
- Reduces startup failures and retry loops

**Parallel Deployment**:
- Deploy multiple custom components in parallel (independent components)
- Use goroutines for concurrent kubectl apply operations
- Aggregate results and report status
- Continue on individual failures (report at end)

**Cleanup Strategy** (kindenv stop):
- Delete custom component deployments
- Delete custom component services
- Preserve namespaces (consistent with existing behavior)
- Preserve secrets (reusable across restarts)
- Clean removal without orphaned resources

### 6. Status Reporting Integration

#### Decision: Extend kindenv status Command

**Display Format**:

```
Custom Components:
  ✅ my-spring-app (default namespace)
     Status: Running (1/1 pods ready)
     Image: registry.example.com/my-app:v1.2.3
     Ports: 8080:8080 (TCP)
     
  ⏳ background-worker (default namespace)
     Status: Pending (0/1 pods ready)
     Image: registry.example.com/worker:latest
     Reason: ImagePullBackOff
     
  ❌ api-service (api namespace)
     Status: Failed
     Reason: CrashLoopBackOff
     Last Error: Error: database connection refused
```

**Status Checks**:
- Query Deployment status via kubectl
- Check pod readiness status
- Report image pull status
- Show resource usage (if requested)
- Display recent errors/events

### 7. Image Registry and Pull Secret Management

#### Decision: Leverage Existing ECR/Harbor Integration

**Existing Infrastructure**:
- ECR credentials already managed for AWS registry
- Harbor credentials for third-party images
- Pull secrets automatically created in namespaces

**Custom Component Integration**:
- Custom components use same pull secret mechanism
- Support public registries (Docker Hub) without credentials
- Support private registries via existing ECR/Harbor config
- Support custom registries (requires manual pull secret)

**Image Pull Strategy**:
- Default: Always pull (IfNotPresent for faster iteration)
- Configurable via image pull policy (future enhancement)
- Clear errors for authentication failures
- Document registry configuration in quickstart

### 8. Resource Limits and Quality of Service

#### Decision: Default Resource Limits with Override Support

**Default Resources** (if not specified):
```yaml
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

**Rationale**:
- Prevent resource starvation in local development
- Allow multiple components without overwhelming laptop
- Kubernetes best practices (requests for scheduling, limits for protection)
- Users can override for resource-intensive apps

**QoS Classes** (Kubernetes behavior):
- Guaranteed: requests == limits (predictable performance)
- Burstable: requests < limits (default for custom components)
- BestEffort: no requests/limits (not recommended, but supported)

### 9. Testing Strategy

#### Decision: Multi-Level Testing Pyramid

**Unit Tests** (80% coverage target):
- Configuration parsing and validation
- Struct marshaling/unmarshaling
- Error message generation
- Port conflict detection
- Resource format validation

**Integration Tests**:
- End-to-end deployment with test cluster
- kubectl command execution
- Secret reference resolution
- Multi-component deployment
- Cleanup verification

**Table-Driven Tests**:
```go
func TestCustomComponentValidation(t *testing.T) {
    tests := []struct {
        name        string
        component   CustomComponent
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid minimal config",
            component: CustomComponent{
                Name:    "test-app",
                Enabled: true,
                Image:   "nginx:latest",
            },
            expectError: false,
        },
        {
            name: "missing image",
            component: CustomComponent{
                Name:    "test-app",
                Enabled: true,
            },
            expectError: true,
            errorMsg:    "image is required",
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.component.Validate()
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Benchmark Tests**:
- Configuration parsing performance
- Validation performance for large configs
- Deployment generation performance

### 10. Documentation and Examples

#### Decision: Comprehensive Quickstart with Progressive Examples

**Documentation Structure**:
1. Minimal example (image only)
2. Environment variables example (direct values)
3. Secret references example (database connection)
4. Full-featured example (all options)
5. Multi-component example
6. Troubleshooting guide

**Example Application**:
- Simple Spring Boot app (realistic use case)
- Demonstrates MySQL and OpenSearch connectivity
- Includes Dockerfile and build instructions
- Shows health check endpoints
- Documents environment variable usage

## Best Practices Applied

### From Existing Codebase
1. **Progressive Enhancement**: Start simple, add complexity gradually
2. **Error Context**: Wrap errors with fmt.Errorf for debugging
3. **Color Output**: Use fatih/color for visual feedback
4. **Validation First**: Validate before executing kubectl commands
5. **Idempotent Operations**: Safe to run multiple times

### From Kubernetes Ecosystem
1. **Declarative Configuration**: YAML-based, not imperative commands
2. **Label Selectors**: Proper labeling for resource management
3. **Namespace Isolation**: Support namespace-level separation
4. **Resource Quotas**: Default limits to prevent resource exhaustion
5. **Health Checks**: Readiness/liveness probes (future enhancement)

### From Go/Cobra Conventions
1. **Struct Composition**: Embed common fields, compose complex types
2. **Interface Abstraction**: Define interfaces for testability
3. **Table-Driven Tests**: Comprehensive test coverage
4. **Context Propagation**: Use context.Context for cancellation
5. **Error Wrapping**: Preserve error chain for debugging

## Technology Choices Summary

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Deployment | Kubernetes Deployment | Production-like, auto-restart, rolling updates |
| Service | NodePort Service | Simple host access, consistent with existing components |
| Config Format | YAML (gopkg.in/yaml.v3) | Existing infrastructure, familiar to users |
| Secrets | Kubernetes Secrets | Native, secure, mirrors production |
| Validation | Multi-phase (config/pre-deploy/runtime) | Early error detection, clear feedback |
| Testing | Go testing + testify | TDD approach, high coverage |
| Error Handling | Wrapped errors (fmt.Errorf) | Context preservation, debugging |
| CLI Output | fatih/color | Existing pattern, clear visual feedback |

### 11. Configuration File Mounting

#### Decision: Use Kubernetes ConfigMaps with Volume Mounts

**Rationale**:
- ConfigMaps are designed for configuration data (non-sensitive)
- Volume mounts allow file-level granularity
- Automatic pod updates when ConfigMap changes (with pod restart)
- Preserves file formatting and special characters
- Read-only mounts prevent accidental modification
- Supports multiple files per component

**Implementation Approach**:

```yaml
configFiles:
  - name: application.yaml
    path: /config/application.yaml
    contents: |
      server:
        port: 8080
      database:
        host: mysql
```

**Generated Kubernetes Resources**:

1. **ConfigMap** (one per component):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-app-config
  namespace: default
data:
  application.yaml: |
    server:
      port: 8080
    database:
      host: mysql
  logging.xml: |
    <configuration>...</configuration>
```

2. **Volume and VolumeMount** in Deployment:
```yaml
spec:
  volumes:
    - name: config-volume
      configMap:
        name: my-app-config
        defaultMode: 0644  # Read-only
  containers:
    - name: my-app
      volumeMounts:
        - name: config-volume
          mountPath: /config/application.yaml
          subPath: application.yaml
          readOnly: true
```

**Alternatives Considered**:

| Alternative | Pros | Cons | Why Rejected |
|------------|------|------|--------------|
| Secrets for all configs | Encrypted at rest | Slower, overkill for non-sensitive | ConfigMaps sufficient, secrets via env vars already supported |
| External file references | Large files easier | Complex validation, file sync issues | Inline keeps config self-contained |
| Init container + download | Supports remote configs | Complex, slow startup | Over-engineered for local dev |
| Bake into image | Fastest | Requires image rebuild for changes | Defeats purpose of config mounting |

**Mount Path Conflict Handling**:
- Kubernetes behavior: Volume mounts override image files/directories
- Implementation: Log warning when mount path detected in image
- Validation: Error if multiple config files target same path

**ConfigMap Size Limits**:
- Kubernetes limit: 1MB per ConfigMap
- Validation: Check total config file contents < 1MB
- Best practice: Keep config files small (<100KB each)

**Update Strategy**:
- ConfigMaps are immutable once created
- Update approach: Delete old ConfigMap + create new (atomic)
- Pod restart required for config changes (rollout restart)

## Open Questions (Resolved)

All initial unknowns have been resolved through research:

1. ✅ Deployment pattern → Kubernetes Deployments
2. ✅ Port exposure → NodePort Services
3. ✅ Secret management → Kubernetes SecretKeyRef
4. ✅ Configuration schema → Extended KindEnvConfig struct
5. ✅ Validation approach → Multi-phase validation
6. ✅ Deployment order → After infrastructure components
7. ✅ Status reporting → Extended kindenv status command
8. ✅ Registry support → Leverage existing ECR/Harbor integration
9. ✅ Resource limits → Default limits with override support
10. ✅ Testing strategy → Unit + integration + table-driven tests
11. ✅ Config file mounting → ConfigMaps with volume mounts (inline contents)

## Next Steps

Phase 0 Complete ✅

Proceed to **Phase 1: Design & Contracts**:
1. Create data-model.md with detailed Go struct definitions
2. Generate contracts/ directory with:
   - custom-component-schema.yaml (YAML validation schema)
   - custom-component-api-interface.go (Go interfaces)
   - deployment-template.yaml (Kubernetes manifest template)
3. Create quickstart.md with step-by-step examples
4. Update agent context with new patterns and dependencies
