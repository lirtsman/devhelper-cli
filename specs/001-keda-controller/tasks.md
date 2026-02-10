# Tasks: KEDA Controller Integration

**Branch**: `001-keda-controller`  
**Input**: Design documents from `/specs/001-keda-controller/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Included following TDD constitutional requirements

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Project Verification)

**Purpose**: Verify existing structure and identify integration points

- [X] T001 Review existing component patterns in cmd/kindenv_start.go (MetricsServer, MySQL, RabbitMQ)
- [X] T002 Review existing configuration structure in internal/kindenv/config.go
- [X] T003 Review existing test patterns in cmd/kindenv_start_test.go
- [X] T004 Verify Helm repository setup pattern in cmd/kindenv_init.go

**Checkpoint**: Understanding of existing patterns confirmed - ready for implementation

---

## Phase 2: User Story 1 - Enable Event-Driven Autoscaling (Priority: P1) 🎯 MVP

**Goal**: Enable developers to install and use KEDA controller in local Kind environment with configuration-based enable/disable, automatic installation, and status monitoring.

**Independent Test**: Enable KEDA in kindenv.yaml, run `devhelper-cli kindenv start`, verify KEDA pods running, check status display, create ScaledObject to verify acceptance.

### Configuration Structure for User Story 1

- [X] T005 [P] [US1] Add Keda struct to Components in internal/kindenv/config.go (after MetricsServer, lines ~115-118)
- [X] T006 [P] [US1] Add KEDA default values to CreateDefaultConfig in internal/kindenv/config.go (line ~960)
- [X] T007 [P] [US1] Update example kindenv.yaml with KEDA configuration section (in components, after metricsServer)

### Helm Repository Setup for User Story 1

- [X] T008 [US1] Add KEDA Helm repository to kindenv_init.go (after metrics-server repo, line ~330)
- [X] T009 [US1] Add KEDA chart availability verification to kindenv_init.go (after repo add)

### Installation Logic for User Story 1

- [X] T010 [US1] Add KEDA installation logic to kindenv_start.go (after MetricsServer install, line ~750)
- [X] T011 [US1] Implement namespace creation for KEDA in kindenv_start.go (within KEDA install block)
- [X] T012 [US1] Implement Helm install command for KEDA in kindenv_start.go (using upgrade --install pattern)
- [X] T013 [US1] Add waitForDeployment call for keda-operator in kindenv_start.go (2-minute timeout)
- [X] T014 [US1] Add user guidance output after successful KEDA installation in kindenv_start.go
- [X] T015 [US1] Implement non-blocking error handling for KEDA installation in kindenv_start.go
- [X] T016 [US1] Add KEDA status to configuration summary in kindenv_start.go (line ~230)

### Status Monitoring for User Story 1

- [X] T017 [US1] Add KEDA status check to kindenv_status.go (after MetricsServer check, line ~260)
- [X] T018 [US1] Implement deployment status parsing for keda-operator in kindenv_status.go
- [X] T019 [US1] Add verbose output for KEDA status in kindenv_status.go

### Cleanup Logic for User Story 1 (FR-012)

- [X] T020 [US1] Add KEDA cleanup check to kindenv_stop.go (after RabbitMQ cleanup, line ~290)
- [X] T021 [US1] Implement KEDA Helm release uninstallation in kindenv_stop.go (helm uninstall keda)
- [X] T022 [US1] Add KEDA namespace deletion in kindenv_stop.go (kubectl delete namespace)
- [X] T023 [US1] Add verbose logging for KEDA cleanup operations in kindenv_stop.go

### Tests for User Story 1 (TDD - Write First)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T024 [P] [US1] Create TestKedaConfiguration test in cmd/kindenv_start_test.go with default config scenario
- [X] T025 [P] [US1] Add TestKedaConfiguration test case for KEDA enabled scenario
- [X] T026 [P] [US1] Add TestKedaConfiguration test case for KEDA disabled scenario
- [X] T027 [P] [US1] Create TestKedaConfigurationValidation test for config struct validation
- [X] T028 [P] [US1] Add TestKedaCleanup test in cmd/kindenv_stop_test.go for cleanup logic

**Checkpoint**: User Story 1 complete - KEDA can be enabled, installed, and status checked. Basic version configuration works.

---

## Phase 3: User Story 2 - Configure KEDA Chart Version (Priority: P2)

**Goal**: Enable developers to specify specific KEDA chart versions for environment parity with production, with proper validation and error handling.

**Independent Test**: Configure specific chart version in kindenv.yaml, start environment, verify installed version matches; test with invalid version to verify error handling.

### Version Validation for User Story 2

- [X] T029 [US2] Add chart version validation to config.Validate() in internal/kindenv/config.go (line ~890)
- [X] T030 [US2] Implement empty version check and error message in config.go validation
- [X] T031 [US2] Add Helm error output parsing in kindenv_start.go for invalid chart version

### Error Handling Enhancement for User Story 2

- [X] T032 [US2] Enhance KEDA installation error messages with version-specific guidance in kindenv_start.go
- [X] T033 [US2] Add chart version in verbose output during installation in kindenv_start.go

### Tests for User Story 2

- [X] T034 [P] [US2] Add TestKedaConfiguration test case for custom chart version in cmd/kindenv_start_test.go
- [X] T035 [P] [US2] Add TestKedaConfiguration test case for invalid chart version in cmd/kindenv_start_test.go
- [X] T036 [P] [US2] Add TestKedaConfiguration test case for empty chart version in cmd/kindenv_start_test.go

**Checkpoint**: User Story 2 complete - Chart version can be configured and validated properly.

---

## Phase 4: User Story 3 - Skip KEDA Installation (Priority: P3)

**Goal**: Enable developers to skip KEDA installation via command-line flag for quick environment startup without editing configuration files.

**Independent Test**: Enable KEDA in config, run `devhelper-cli kindenv start --skip-keda`, verify KEDA not installed; check status shows KEDA as disabled.

### Skip Flag Implementation for User Story 3

- [X] T037 [US3] Add --skip-keda flag definition to kindenv_start.go init() function (line ~2700)
- [X] T038 [US3] Add flag processing in kindenv_start.go Run function (after config load, line ~150)
- [X] T039 [US3] Update status display to show when KEDA is skipped via flag in kindenv_start.go

### Tests for User Story 3

- [X] T040 [P] [US3] Add TestKedaSkipFlag test in cmd/kindenv_start_test.go for skip flag behavior
- [X] T041 [P] [US3] Add TestKedaSkipFlag test case verifying KEDA enabled in config but skipped by flag

**Checkpoint**: User Story 3 complete - Skip flag works correctly, overriding configuration.

---

## Phase 5: Integration & Polish

**Purpose**: Integration testing, documentation, and cross-cutting improvements

### Integration Testing

- [X] T042 [P] Create integration test script for full KEDA installation flow in specs/001-keda-controller/
- [X] T043 [P] Create integration test script for skip flag behavior in specs/001-keda-controller/
- [X] T044 [P] Create integration test for ScaledObject creation after KEDA install
- [X] T045 [P] Create integration test for KEDA cleanup on stop in specs/001-keda-controller/

### Documentation

- [X] T046 [P] Create quickstart.md with KEDA usage examples in specs/001-keda-controller/
- [X] T047 [P] Add ScaledObject example for RabbitMQ autoscaling in quickstart.md
- [X] T048 [P] Create contracts/keda-config.yaml with example configuration
- [X] T049 Update main README.md to include KEDA in components list
- [X] T050 Add KEDA troubleshooting section to quickstart.md

### Code Quality

- [X] T051 Run go fmt on all modified files
- [X] T052 Run go vet on all modified files
- [X] T053 Run golangci-lint on all modified files
- [X] T054 Add godoc comments for KEDA configuration struct in config.go
- [X] T055 Verify error messages include context and solutions

### Final Validation

- [X] T056 Manual test: Fresh Kind cluster with KEDA enabled
- [X] T057 Manual test: KEDA disabled in config
- [X] T058 Manual test: Skip flag with KEDA enabled
- [X] T059 Manual test: Custom chart version installation
- [X] T060 Manual test: Status check with KEDA running
- [X] T061 Manual test: Create ScaledObject resource
- [X] T062 Manual test: Stop environment and verify KEDA cleanup
- [X] T063 Run all unit tests and verify 80%+ coverage
- [X] T064 Verify Constitution compliance checklist

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **User Story 1 (Phase 2)**: Depends on Setup - Core functionality, MUST complete first
- **User Story 2 (Phase 3)**: Depends on User Story 1 (enhances configuration validation)
- **User Story 3 (Phase 4)**: Depends on User Story 1 (adds override to existing installation)
- **Polish (Phase 5)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent - No dependencies on other stories (requires Setup only)
- **User Story 2 (P2)**: Depends on User Story 1 configuration structure
- **User Story 3 (P3)**: Depends on User Story 1 installation logic

### Within Each User Story

**User Story 1**:
- Configuration tasks (T005-T007) can be done in parallel [P]
- Helm repository setup (T008-T009) can be done after configuration
- Installation logic (T010-T016) requires configuration complete
- Status monitoring (T017-T019) requires installation logic exists
- Cleanup logic (T020-T023) requires configuration and installation pattern understanding
- Tests (T024-T028) can be written in parallel [P] before implementation

**User Story 2**:
- Validation tasks (T029-T031) can be done in parallel with error handling [P]
- Error handling (T032-T033) can be done after installation logic exists
- Tests (T034-T036) can be written in parallel [P]

**User Story 3**:
- Flag implementation (T037-T039) sequential but quick
- Tests (T040-T041) can be written in parallel [P]

### Parallel Opportunities

#### User Story 1 Parallel Tasks:
```bash
# Configuration structure (can all run together):
T005: Add Keda struct to config.go
T006: Add default values to config.go
T007: Update kindenv.yaml example

# Tests (can all run together after understanding requirements):
T024: TestKedaConfiguration default
T025: TestKedaConfiguration enabled
T026: TestKedaConfiguration disabled
T027: TestKedaConfigurationValidation
T028: TestKedaCleanup
```

#### User Story 2 Parallel Tasks:
```bash
# Validation and error handling:
T029: Version validation
T030: Empty version check
T031: Helm error parsing

# Tests (can all run together):
T034: Custom version test
T035: Invalid version test
T036: Empty version test
```

#### User Story 3 Parallel Tasks:
```bash
# Tests:
T040: Skip flag test
T041: Skip flag override test
```

#### Polish Phase Parallel Tasks:
```bash
# Documentation (all independent):
T046: Create quickstart.md
T047: Add ScaledObject example
T048: Create config example
T049: Update README
T050: Add troubleshooting

# Integration tests (all independent):
T042: Full installation test
T043: Skip flag test
T044: ScaledObject creation test
T045: Cleanup test
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: User Story 1 (T005-T028)
   - Write tests first (T024-T028)
   - Implement configuration (T005-T007)
   - Implement repository setup (T008-T009)
   - Implement installation (T010-T016)
   - Implement status (T017-T019)
   - Implement cleanup (T020-T023)
3. **STOP and VALIDATE**: Test User Story 1 independently
4. Deploy/demo if ready - developers can now use KEDA in kindenv!

### Incremental Delivery

1. **Iteration 1**: Setup + User Story 1 → Test → Deploy (MVP!)
   - Developers can enable KEDA, use it, and clean up properly
2. **Iteration 2**: Add User Story 2 → Test → Deploy
   - Developers can now specify versions
3. **Iteration 3**: Add User Story 3 → Test → Deploy
   - Developers can now use skip flag
4. **Iteration 4**: Polish → Final validation → Deploy
   - Documentation and integration tests complete

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup together** (T001-T004)
2. **User Story 1** (can split within story):
   - Developer A: Configuration + tests (T005-T007, T024-T028)
   - Developer B: Repository setup (T008-T009)
   - Developer C: Installation logic (T010-T016)
   - Developer D: Status monitoring (T017-T019)
   - Developer E: Cleanup logic (T020-T023)
3. **User Story 2**: One developer (T029-T036) - small enhancement
4. **User Story 3**: One developer (T037-T041) - small enhancement
5. **Polish**: Split documentation and tests across team

---

## Estimated Time

- **Phase 1 (Setup)**: 30 minutes
- **Phase 2 (User Story 1)**: 7 hours
  - Configuration: 1 hour
  - Repository setup: 30 minutes
  - Installation logic: 2.5 hours
  - Status monitoring: 1 hour
  - Cleanup logic: 1 hour
  - Tests: 1 hour
- **Phase 3 (User Story 2)**: 1.5 hours
  - Validation: 45 minutes
  - Error handling: 30 minutes
  - Tests: 15 minutes
- **Phase 4 (User Story 3)**: 1 hour
  - Flag implementation: 30 minutes
  - Tests: 30 minutes
- **Phase 5 (Polish)**: 3 hours
  - Integration tests: 1 hour
  - Documentation: 1.5 hours
  - Validation: 30 minutes

**Total**: ~13 hours (with testing, documentation, and cleanup)

---

## Notes

- All [P] tasks can be done in parallel (different files, no dependencies)
- [Story] labels map tasks to user stories for traceability
- Tests should be written FIRST and FAIL before implementation (TDD)
- Each user story checkpoint represents a deployable increment
- US1 is the MVP - delivers core value including cleanup
- US2 and US3 are enhancements that add convenience
- Follow existing patterns from MetricsServer, MySQL, RabbitMQ (especially cleanup patterns)
- Non-blocking installation ensures environment stability
- Cleanup follows same pattern as MySQL/RabbitMQ in kindenv_stop.go
- Constitution requires 80%+ test coverage - included in tasks