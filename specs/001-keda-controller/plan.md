# Implementation Plan: KEDA Controller Integration

**Branch**: `001-keda-controller` | **Date**: 2026-02-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-keda-controller/spec.md`

## Summary

Add KEDA (Kubernetes Event-Driven Autoscaling) controller to the kindenv environment to enable developers to test event-driven autoscaling functionality locally. KEDA will be integrated following the same pattern as existing Helm-based components (Metrics Server, Temporal Worker Operator, RabbitMQ, MySQL) with configuration-based enable/disable, version management, namespace setup, and status monitoring.

## Technical Context

**Language/Version**: Go 1.23+ (current toolchain: 1.24.1)  
**Primary Dependencies**: 
- Kubernetes client-go (existing)
- Helm Go SDK (existing)
- Cobra CLI framework (existing)
- KEDA Helm chart from `kedacore/charts` repository

**Storage**: Configuration in `kindenv.yaml` (YAML), Kubernetes cluster state  
**Testing**: Go testing package with table-driven tests, existing test patterns from `cmd/kindenv_start_test.go`  
**Target Platform**: macOS, Linux (development environments with Kind)  
**Project Type**: CLI tool (single project structure)  
**Performance Goals**: 
- KEDA installation completes within 5 minutes on standard development hardware
- Status checks respond within 2 seconds
- Environment continues to start even if KEDA installation fails (non-blocking)

**Constraints**: 
- Must follow existing component integration patterns (RabbitMQ, MySQL, Metrics Server)
- Must be idempotent and safe to run multiple times
- Must support command-line flag override (--skip-keda)
- Must not break existing functionality

**Scale/Scope**: 
- Single component integration (~12 functional requirements)
- Affects 4 files: config.go, kindenv_init.go, kindenv_start.go, kindenv_status.go
- Approximately 200-300 lines of new code plus tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Core Principles Compliance

✅ **Code Quality Standards**: 
- Will follow Go style guide and pass `go fmt`, `go vet`, `golangci-lint`
- Will use clear, self-documenting code with godoc comments
- Error handling will use `fmt.Errorf` for wrapped errors
- Will reuse existing patterns from component installations (no duplication)

✅ **Test-Driven Development**:
- Will write table-driven tests following existing patterns in `cmd/kindenv_start_test.go`
- Will achieve minimum 80% code coverage for new component logic
- Will include integration tests for KEDA installation flow
- Will follow Red-Green-Refactor cycle

✅ **Cobra CLI Best Practices**:
- Consistent with existing command structure: `devhelper-cli kindenv start`
- Will add `--skip-keda` flag following existing `--skip-*` pattern
- Will provide clear error messages with context
- Will use color-coded output (yellow for progress, green for success, red for errors)
- Status output will be consistent with existing component status display

✅ **Command Design Standards**:
- Installation is idempotent (safe to run multiple times via Helm upgrade --install)
- No dry-run needed (Helm manages this)
- Verbose mode supported via existing `--verbose` flag
- Non-destructive: installation failure doesn't break environment
- Exit codes consistent with existing patterns

✅ **Error Handling & User Feedback**:
- Error messages include context and suggested solutions
- Progress indicators for long-running operations
- Color-coded output using existing helper functions (yellow, green, red)
- Clear success/failure indicators with next steps

### Constitution Compliance Summary

**Status**: ✅ FULL COMPLIANCE - No violations or justifications needed

This feature follows established patterns and requires no architectural changes. All code will integrate seamlessly with existing Cobra CLI structure, error handling, and testing frameworks.

## Project Structure

### Documentation (this feature)

```text
specs/001-keda-controller/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (implementation plan)
├── research.md          # Phase 0 output (KEDA integration patterns)
├── data-model.md        # Phase 1 output (Configuration structure)
├── quickstart.md        # Phase 1 output (Developer guide)
├── contracts/           # Phase 1 output (Helm values and K8s manifests)
│   └── keda-config.yaml # Example KEDA configuration
└── checklists/
    └── requirements.md  # Quality checklist (completed)
```

### Source Code (repository root)

```text
devhelper-cli/
├── cmd/
│   ├── kindenv_init.go       # Add KEDA Helm repository
│   ├── kindenv_start.go      # Install KEDA controller
│   ├── kindenv_start_test.go # Add KEDA installation tests
│   └── kindenv_status.go     # Add KEDA status checks
├── internal/
│   └── kindenv/
│       └── config.go         # Add KEDA configuration struct
├── kindenv.yaml              # Example with KEDA configuration
└── examples/
    └── custom-component-app/
        └── kindenv.yaml      # Example with KEDA
```

**Structure Decision**: Single project structure maintained. KEDA integrates as a new component following the established pattern used by MetricsServer, TemporalWorkerOperator, IndicesOperator, MySQL, and RabbitMQ.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations to track. This feature follows all constitutional requirements and established patterns.

---

## Phase 0: Research & Technical Decisions

### Research Topics

1. **KEDA Helm Repository and Chart Details**
   - Repository URL: `https://kedacore.github.io/charts`
   - Chart name: `kedacore/keda`
   - Default namespace: `keda`
   - Latest stable version: 2.19.0 (as of research date)

2. **KEDA Installation Patterns**
   - Helm installation command: `helm install keda kedacore/keda --namespace keda --create-namespace`
   - Verification: Check pods with label `app.kubernetes.io/name=keda-operator` in `keda` namespace
   - CRDs are managed by Helm chart automatically (v2.2.1+)
   - No additional secrets or authentication required for basic installation

3. **Component Integration Pattern Analysis**
   - **MetricsServer pattern** (simplest, similar to KEDA):
     - Config: `Enabled` (bool), `ChartVersion` (string)
     - Namespace: `kube-system` (system namespace)
     - No NodePorts, no secrets, no persistence
     - Installation: Single Helm install with `--set args={--kubelet-insecure-tls}`
     - Status: Check deployment in namespace
   
   - **MySQL/RabbitMQ pattern** (complex, with secrets and persistence):
     - Config: Enabled, Namespace, ChartVersion, NodePorts, Resources, Persistence, InitScripts
     - Creates dedicated namespace
     - Creates secrets for authentication
     - Supports ECR/Harbor image registries
     - Extensive Helm customization
   
   - **KEDA should follow MetricsServer pattern**: Simple component, no secrets, no persistence, minimal configuration

4. **Status Check Patterns**
   - Use `kubectl get deployment -n <namespace>` to check deployment status
   - Parse output to determine if component is running
   - Display in verbose mode with additional details
   - Consistent formatting with other components

5. **Configuration Structure Pattern**
   ```go
   Keda struct {
       Enabled      bool   `yaml:"enabled"`
       Namespace    string `yaml:"namespace"`
       ChartVersion string `yaml:"chartVersion"`
   } `yaml:"keda"`
   ```

### Key Decisions

**Decision 1: Follow MetricsServer Pattern**
- **Rationale**: KEDA is a system-level controller like MetricsServer, not an application service like MySQL/RabbitMQ. It requires minimal configuration and no secrets.
- **Alternatives considered**: 
  - RabbitMQ/MySQL complex pattern: Rejected because KEDA doesn't need NodePorts, persistence, or secrets
  - Custom pattern: Rejected to maintain consistency

**Decision 2: Use Dedicated `keda` Namespace**
- **Rationale**: KEDA documentation recommends dedicated namespace for isolation and organization
- **Alternatives considered**: 
  - `kube-system` namespace: Rejected because KEDA is optional/third-party, not core Kubernetes
  - `default` namespace: Rejected for better organization

**Decision 3: Chart Version Configuration**
- **Rationale**: Allow version pinning for environment parity with production
- **Default**: Use a stable version (2.16.0 - widely tested) rather than latest
- **Alternatives considered**: 
  - Always use latest: Rejected for stability and reproducibility
  - No version config: Rejected because version control is standard for other components

**Decision 4: Non-Blocking Installation**
- **Rationale**: KEDA is optional; its failure shouldn't block other components
- **Implementation**: Wrap installation in error handler that logs warning and continues
- **Consistent with**: All other optional components (Temporal, Redis, Dapr, etc.)

**Decision 5: Minimal Helm Customization**
- **Rationale**: KEDA default configuration works for local development
- **Implementation**: No custom `--set` flags except standard Helm practices
- **Alternatives considered**: 
  - Custom resource limits: Rejected because defaults are suitable for local Kind
  - Custom webhook configuration: Rejected because defaults are appropriate

---

## Phase 1: Design & Configuration

### Configuration Data Model

**File**: `internal/kindenv/config.go` (lines ~115-118, after MetricsServer)

```go
Keda struct {
    Enabled      bool   `yaml:"enabled"`
    Namespace    string `yaml:"namespace"`
    ChartVersion string `yaml:"chartVersion"`
} `yaml:"keda"`
```

**Default Values** (in `CreateDefaultConfig` function, ~line 960):

```go
config.Components.Keda.Enabled = false  // Opt-in by default
config.Components.Keda.Namespace = "keda"
config.Components.Keda.ChartVersion = "2.16.0"  // Stable, widely-tested version
```

**Validation Rules** (in `config.Validate` function, ~line 890):
- No specific validation needed (simple boolean and strings)
- Namespace must be valid Kubernetes namespace name (handled by kubectl)
- ChartVersion format validated by Helm during installation

### Command-Line Interface Changes

**File**: `cmd/kindenv_start.go`

**New Flag** (in `init()` function, ~line 2700):
```go
kindenvStartCmd.Flags().Bool("skip-keda", false, "Skip KEDA controller installation")
```

**Flag Processing** (in `Run:` function, ~line 150):
```go
skipKeda, _ := cmd.Flags().GetBool("skip-keda")
// ... after loading config ...
if skipKeda {
    config.Components.Keda.Enabled = false
}
```

**Status Display** (in configuration summary, ~line 230):
```go
fmt.Printf("- KEDA: %v\n", config.Components.Keda.Enabled)
```

### Installation Logic

**File**: `cmd/kindenv_start.go` (insert after MetricsServer installation, ~line 750)

```go
// Install KEDA
if config.Components.Keda.Enabled {
    fmt.Println(yellow("Installing KEDA"))

    // Create namespace
    namespaceYaml, err := executeCommand("kubectl", "create", "namespace", 
        config.Components.Keda.Namespace, "--dry-run=client", "-o", "yaml")
    if err != nil {
        fmt.Printf("%s Error creating KEDA namespace: %v\n", red("❌"), err)
        fmt.Println(yellow("Continuing despite KEDA namespace creation failure..."))
    } else {
        cmd := exec.Command("kubectl", "apply", "-f", "-")
        cmd.Stdin = strings.NewReader(namespaceYaml)
        if err := cmd.Run(); err != nil {
            fmt.Printf("%s Error applying KEDA namespace: %v\n", red("❌"), err)
            fmt.Println(yellow("Continuing despite KEDA namespace apply failure..."))
        }
    }

    // Define Helm arguments
    helmArgs := []string{
        "upgrade",
        "--install",
        "keda", "kedacore/keda",
        "--namespace", config.Components.Keda.Namespace,
        "--version", config.Components.Keda.ChartVersion,
        "--create-namespace",
    }

    // Execute Helm command
    if verbose {
        fmt.Printf("Command: helm %s\n", strings.Join(helmArgs, " "))
    }

    helmOutput, err := executeCommand("helm", helmArgs...)
    if err != nil {
        fmt.Printf("%s Error installing KEDA: %v\n", red("❌"), err)
        if helmOutput != "" {
            fmt.Println("Output:")
            fmt.Println(helmOutput)
        }
        fmt.Println(yellow("Continuing despite KEDA installation failure..."))
    } else {
        fmt.Printf("%s KEDA installed successfully\n", green("✅"))

        // Wait for KEDA to be ready
        fmt.Println(yellow("Waiting for KEDA to be ready..."))

        // Wait a moment for resources to be created
        time.Sleep(5 * time.Second)

        err = waitForDeployment(config.Components.Keda.Namespace, "keda-operator", 2)
        if err != nil {
            fmt.Printf("%s Error waiting for KEDA operator: %v\n", red("❌"), err)
            fmt.Println(yellow("Continuing despite KEDA not being ready..."))
        } else {
            fmt.Printf("%s KEDA operator is ready\n", green("✅"))
            fmt.Println(yellow("You can now create ScaledObject and ScaledJob resources"))
            fmt.Println("  Learn more: https://keda.sh/docs/latest/concepts/scaling-deployments/")
        }
    }
}
```

### Status Check Logic

**File**: `cmd/kindenv_status.go` (insert after MetricsServer status, ~line 260)

```go
// Check KEDA status
if config.Components.Keda.Enabled {
    kedaCmd := exec.Command("kubectl", "get", "deployment", "-n", 
        config.Components.Keda.Namespace, "keda-operator", "--no-headers")
    kedaOutput, err := kedaCmd.CombinedOutput()
    if err != nil || len(kedaOutput) == 0 {
        fmt.Printf("%s KEDA operator: Not found\n", red("❌"))
        if verbose && len(kedaOutput) > 0 {
            fmt.Printf("Output: %s\n", string(kedaOutput))
        }
    } else {
        outputStr := string(kedaOutput)
        if strings.Contains(outputStr, "1/1") || strings.Contains(outputStr, "Running") {
            fmt.Printf("%s KEDA operator: Running\n", green("✅"))
        } else {
            fmt.Printf("%s KEDA operator: Not ready\n", yellow("⚠️"))
        }
        if verbose {
            fmt.Printf("Output: %s\n", outputStr)
        }
    }
}
```

### Repository Setup Logic

**File**: `cmd/kindenv_init.go` (add to Helm repository initialization, ~line 330)

```go
// Add KEDA Helm repository
fmt.Println("Adding KEDA Helm repository...")
_, err = executeCommandWithOutput("helm", "repo", "add", "kedacore", 
    "https://kedacore.github.io/charts")
if err != nil {
    fmt.Printf("⚠️  Warning: Failed to add KEDA Helm repository: %v\n", err)
} else {
    fmt.Println("✅ KEDA Helm repository added")
}

// Verify KEDA chart availability
kedaOutput, err := executeCommandWithOutput("helm", "search", "repo", "kedacore/keda")
if err != nil || !strings.Contains(kedaOutput, "kedacore/keda") {
    fmt.Printf("⚠️  Warning: KEDA chart not found\n")
} else {
    fmt.Println("✅ KEDA chart is available")
}
```

### Example Configuration

**File**: `kindenv.yaml` (add to components section)

```yaml
components:
  # ... existing components ...
  
  keda:
    enabled: false
    namespace: keda
    chartVersion: 2.16.0
```

---

## Phase 2: Testing Strategy

### Unit Tests

**File**: `cmd/kindenv_start_test.go`

**Test 1: Configuration Validation**
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
                Components: struct {
                    // ... other components ...
                    Keda struct {
                        Enabled      bool   `yaml:"enabled"`
                        Namespace    string `yaml:"namespace"`
                        ChartVersion string `yaml:"chartVersion"`
                    } `yaml:"keda"`
                }{
                    Keda: struct {
                        Enabled      bool
                        Namespace    string
                        ChartVersion string
                    }{
                        Enabled:      true,
                        Namespace:    "keda",
                        ChartVersion: "2.16.0",
                    },
                },
            },
            expectedError: false,
        },
        {
            name: "KEDA disabled",
            config: &kindenv.KindEnvConfig{
                Components: struct {
                    Keda struct {
                        Enabled      bool
                        Namespace    string
                        ChartVersion string
                    }
                }{
                    Keda: struct {
                        Enabled      bool
                        Namespace    string
                        ChartVersion string
                    }{
                        Enabled: false,
                    },
                },
            },
            expectedError: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.expectedError && err == nil {
                t.Error("Expected error but got none")
            }
            if !tt.expectedError && err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
        })
    }
}
```

**Test 2: Flag Override Behavior**
- Test that `--skip-keda` flag properly disables KEDA even when enabled in config
- Test that status command shows correct state after flag override

**Test 3: Installation Error Handling**
- Test that environment continues when KEDA installation fails
- Test error message format and content

### Integration Tests

1. **Full Installation Test**: Install KEDA in test Kind cluster, verify pods running
2. **Skip Flag Test**: Start environment with --skip-keda, verify KEDA not installed
3. **Status Check Test**: Install KEDA, run status command, verify output
4. **ScaledObject Test**: Install KEDA, create ScaledObject, verify acceptance

### Manual Testing Checklist

- [ ] Fresh cluster: Enable KEDA in config, run `kindenv start`, verify installation
- [ ] Skip flag: Run `kindenv start --skip-keda`, verify KEDA not installed
- [ ] Status check: Run `kindenv status` with KEDA enabled, verify output
- [ ] Verbose mode: Run with `--verbose`, verify detailed output
- [ ] Error handling: Break Helm repo access, verify graceful failure
- [ ] Version pinning: Try different chart versions, verify correct version installed
- [ ] ScaledObject creation: Create sample ScaledObject, verify it's accepted

---

## Phase 3: Documentation

### Quickstart Guide

Create `specs/001-keda-controller/quickstart.md` with:
- Overview of KEDA and event-driven autoscaling
- Configuration instructions for kindenv.yaml
- Example ScaledObject for RabbitMQ queue autoscaling
- Common troubleshooting scenarios
- Links to KEDA documentation

### README Updates

Update main README.md to include:
- KEDA in list of available components
- Brief description of KEDA capabilities
- Link to KEDA quickstart guide

---

## Dependencies & Prerequisites

### External Dependencies
- KEDA Helm chart repository: `https://kedacore.github.io/charts`
- Chart version: 2.16.0 (default, configurable)
- Kubernetes 1.30+ (already required by Kind environment)
- Helm 3.x (already available in kindenv)

### Internal Dependencies
- Existing Helm integration in `kindenv_start.go`
- Existing kubectl command execution infrastructure
- Existing color output functions (yellow, green, red)
- Existing `waitForDeployment` function
- Existing `executeCommand` function

### No Breaking Changes
- All changes are additive
- Default behavior: KEDA disabled (opt-in)
- No impact on existing components
- Backward compatible with existing configurations

---

## Success Criteria Validation

Mapping specification success criteria to implementation:

- **SC-001**: ✅ Developers can enable KEDA in kindenv.yaml → Implemented in config.go
- **SC-002**: ✅ Status shows KEDA within 2 minutes → Implemented in kindenv_status.go
- **SC-003**: ✅ Pods ready within 3 minutes → Uses waitForDeployment with timeout
- **SC-004**: ✅ Can create ScaledObject resources → KEDA CRDs installed automatically
- **SC-005**: ✅ Environment continues on failure → Non-blocking error handling
- **SC-006**: ✅ Clear validation feedback → Error messages with context
- **SC-007**: ✅ Installation within 5 minutes → Helm + wait < 5 minutes typical

---

## Implementation Sequence

1. **Phase 0 - Research** (Completed above)
   - Research KEDA Helm installation patterns
   - Analyze existing component integration patterns
   - Document key decisions

2. **Phase 1 - Configuration** 
   - Add KEDA struct to `config.go`
   - Add default values to `CreateDefaultConfig`
   - Update example `kindenv.yaml`

3. **Phase 2 - Repository Setup**
   - Add KEDA Helm repository in `kindenv_init.go`
   - Add chart availability verification

4. **Phase 3 - Installation**
   - Add KEDA installation logic to `kindenv_start.go`
   - Add --skip-keda flag
   - Add status display in configuration summary

5. **Phase 4 - Status Monitoring**
   - Add KEDA status check in `kindenv_status.go`
   - Follow existing status display patterns

6. **Phase 5 - Testing**
   - Write unit tests for configuration
   - Write integration tests for installation
   - Manual testing per checklist

7. **Phase 6 - Documentation**
   - Create quickstart guide
   - Update README
   - Add example configurations

---

## Risk Assessment & Mitigation

### Risk 1: Helm Repository Unreachable
- **Impact**: Medium - Installation fails if repository unavailable
- **Mitigation**: Non-blocking error handling, clear error message with resolution steps
- **Fallback**: User can manually add repository and retry

### Risk 2: Chart Version Incompatibility
- **Impact**: Low - Specific chart version may not exist or be incompatible
- **Mitigation**: Default to widely-tested stable version (2.16.0), provide clear error
- **Fallback**: User can update chart version in configuration

### Risk 3: CRD Conflicts
- **Impact**: Low - Pre-existing KEDA CRDs could cause conflicts
- **Mitigation**: Helm upgrade --install handles CRD updates automatically (v2.2.1+)
- **Fallback**: Clear error message directing user to KEDA troubleshooting guide

### Risk 4: Namespace Conflicts
- **Impact**: Low - `keda` namespace may already exist with different resources
- **Mitigation**: Use --create-namespace flag, kubectl apply is idempotent
- **Fallback**: Allow user to specify alternate namespace in configuration

---

## Timeline Estimate

- **Phase 0 (Research)**: ✅ Completed
- **Phase 1 (Configuration)**: 1 hour
- **Phase 2 (Repository Setup)**: 30 minutes
- **Phase 3 (Installation)**: 2 hours
- **Phase 4 (Status Monitoring)**: 1 hour
- **Phase 5 (Testing)**: 3 hours
- **Phase 6 (Documentation)**: 2 hours

**Total Estimated Time**: ~10 hours

---

## Next Steps

After this plan is approved:

1. Run `/speckit.tasks` to break down implementation into specific tasks
2. Begin implementation starting with Phase 1 (Configuration)
3. Follow TDD approach: write tests first, then implementation
4. Create PR with all changes, tests, and documentation
5. Conduct code review focusing on constitutional compliance
6. Merge and validate in real Kind environment

---

**Plan Status**: ✅ Complete and Ready for Implementation

All research is complete, patterns are established, and implementation details are specified. No blocking issues or NEEDS CLARIFICATION items remain. Ready to proceed to `/speckit.tasks` for task breakdown.