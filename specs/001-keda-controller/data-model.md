# Data Model: KEDA Controller Configuration

**Feature**: 001-keda-controller  
**Date**: 2026-02-10  
**Status**: Design Phase

## Overview

This document defines the data structures and configuration model for KEDA (Kubernetes Event-Driven Autoscaling) integration in the kindenv environment. The model follows the established pattern used by MetricsServer and other simple system components.

---

## Configuration Entities

### 1. Keda Configuration

**Entity**: `KedaConfig`  
**Location**: `internal/kindenv/config.go` (Components struct)  
**Purpose**: Controls KEDA controller installation and configuration

#### Go Struct Definition

```go
type KindEnvConfig struct {
    // ... other fields ...
    
    Components struct {
        // ... other components ...
        
        MetricsServer struct {
            Enabled      bool   `yaml:"enabled"`
            ChartVersion string `yaml:"chartVersion"`
        } `yaml:"metricsServer"`
        
        Keda struct {
            Enabled      bool   `yaml:"enabled"`
            Namespace    string `yaml:"namespace"`
            ChartVersion string `yaml:"chartVersion"`
        } `yaml:"keda"`
        
        MySQL struct {
            // ...
        } `yaml:"mysql"`
        
        // ... other components ...
    } `yaml:"components"`
}
```

#### YAML Representation

```yaml
components:
  # ... other components ...
  
  metricsServer:
    enabled: true
    chartVersion: 3.10.0
  
  keda:
    enabled: false              # Boolean: Enable/disable KEDA installation
    namespace: keda             # String: Kubernetes namespace for KEDA
    chartVersion: 2.16.0        # String: Helm chart version to install
  
  mysql:
    enabled: true
    # ...
```

#### Field Specifications

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | boolean | Yes | `false` | Enables KEDA controller installation |
| `namespace` | string | Yes | `"keda"` | Kubernetes namespace for KEDA components |
| `chartVersion` | string | Yes | `"2.16.0"` | KEDA Helm chart version |

#### Validation Rules

```go
// No explicit validation needed for KEDA configuration
// - enabled: boolean type ensures valid values (true/false)
// - namespace: validated by kubectl during namespace creation
// - chartVersion: validated by Helm during chart installation

// If validation were added, it would look like:
func (c *KindEnvConfig) validateKeda() error {
    if c.Components.Keda.Enabled {
        // Namespace must not be empty
        if strings.TrimSpace(c.Components.Keda.Namespace) == "" {
            return fmt.Errorf("KEDA namespace cannot be empty")
        }
        
        // Chart version must not be empty
        if strings.TrimSpace(c.Components.Keda.ChartVersion) == "" {
            return fmt.Errorf("KEDA chartVersion cannot be empty")
        }
        
        // Namespace must be valid Kubernetes name (lowercase alphanumeric + hyphen)
        if !isValidK8sName(c.Components.Keda.Namespace) {
            return fmt.Errorf("invalid KEDA namespace name: must be lowercase alphanumeric with hyphens")
        }
    }
    return nil
}
```

#### Default Values

```go
// In CreateDefaultConfig function (config.go, ~line 960)
config.Components.Keda.Enabled = false
config.Components.Keda.Namespace = "keda"
config.Components.Keda.ChartVersion = "2.16.0"
```

**Rationale for Defaults**:
- `enabled: false` - Opt-in model for optional component
- `namespace: "keda"` - KEDA community best practice
- `chartVersion: "2.16.0"` - Stable, widely-tested version

---

## Runtime Entities

### 2. KEDA Deployment State

**Entity**: KEDA Kubernetes Resources  
**Lifecycle**: Created during `kindenv start`, checked during `kindenv status`  
**Purpose**: Track KEDA controller runtime state

#### Kubernetes Resources Created

```yaml
# Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: keda  # From config.Components.Keda.Namespace

---
# Helm Release
# Managed by: helm upgrade --install keda kedacore/keda
# Release Name: keda
# Namespace: keda
# Chart: kedacore/keda
# Version: 2.16.0  # From config.Components.Keda.ChartVersion

---
# Deployments Created by Helm Chart
apiVersion: apps/v1
kind: Deployment
metadata:
  name: keda-operator
  namespace: keda
  labels:
    app.kubernetes.io/name: keda-operator
    app.kubernetes.io/part-of: keda
spec:
  replicas: 1
  # ... (managed by Helm)

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: keda-metrics-apiserver
  namespace: keda
  labels:
    app.kubernetes.io/name: keda-metrics-apiserver
    app.kubernetes.io/part-of: keda
spec:
  replicas: 1
  # ... (managed by Helm)

---
# Custom Resource Definitions (CRDs)
# Created automatically by Helm chart:
# - scaledobjects.keda.sh
# - scaledjobs.keda.sh
# - triggerauthentications.keda.sh
# - clustertriggerauthentications.keda.sh
```

#### State Tracking

```go
// Status check queries these resources:
type KedaDeploymentState struct {
    OperatorDeployment struct {
        Name      string // "keda-operator"
        Namespace string // "keda"
        Ready     bool   // 1/1 replicas ready
        Status    string // "Running" | "Pending" | "CrashLoopBackOff"
    }
    
    MetricsServerDeployment struct {
        Name      string // "keda-metrics-apiserver"
        Namespace string // "keda"
        Ready     bool   // 1/1 replicas ready
        Status    string // "Running" | "Pending" | "CrashLoopBackOff"
    }
}
```

---

## Configuration Flow

### 1. Configuration Loading

```
User edits kindenv.yaml
    ↓
LoadConfig() reads YAML
    ↓
Unmarshals into KindEnvConfig struct
    ↓
Validates configuration (optional for KEDA)
    ↓
Configuration available to commands
```

### 2. Installation Flow

```
kindenv start command
    ↓
Check config.Components.Keda.Enabled
    ↓ (if true)
Apply --skip-keda flag override
    ↓ (if not skipped)
Create namespace: config.Components.Keda.Namespace
    ↓
Execute helm upgrade --install
    └─ Chart: kedacore/keda
    └─ Version: config.Components.Keda.ChartVersion
    └─ Namespace: config.Components.Keda.Namespace
    ↓
Wait for deployment ready (2 minutes)
    ↓
Display success message
```

### 3. Status Check Flow

```
kindenv status command
    ↓
Check config.Components.Keda.Enabled
    ↓ (if true)
kubectl get deployment -n <namespace> keda-operator
    ↓
Parse output for ready state
    ↓
Display status (✅ Running | ⚠️ Not ready | ❌ Not found)
```

---

## Configuration Examples

### Example 1: KEDA Disabled (Default)

```yaml
components:
  keda:
    enabled: false
    namespace: keda
    chartVersion: 2.16.0
```

**Result**: KEDA not installed, no namespace created

### Example 2: KEDA Enabled with Defaults

```yaml
components:
  keda:
    enabled: true
    namespace: keda
    chartVersion: 2.16.0
```

**Result**: 
- Namespace `keda` created
- KEDA 2.16.0 installed
- Operator and metrics server running

### Example 3: KEDA with Custom Version

```yaml
components:
  keda:
    enabled: true
    namespace: keda
    chartVersion: 2.19.0  # Latest version
```

**Result**: KEDA 2.19.0 installed instead of default

### Example 4: KEDA with Custom Namespace

```yaml
components:
  keda:
    enabled: true
    namespace: autoscaling  # Custom namespace
    chartVersion: 2.16.0
```

**Result**: KEDA installed in `autoscaling` namespace

---

## Command-Line Overrides

### Flag Data Model

```go
type StartCommandFlags struct {
    // ... other flags ...
    SkipKeda bool  // --skip-keda flag
}

// Flag processing:
// 1. Load configuration from YAML
// 2. Apply flag overrides
// 3. Use final configuration

if skipKeda {
    config.Components.Keda.Enabled = false
}
```

### Override Precedence

```
Highest Priority: Command-line flags (--skip-keda)
    ↓
Middle Priority: Configuration file (kindenv.yaml)
    ↓
Lowest Priority: Default values (in code)
```

**Example**:
```bash
# YAML has: keda.enabled = true
# Command: kindenv start --skip-keda
# Result: KEDA is NOT installed (flag wins)
```

---

## Data Relationships

### Relationship Diagram

```
KindEnvConfig
    └── Components
        ├── MetricsServer (similar pattern)
        ├── Keda
        │   ├── Enabled ──────────────┐
        │   ├── Namespace             │
        │   └── ChartVersion          │
        ├── MySQL (different pattern) │
        └── RabbitMQ                  │
                                      │
                                      ▼
                        (if Enabled == true)
                                      │
                                      ▼
                            Kubernetes Resources
                                ├── Namespace (keda)
                                ├── HelmRelease (keda)
                                ├── Deployment (keda-operator)
                                ├── Deployment (keda-metrics-apiserver)
                                └── CRDs (scaledobjects, scaledjobs, etc.)
```

### Dependencies

**KEDA Has No Dependencies On**:
- MySQL
- RabbitMQ
- Temporal
- Redis
- Dapr
- OpenSearch

**Other Components Do Not Depend On KEDA**:
- KEDA is purely optional
- Applications can reference ScaledObjects only if KEDA is installed

**Helm Repository Dependency**:
- Repository URL: `https://kedacore.github.io/charts`
- Must be added during `kindenv init`
- Checked for availability before installation

---

## Comparison with Similar Components

### MetricsServer vs KEDA

| Aspect | MetricsServer | KEDA |
|--------|---------------|------|
| Config Fields | 2 (Enabled, ChartVersion) | 3 (Enabled, Namespace, ChartVersion) |
| Namespace | kube-system (fixed) | keda (configurable) |
| Secrets | None | None |
| Persistence | None | None |
| NodePorts | None | None |
| Complexity | Low | Low |

### MySQL vs KEDA

| Aspect | MySQL | KEDA |
|--------|-------|------|
| Config Fields | 8+ fields | 3 fields |
| Namespace | Configurable | Configurable |
| Secrets | Required | None |
| Persistence | Optional | None |
| NodePorts | Required | None |
| Complexity | High | Low |

**Pattern Choice**: KEDA follows MetricsServer (simple) not MySQL (complex)

---

## Type Safety & Validation

### Type Safety

```go
// All fields are strongly typed
type Keda struct {
    Enabled      bool   // Not string or interface{}
    Namespace    string // Not interface{}
    ChartVersion string // Not interface{}
}

// YAML unmarshaling validates types automatically
// Invalid: enabled: "yes"  → Error: cannot unmarshal string into bool
// Invalid: enabled: 1      → Error: cannot unmarshal int into bool
// Valid:   enabled: true   → Success
```

### YAML Tag Validation

```go
`yaml:"enabled"`      // YAML key must be "enabled"
`yaml:"namespace"`    // YAML key must be "namespace"
`yaml:"chartVersion"` // YAML key must be "chartVersion"
```

---

## Testing Data Models

### Test Configuration Factory

```go
func newTestKedaConfig(enabled bool, namespace string, version string) *kindenv.KindEnvConfig {
    config := &kindenv.KindEnvConfig{}
    config.Components.Keda.Enabled = enabled
    config.Components.Keda.Namespace = namespace
    config.Components.Keda.ChartVersion = version
    return config
}

// Usage in tests:
config := newTestKedaConfig(true, "keda", "2.16.0")
```

### Test Cases Matrix

| Test Case | Enabled | Namespace | ChartVersion | Expected Result |
|-----------|---------|-----------|--------------|-----------------|
| Default | false | "keda" | "2.16.0" | No installation |
| Enabled | true | "keda" | "2.16.0" | Install in keda namespace |
| Custom NS | true | "autoscaling" | "2.16.0" | Install in autoscaling |
| Custom Ver | true | "keda" | "2.19.0" | Install version 2.19.0 |
| Invalid NS | true | "" | "2.16.0" | Error (empty namespace) |
| Invalid Ver | true | "keda" | "" | Error (empty version) |

---

## Migration & Backward Compatibility

### Adding KEDA to Existing Configuration

**Before** (existing kindenv.yaml):
```yaml
components:
  metricsServer:
    enabled: true
    chartVersion: 3.10.0
  mysql:
    enabled: true
    # ...
```

**After** (with KEDA added):
```yaml
components:
  metricsServer:
    enabled: true
    chartVersion: 3.10.0
  
  keda:
    enabled: false        # Added, default to disabled
    namespace: keda
    chartVersion: 2.16.0
  
  mysql:
    enabled: true
    # ...
```

**Backward Compatibility**:
- ✅ Existing configurations without KEDA section work (defaults applied)
- ✅ Default `enabled: false` means no behavior change
- ✅ No breaking changes to existing components
- ✅ Additive change only

---

## Summary

**Data Model Complexity**: Low  
**Number of Configuration Fields**: 3  
**Number of Runtime Entities**: 4 (Namespace, HelmRelease, 2 Deployments)  
**Validation Complexity**: Minimal (type checking only)  
**Pattern Similarity**: MetricsServer (90% similar)  

**Key Design Decisions**:
1. Simple three-field configuration (Enabled, Namespace, ChartVersion)
2. No secrets or persistence needed
3. Strongly-typed Go structs with YAML tags
4. Default disabled (opt-in model)
5. Configurable namespace (but defaults to `keda`)
6. Stable version default (2.16.0)

**Ready for Implementation**: ✅ Yes - Clear data model with established patterns