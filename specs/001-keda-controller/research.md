# Research: KEDA Controller Integration

**Feature**: 001-keda-controller  
**Date**: 2026-02-10  
**Researcher**: Implementation Planning Phase

## Executive Summary

KEDA (Kubernetes Event-Driven Autoscaling) is a lightweight component that extends Kubernetes to provide event-driven autoscaling capabilities. This research validates the approach to integrate KEDA into the kindenv environment following established patterns from similar components (MetricsServer, Temporal Worker Operator).

**Key Finding**: KEDA should follow the **MetricsServer pattern** (simple component) rather than MySQL/RabbitMQ pattern (complex component with secrets/persistence).

---

## KEDA Overview

### What is KEDA?

KEDA is a Kubernetes-based Event Driven Autoscaler that enables fine-grained autoscaling for event-driven workloads. Key capabilities:

- Scale applications based on external event sources (queues, streams, metrics)
- Scale to zero when no events are present
- Scale out based on event queue depth, lag, or custom metrics
- Works with any Kubernetes workload (Deployments, StatefulSets, Jobs)
- Supports 50+ event sources (scalers) including RabbitMQ, Kafka, Azure Queue, AWS SQS, Prometheus, etc.

### Architecture Components

1. **KEDA Operator**: Manages ScaledObject and ScaledJob CRDs, creates HPA resources
2. **Metrics Server**: Exposes external metrics to Kubernetes Metrics API
3. **Admission Webhooks**: Validates ScaledObject and ScaledJob resources (optional)

### Namespace Convention

KEDA documentation and community best practices recommend using a dedicated `keda` namespace for isolation and organization.

---

## Installation Research

### Official Helm Chart Details

**Repository**: `https://kedacore.github.io/charts`  
**Chart Name**: `kedacore/keda`  
**Current Stable Version**: 2.19.0 (as of 2026-02-10)  
**Recommended Version for Production**: 2.16.0 (widely tested, stable)

### Installation Command

```bash
# Add repository
helm repo add kedacore https://kedacore.github.io/charts
helm repo update

# Install KEDA
helm install keda kedacore/keda \
  --namespace keda \
  --create-namespace
```

### Version Management

- Helm chart manages CRDs automatically (since v2.2.1)
- No separate CRD installation required
- Upgrading is handled via `helm upgrade --install` (idempotent)

### Configuration Requirements

**Minimal**: KEDA works with default Helm values for local development

**No Additional Requirements**:
- No secrets needed (unlike MySQL/RabbitMQ)
- No persistent storage (unlike MySQL/RabbitMQ)
- No NodePorts needed (internal cluster component)
- No init scripts or custom configuration

---

## Component Pattern Analysis

### Pattern Comparison Matrix

| Feature | MetricsServer | RabbitMQ | MySQL | KEDA (Proposed) |
|---------|---------------|----------|-------|-----------------|
| Config Complexity | Low | High | High | **Low** |
| Dedicated Namespace | No (kube-system) | Yes | Yes | **Yes (keda)** |
| Secrets Required | No | Yes | Yes | **No** |
| Persistence | No | Yes | Yes | **No** |
| NodePorts | No | Yes | Yes | **No** |
| Image Registry Config | No | Yes | Yes | **No** |
| Init Scripts | No | No | Yes | **No** |
| Custom Helm Values | Minimal | Extensive | Extensive | **Minimal** |

### Pattern Selection: MetricsServer

**Decision**: KEDA follows the **MetricsServer pattern**

**Rationale**:
1. Both are system-level Kubernetes extensions
2. Both provide metrics/scaling capabilities
3. Both require minimal configuration
4. Both are optional components
5. Both run in dedicated system namespaces

**Configuration Structure** (similar to MetricsServer):
```go
Keda struct {
    Enabled      bool   `yaml:"enabled"`
    Namespace    string `yaml:"namespace"`
    ChartVersion string `yaml:"chartVersion"`
} `yaml:"keda"`
```

---

## Existing Component Integration Patterns

### Pattern 1: MetricsServer (cmd/kindenv_start.go, lines 692-750)

**Characteristics**:
- Simple Helm install with single `--set` flag
- System namespace usage
- Non-blocking error handling
- 2-minute wait for deployment readiness
- Clear user feedback with next steps

**Key Code Pattern**:
```go
if config.Components.MetricsServer.Enabled {
    fmt.Println(yellow("Installing Metrics Server"))
    
    helmArgs := []string{
        "upgrade", "--install",
        "metrics-server", "metrics-server/metrics-server",
        "--namespace", "kube-system",
        "--version", config.Components.MetricsServer.ChartVersion,
        "--set", "args={--kubelet-insecure-tls}",
    }
    
    helmOutput, err := executeCommand("helm", helmArgs...)
    if err != nil {
        fmt.Printf("%s Error installing: %v\n", red("❌"), err)
        fmt.Println(yellow("Continuing despite failure..."))
    } else {
        fmt.Printf("%s Installed successfully\n", green("✅"))
        err = waitForDeployment("kube-system", "metrics-server", 2)
        // ... user guidance ...
    }
}
```

### Pattern 2: MySQL (cmd/kindenv_start.go, lines 894-1050)

**Characteristics** (NOT applicable to KEDA):
- Complex namespace creation with ECR credential setup
- Secret creation in component namespace
- ConfigMap for init scripts with file path resolution
- Extensive Helm customization (15+ --set flags)
- Resource limits, persistence configuration
- NodePort service setup

**Why Not This Pattern**: KEDA doesn't need any of these features

### Pattern 3: Repository Setup (cmd/kindenv_init.go, lines 330-350)

**Standard Pattern for Adding Helm Repos**:
```go
fmt.Println("Adding <Component> Helm repository...")
_, err = executeCommandWithOutput("helm", "repo", "add", "repo-name", "repo-url")
if err != nil {
    fmt.Printf("⚠️  Warning: Failed to add repository: %v\n", err)
} else {
    fmt.Println("✅ Repository added")
}

// Verify chart availability
output, err := executeCommandWithOutput("helm", "search", "repo", "repo-name/chart")
if err != nil || !strings.Contains(output, "repo-name/chart") {
    fmt.Printf("⚠️  Warning: Chart not found\n")
} else {
    fmt.Println("✅ Chart is available")
}
```

---

## Status Check Patterns

### Current Implementation (cmd/kindenv_status.go)

**MetricsServer Status Check** (lines 250-270):
```go
if config.Components.MetricsServer.Enabled {
    cmd := exec.Command("kubectl", "get", "deployment", 
        "-n", "kube-system", "metrics-server", "--no-headers")
    output, err := cmd.CombinedOutput()
    
    if err != nil || len(output) == 0 {
        fmt.Printf("%s Not found\n", red("❌"))
    } else if strings.Contains(string(output), "1/1") {
        fmt.Printf("%s Running\n", green("✅"))
    } else {
        fmt.Printf("%s Not ready\n", yellow("⚠️"))
    }
    
    if verbose {
        fmt.Printf("Output: %s\n", string(output))
    }
}
```

**KEDA Status Check** (proposed):
- Check deployment: `keda-operator` in `keda` namespace
- Check deployment: `keda-metrics-apiserver` in `keda` namespace (optional)
- Follow same output formatting pattern

---

## Configuration Best Practices

### Default Values Research

**Enabled**: `false` (opt-in)
- **Rationale**: KEDA is optional; users should explicitly enable it
- **Consistent with**: Most other optional components (Temporal, Redis, Dapr)
- **Exception**: Core components like OpenSearch/MySQL are enabled by default

**Namespace**: `keda`
- **Rationale**: KEDA documentation recommends dedicated namespace
- **Consistent with**: Other dedicated namespaces (temporal, opensearch, mysql)
- **Alternative considered**: `kube-system` - rejected because KEDA is third-party

**ChartVersion**: `2.16.0`
- **Rationale**: Stable, widely-tested version with known behavior
- **Alternative**: `2.19.0` (latest) - rejected for stability preference
- **Alternative**: No default - rejected because all components have version defaults

### Configuration File Location

Update `kindenv.yaml` in components section (after metricsServer):

```yaml
components:
  # ... existing components ...
  
  metricsServer:
    enabled: true
    chartVersion: 3.10.0
  
  keda:
    enabled: false
    namespace: keda
    chartVersion: 2.16.0
  
  mysql:
    enabled: true
    # ...
```

---

## Error Handling & User Experience

### Non-Blocking Installation

**Pattern from All Components**:
```go
if err != nil {
    fmt.Printf("%s Error installing KEDA: %v\n", red("❌"), err)
    if helmOutput != "" {
        fmt.Println("Output:")
        fmt.Println(helmOutput)
    }
    fmt.Println(yellow("Continuing despite KEDA installation failure..."))
} else {
    fmt.Printf("%s KEDA installed successfully\n", green("✅"))
    // ... wait for readiness ...
}
```

**Rationale**: 
- KEDA is optional
- Other components should not be blocked
- User can troubleshoot and retry

### User Feedback Guidelines

**Success Path**:
1. "Installing KEDA" (yellow)
2. Helm command execution (verbose only)
3. "✅ KEDA installed successfully" (green)
4. "Waiting for KEDA to be ready..." (yellow)
5. "✅ KEDA operator is ready" (green)
6. User guidance: "You can now create ScaledObject and ScaledJob resources"

**Failure Path**:
1. "Installing KEDA" (yellow)
2. "❌ Error installing KEDA: <error>" (red)
3. Helm output (if available)
4. "⚠️ Continuing despite KEDA installation failure..." (yellow)

---

## Testing Strategy

### Unit Test Patterns (from cmd/kindenv_start_test.go)

**Table-Driven Tests**:
```go
func TestKedaConfiguration(t *testing.T) {
    tests := []struct {
        name           string
        config         *kindenv.KindEnvConfig
        expectedError  bool
    }{
        {
            name: "default KEDA configuration",
            config: &kindenv.KindEnvConfig{
                // ... config with KEDA enabled ...
            },
            expectedError: false,
        },
        // ... more test cases ...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            // ... assertions ...
        })
    }
}
```

### Integration Test Scenarios

1. **Happy Path**: Enable KEDA → start environment → verify running
2. **Skip Flag**: Use --skip-keda → verify KEDA not installed
3. **Error Recovery**: Break repo access → verify graceful failure → fix → retry
4. **Version Pinning**: Configure specific version → verify correct version installed
5. **ScaledObject Creation**: Install KEDA → create ScaledObject → verify acceptance

---

## Security & Permissions

### No Additional Security Requirements

KEDA runs with standard Kubernetes service account permissions:
- Read access to resources it scales (Deployments, StatefulSets)
- Write access to HPA resources
- CRD management (handled by Helm)

### No Secrets Required

Unlike MySQL/RabbitMQ:
- No database passwords
- No authentication credentials
- No TLS certificates

**Exception**: If scalers require credentials (e.g., RabbitMQ connection string), those are configured in ScaledObject resources by users, not during KEDA installation.

---

## Performance Considerations

### Installation Time

**Estimated Time**: 2-3 minutes
- Namespace creation: ~1 second
- Helm install: 30-60 seconds
- Pod startup: 1-2 minutes
- Total: < 3 minutes typical

**Timeout Settings**:
- Wait for deployment: 2 minutes (sufficient for KEDA)
- Helm timeout: default (5 minutes)

### Resource Usage

**KEDA Operator** (default limits):
- CPU: 100m-1000m
- Memory: 128Mi-1Gi

**KEDA Metrics Server** (default limits):
- CPU: 100m-1000m
- Memory: 128Mi-1Gi

**Impact on Kind Cluster**: Minimal - comparable to MetricsServer

---

## Alternatives Considered

### Alternative 1: YAML Manifests (Not Helm)

**Rejected Because**:
- Inconsistent with all other components (all use Helm)
- Harder to maintain and version
- Helm provides better upgrade path
- CRD management is cleaner with Helm

### Alternative 2: Always Install (No Config)

**Rejected Because**:
- KEDA is optional functionality
- Not all users need autoscaling
- Consistent pattern: optional components are configurable
- Opt-in is better than opt-out for optional features

### Alternative 3: Complex Configuration Like MySQL

**Rejected Because**:
- KEDA doesn't need secrets, persistence, or NodePorts
- Over-engineering for simple use case
- Harder to maintain
- Inconsistent with MetricsServer (similar component)

### Alternative 4: Latest Version as Default

**Rejected Because**:
- Stability over novelty for default configuration
- Version 2.16.0 is well-tested and stable
- Users can override if they need latest features
- Consistent with other components (use stable versions)

---

## Dependencies & Compatibility

### Kubernetes Version Requirements

**KEDA Requirement**: Kubernetes 1.30+  
**Kind Environment**: Already requires 1.30+  
**Compatibility**: ✅ No issues

### Helm Version Requirements

**KEDA Requirement**: Helm 3.x  
**Kind Environment**: Helm 3.x already available  
**Compatibility**: ✅ No issues

### Component Dependencies

**KEDA Has No Dependencies**:
- Runs independently of other components
- Does not require Temporal, Redis, MySQL, etc.
- Can be installed before or after other components

**Other Components Don't Depend on KEDA**:
- Optional for autoscaling use cases
- No breaking changes to existing functionality

---

## Documentation Requirements

### User-Facing Documentation

1. **Quickstart Guide** (`specs/001-keda-controller/quickstart.md`):
   - Overview of KEDA capabilities
   - Configuration instructions
   - Example ScaledObject for RabbitMQ
   - Troubleshooting common issues

2. **README Update**:
   - Add KEDA to component list
   - Brief description of autoscaling capabilities
   - Link to quickstart guide

3. **Configuration Example** (`kindenv.yaml`):
   - Add KEDA section with comments
   - Show enabled and disabled states

### Developer Documentation

1. **Code Comments**:
   - Godoc comments for config struct
   - Inline comments for Helm arguments
   - Error handling rationale

2. **Test Documentation**:
   - Test case descriptions
   - Expected behaviors
   - Edge cases covered

---

## Risk Assessment

### Risk 1: Helm Repository Availability

**Likelihood**: Low  
**Impact**: Medium (installation fails)  
**Mitigation**: 
- Non-blocking error handling
- Clear error message with repository URL
- User can manually add repository and retry

### Risk 2: Version Incompatibility

**Likelihood**: Low  
**Impact**: Low (user can update version)  
**Mitigation**:
- Default to stable, well-tested version
- Allow version override in configuration
- Helm will report specific errors

### Risk 3: CRD Conflicts

**Likelihood**: Very Low  
**Impact**: Medium (installation fails)  
**Mitigation**:
- Helm manages CRDs automatically (v2.2.1+)
- `helm upgrade --install` handles updates
- Clear error message directing to KEDA docs

### Risk 4: Resource Exhaustion

**Likelihood**: Very Low  
**Impact**: Low (Kind cluster slowdown)  
**Mitigation**:
- KEDA has small resource footprint
- Default limits are reasonable for local dev
- User can disable if needed

---

## Conclusion

**Research Status**: ✅ Complete

**Key Findings**:
1. KEDA should follow the MetricsServer integration pattern (simple component)
2. Minimal configuration required (Enabled, Namespace, ChartVersion)
3. No secrets, persistence, or NodePorts needed
4. Helm chart manages all complexity including CRDs
5. Non-blocking installation consistent with other optional components

**Confidence Level**: High - Clear integration path with established patterns

**Ready to Proceed**: Yes - All technical questions resolved

**Next Phase**: Design & Implementation (Phase 1)