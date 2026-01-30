# Implementation Plan: MySQL 8 Support for KindEnv

**Branch**: `001-mysql8-kindenv` | **Date**: 2026-01-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-mysql8-kindenv/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add MySQL 8 as a new component to kindenv that integrates with existing configuration patterns. Uses Bitnami MySQL Helm chart with bitnamilegacy image repository, reuses existing secrets.mysql structure for credentials, and provides NodePort service with Kind port mapping to expose MySQL on standard port 3306. Includes configurable resource limits, optional persistence, and health monitoring integration.

## Technical Context

**Language/Version**: Go 1.21+ (existing codebase)  
**Primary Dependencies**: Cobra CLI framework, Kubernetes client-go, Helm Go SDK, existing kindenv configuration system  
**Storage**: MySQL 8 via Bitnami Helm chart, Kubernetes Secrets, optional PersistentVolumes  
**Testing**: Go testing package, table-driven tests, integration tests with Kind cluster  
**Target Platform**: Local development environments (macOS, Linux, Windows) with Kind/Docker
**Project Type**: CLI extension - extends existing kindenv command structure  
**Performance Goals**: MySQL startup <2 minutes, configuration changes <30 seconds, 95% startup success rate  
**Constraints**: Reuse existing secrets.mysql structure, bitnamilegacy image repository, Kind port mapping to 3306  
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
- Progress indicators for MySQL installation operations
- Follows existing command naming pattern: `cmd/kindenv_start.go`

### Command Design Standards ✅
- MySQL installation is idempotent (safe to run multiple times)
- Integrates with existing `--verbose` mode for detailed logging
- Consistent exit codes following existing patterns
- Help text will include practical examples following existing patterns

### Error Handling & User Feedback ✅
- Error messages include context and suggested solutions
- Integrates with existing structured logging patterns
- Uses existing color-coded output libraries (fatih/color)
- Progress indicators for operations taking more than 2 seconds
- Clear success/failure indicators with next steps guidance

**GATE STATUS**: ✅ PASS - All constitutional requirements satisfied

### Post-Design Constitution Re-Check ✅

After completing Phase 1 design and contracts:

**Code Quality Standards**: ✅ Confirmed
- Interface contracts defined in `contracts/mysql-api-interface.go` follow Go conventions
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
- MySQL operations are idempotent (safe to run multiple times)
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
specs/[###-feature]/
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
├── kindenv_start.go          # Extend with MySQL installation logic
├── kindenv_status.go         # Extend with MySQL health monitoring
├── kindenv_init.go           # Extend with MySQL Helm repository setup
└── kindenv_start_test.go     # Add MySQL-specific test cases

internal/kindenv/
├── config.go                 # Extend with MySQL component configuration
├── config_test.go            # Add MySQL configuration validation tests
└── mysql.go                  # NEW: MySQL-specific installation and management logic

examples/
└── kindenv.yaml              # Update with MySQL configuration examples

docs/
└── KINDENV.md                # Update with MySQL usage documentation
```

**Structure Decision**: Extends existing CLI structure following established patterns. MySQL functionality integrates into existing kindenv command files rather than creating separate commands, maintaining consistency with Redis and Temporal components. New `mysql.go` file in `internal/kindenv/` package contains MySQL-specific logic following the existing pattern.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
