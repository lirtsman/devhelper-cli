# Tasks: Custom Components for KindEnv

**Branch**: `002-custom-components`  
**Feature**: Enable developers to deploy custom services with images, env vars, secrets, ports, resources, and config files  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Organization**: Tasks are grouped by user story (P1-P6) to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5, US6)
- All file paths are relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 [P] Create internal/kindenv/customcomponent.go file structure
- [x] T002 [P] Create internal/kindenv/configmap.go file structure  
- [x] T003 [P] Create internal/kindenv/volume.go file structure
- [x] T004 [P] Create internal/kindenv/validation.go file structure
- [x] T005 [P] Create examples/custom-component-app/ directory structure

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core data structures and configuration parsing - MUST complete before user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Core Data Structures

- [x] T006 [P] Add CustomComponent struct to internal/kindenv/config.go
- [x] T007 [P] Add EnvVar and EnvVarSource structs to internal/kindenv/config.go
- [x] T008 [P] Add PortMapping struct to internal/kindenv/config.go
- [x] T009 [P] Add ResourceRequirements and ResourceList structs to internal/kindenv/config.go
- [x] T010 [P] Add ConfigFile struct to internal/kindenv/config.go

### Configuration Parsing

- [x] T011 Add CustomComponents []CustomComponent field to KindEnvConfig struct in internal/kindenv/config.go
- [x] T012 Update LoadConfig() to parse customComponents section in internal/kindenv/config.go
- [x] T013 Implement SetDefaults() for CustomComponent in internal/kindenv/config.go

### Validation Framework

- [x] T014 [P] Implement CustomComponent.Validate() method in internal/kindenv/validation.go
- [x] T015 [P] Implement EnvVar.Validate() method in internal/kindenv/validation.go
- [x] T016 [P] Implement PortMapping.Validate() method in internal/kindenv/validation.go
- [x] T017 [P] Implement ResourceRequirements.Validate() method in internal/kindenv/validation.go
- [x] T018 [P] Implement ConfigFile.Validate() method in internal/kindenv/validation.go

### Test Foundation

- [x] T019 [P] Create internal/kindenv/config_test.go tests for CustomComponent parsing
- [x] T020 [P] Create internal/kindenv/validation_test.go with table-driven validation tests
- [x] T021 [P] Create internal/kindenv/customcomponent_test.go test file structure

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Deploy Custom Application with Basic Configuration (Priority: P1) 🎯 MVP

**Goal**: Enable developers to deploy custom services with image specification and direct environment variables

**Independent Test**: Add minimal custom component (image + env vars) to kindenv.yaml, run `kindenv start`, verify pod running with `kubectl get pods`

### Tests for User Story 1 (TDD - Write First)

- [ ] T022 [P] [US1] Write test for minimal component deployment in internal/kindenv/customcomponent_test.go
- [ ] T023 [P] [US1] Write test for component with environment variables in internal/kindenv/customcomponent_test.go
- [ ] T024 [P] [US1] Write test for enabled/disabled flag behavior in internal/kindenv/customcomponent_test.go
- [ ] T025 [P] [US1] Write test for namespace specification in internal/kindenv/customcomponent_test.go
- [ ] T026 [P] [US1] Write integration test for basic deployment in cmd/kindenv_start_test.go

**Verify tests FAIL before proceeding to implementation**

### Implementation for User Story 1

- [ ] T027 [US1] Implement deployCustomComponents() function in internal/kindenv/customcomponent.go
- [ ] T028 [US1] Implement generateDeploymentYAML() for basic deployment in internal/kindenv/customcomponent.go
- [ ] T029 [US1] Implement environment variable transformation (direct values) in internal/kindenv/customcomponent.go
- [ ] T030 [US1] Extend kindenv start command to call deployCustomComponents() in cmd/kindenv_start.go
- [ ] T031 [US1] Add custom component status to kindenv status command in cmd/kindenv_status.go
- [ ] T032 [US1] Add custom component cleanup to kindenv stop command in cmd/kindenv_stop.go
- [ ] T033 [US1] Add progress indicators for custom component deployment in cmd/kindenv_start.go
- [ ] T034 [US1] Add error handling and user feedback for deployment failures in cmd/kindenv_start.go

**Checkpoint**: Can deploy minimal custom component with image and direct env vars independently

---

## Phase 4: User Story 2 - Connect to Services via Secrets (Priority: P2)

**Goal**: Enable custom components to use secret references for database credentials

**Independent Test**: Configure component with secretKeyRef env vars, deploy, verify pod receives secret values

### Tests for User Story 2 (TDD - Write First)

- [ ] T035 [P] [US2] Write test for secretKeyRef environment variables in internal/kindenv/customcomponent_test.go
- [ ] T036 [P] [US2] Write test for mixed direct values and secret references in internal/kindenv/customcomponent_test.go
- [ ] T037 [P] [US2] Write test for missing secret detection in internal/kindenv/validation_test.go
- [ ] T038 [P] [US2] Write integration test for secret-based MySQL connection in cmd/kindenv_start_test.go

**Verify tests FAIL before proceeding to implementation**

### Implementation for User Story 2

- [ ] T039 [US2] Implement secretKeyRef transformation in generateDeploymentYAML() in internal/kindenv/customcomponent.go
- [ ] T040 [US2] Implement validateSecretReferences() for pre-deployment check in internal/kindenv/validation.go
- [ ] T041 [US2] Add secret existence validation before deployment in cmd/kindenv_start.go
- [ ] T042 [US2] Add error messaging for missing secrets in internal/kindenv/validation.go

**Checkpoint**: Can deploy components with both direct env vars (US1) and secret references (US2) independently

---

## Phase 5: User Story 3 - Configure Custom Command and Arguments (Priority: P3)

**Goal**: Enable command and args override for custom entrypoints

**Independent Test**: Configure component with command array, deploy, verify via `kubectl describe pod`

### Tests for User Story 3 (TDD - Write First)

- [ ] T043 [P] [US3] Write test for command override in internal/kindenv/customcomponent_test.go
- [ ] T044 [P] [US3] Write test for args without command in internal/kindenv/customcomponent_test.go
- [ ] T045 [P] [US3] Write test for command + args together in internal/kindenv/customcomponent_test.go

**Verify tests FAIL before proceeding to implementation**

### Implementation for User Story 3

- [ ] T046 [US3] Add command field handling in generateDeploymentYAML() in internal/kindenv/customcomponent.go
- [ ] T047 [US3] Add args field handling in generateDeploymentYAML() in internal/kindenv/customcomponent.go
- [ ] T048 [US3] Validate command and args are properly formatted arrays in internal/kindenv/validation.go

**Checkpoint**: Can deploy components with custom commands and args (US3) alongside basic deployment (US1) and secrets (US2)

---

## Phase 6: User Story 4 - Expose Custom Service Ports (Priority: P4)

**Goal**: Enable port mappings from container to host machine via NodePort services

**Independent Test**: Configure component with port mapping, deploy, access service from host via localhost

### Tests for User Story 4 (TDD - Write First)

- [ ] T049 [P] [US4] Write test for single port mapping in internal/kindenv/customcomponent_test.go
- [ ] T050 [P] [US4] Write test for multiple port mappings in internal/kindenv/customcomponent_test.go
- [ ] T051 [P] [US4] Write test for port conflict detection in internal/kindenv/validation_test.go
- [ ] T052 [P] [US4] Write test for NodePort auto-assignment in internal/kindenv/customcomponent_test.go
- [ ] T053 [P] [US4] Write integration test for port accessibility in cmd/kindenv_start_test.go

**Verify tests FAIL before proceeding to implementation**

### Implementation for User Story 4

- [ ] T054 [US4] Implement generateServiceYAML() for NodePort service in internal/kindenv/customcomponent.go
- [ ] T055 [US4] Implement assignPorts() for NodePort auto-assignment in internal/kindenv/customcomponent.go
- [ ] T056 [US4] Implement validatePortConflicts() to check against existing components in internal/kindenv/validation.go
- [ ] T057 [US4] Integrate service creation in deployCustomComponents() in internal/kindenv/customcomponent.go
- [ ] T058 [US4] Update Kind cluster port mappings when custom ports are added in cmd/kindenv_start.go
- [ ] T059 [US4] Add port information to status output in cmd/kindenv_status.go
- [ ] T060 [US4] Clean up services in kindenv stop command in cmd/kindenv_stop.go

**Checkpoint**: Can deploy components with exposed ports (US4) alongside all previous features

---

## Phase 7: User Story 5 - Configure Resource Limits (Priority: P5)

**Goal**: Enable CPU and memory resource requests and limits

**Independent Test**: Configure component with resource limits, deploy, verify via `kubectl describe pod`

### Tests for User Story 5 (TDD - Write First)

- [ ] T061 [P] [US5] Write test for custom resource limits in internal/kindenv/customcomponent_test.go
- [ ] T062 [P] [US5] Write test for default resource application in internal/kindenv/customcomponent_test.go
- [ ] T063 [P] [US5] Write test for resource format validation in internal/kindenv/validation_test.go
- [ ] T064 [P] [US5] Write test for requests <= limits validation in internal/kindenv/validation_test.go

**Verify tests FAIL before proceeding to implementation**

### Implementation for User Story 5

- [ ] T065 [US5] Add resource specification to deployment YAML generation in internal/kindenv/customcomponent.go
- [ ] T066 [US5] Implement defaultResourceRequirements() function in internal/kindenv/customcomponent.go
- [ ] T067 [US5] Implement validateResourceFormat() for CPU/memory format in internal/kindenv/validation.go
- [ ] T068 [US5] Implement validateResourceLimits() to ensure limits >= requests in internal/kindenv/validation.go
- [ ] T069 [US5] Add resource usage to status output (if available) in cmd/kindenv_status.go

**Checkpoint**: Can deploy components with custom resource limits (US5) alongside all previous features

---

## Phase 8: User Story 6 - Mount Custom Configuration Files (Priority: P6)

**Goal**: Enable mounting custom config files via ConfigMaps with inline contents

**Independent Test**: Configure component with configFiles, deploy, verify file exists with correct contents via `kubectl exec`

### Tests for User Story 6 (TDD - Write First)

- [ ] T070 [P] [US6] Write test for ConfigFile validation in internal/kindenv/validation_test.go
- [ ] T071 [P] [US6] Write test for ConfigMap generation in internal/kindenv/configmap_test.go
- [ ] T072 [P] [US6] Write test for multiple config files in internal/kindenv/configmap_test.go
- [ ] T073 [P] [US6] Write test for duplicate path detection in internal/kindenv/validation_test.go
- [ ] T074 [P] [US6] Write test for ConfigMap size limit validation in internal/kindenv/validation_test.go
- [ ] T075 [P] [US6] Write test for volume mount generation in internal/kindenv/volume_test.go
- [ ] T076 [P] [US6] Write integration test for config file mounting in cmd/kindenv_start_test.go

**Verify tests FAIL before proceeding to implementation**

### Implementation for User Story 6

- [ ] T077 [US6] Implement generateConfigMapYAML() in internal/kindenv/configmap.go
- [ ] T078 [US6] Implement createConfigMap() to apply ConfigMap to cluster in internal/kindenv/configmap.go
- [ ] T079 [US6] Implement deleteConfigMap() for cleanup in internal/kindenv/configmap.go
- [ ] T080 [US6] Implement generateVolumes() for ConfigMap volumes in internal/kindenv/volume.go
- [ ] T081 [US6] Implement generateVolumeMounts() with subPath support in internal/kindenv/volume.go
- [ ] T082 [US6] Integrate ConfigMap creation in deployCustomComponents() in internal/kindenv/customcomponent.go
- [ ] T083 [US6] Integrate volume/volumeMounts in generateDeploymentYAML() in internal/kindenv/customcomponent.go
- [ ] T084 [US6] Add ConfigMap cleanup to kindenv stop command in cmd/kindenv_stop.go
- [ ] T085 [US6] Implement mount path override detection and warning logging in internal/kindenv/validation.go
- [ ] T086 [US6] Update status command to show ConfigMap status in cmd/kindenv_status.go

**Checkpoint**: Can deploy components with config files (US6) alongside all previous features (US1-US5)

---

## Phase 9: Integration & Cross-Story Testing

**Purpose**: Verify all user stories work together and independently

- [ ] T087 [P] Write test for multiple custom components with different features in cmd/kindenv_start_test.go
- [ ] T088 [P] Write test for component with all features enabled in cmd/kindenv_start_test.go
- [ ] T089 [P] Write test for parallel deployment of multiple components in cmd/kindenv_start_test.go
- [ ] T090 Write end-to-end test: deploy component with MySQL + OpenSearch + config files in cmd/kindenv_start_test.go
- [ ] T091 [P] Write test for kindenv stop cleanup of all resources in cmd/kindenv_stop_test.go
- [ ] T092 [P] Write test for status reporting of multiple components in cmd/kindenv_status_test.go

---

## Phase 10: Example Application

**Purpose**: Provide working example for developers

- [ ] T093 [P] Create example Spring Boot Dockerfile in examples/custom-component-app/Dockerfile
- [ ] T094 [P] Create example application.yaml config in examples/custom-component-app/config/
- [ ] T095 [P] Create example logback.xml config in examples/custom-component-app/config/
- [ ] T096 Create simple Spring Boot application (or Go app) in examples/custom-component-app/
- [ ] T097 Create kindenv.yaml example config in examples/custom-component-app/
- [ ] T098 Create README.md with setup and run instructions in examples/custom-component-app/

---

## Phase 11: Documentation

**Purpose**: User-facing documentation and examples

- [ ] T099 [P] Update kindenv.yaml with custom component examples
- [ ] T100 [P] Create CUSTOM_COMPONENTS.md user guide in docs/ or root
- [ ] T101 [P] Update main README.md with custom components section
- [ ] T102 [P] Add custom component examples to DEVELOPMENT.md
- [ ] T103 Update CHANGELOG.md with custom components feature

---

## Phase 12: Polish & Validation

**Purpose**: Code quality, performance, and final validation

- [ ] T104 [P] Run go fmt on all modified files
- [ ] T105 [P] Run go vet on all modified files
- [ ] T106 Run golangci-lint and fix any issues
- [ ] T107 [P] Verify test coverage >= 80% with go test -cover
- [ ] T108 Run full test suite and ensure all tests pass
- [ ] T109 Manually test quickstart.md examples end-to-end
- [ ] T110 [P] Review and refactor for code duplication
- [ ] T111 [P] Add godoc comments for all exported types and functions
- [ ] T112 Performance benchmark for configuration validation
- [ ] T113 Verify backward compatibility (existing kindenv.yaml files still work)

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational) ← MUST COMPLETE BEFORE user stories
    ↓
    ├─→ Phase 3 (US1) ← MVP - Can start after Phase 2
    ├─→ Phase 4 (US2) ← Can start after Phase 2 (independent of US1)
    ├─→ Phase 5 (US3) ← Can start after Phase 2 (independent of US1, US2)
    ├─→ Phase 6 (US4) ← Can start after Phase 2 (independent)
    ├─→ Phase 7 (US5) ← Can start after Phase 2 (independent)
    └─→ Phase 8 (US6) ← Can start after Phase 2 (independent)
    ↓
Phase 9 (Integration) ← After all desired user stories
    ↓
Phase 10 (Example App) ← Can run in parallel with Phase 9
    ↓
Phase 11 (Documentation) ← Can run in parallel with Phases 9-10
    ↓
Phase 12 (Polish) ← After all implementation complete
```

### User Story Independence

Each user story (US1-US6) can be implemented and tested independently after Phase 2:

- **US1**: Basic deployment (image, env vars) - No dependencies
- **US2**: Secret references - Builds on US1 env var infrastructure but independently testable
- **US3**: Command/args - Independent feature, no dependencies on US1-US2
- **US4**: Port mappings - Independent feature, no dependencies on others
- **US5**: Resource limits - Independent feature, no dependencies on others
- **US6**: Config file mounting - Independent feature, no dependencies on others

### Parallel Opportunities

**Within Phase 2 (Foundational)**:
```bash
# All struct definitions can be added in parallel:
T006, T007, T008, T009, T010 → All [P]

# All validation methods can be implemented in parallel:
T014, T015, T016, T017, T018 → All [P]

# All test files can be created in parallel:
T019, T020, T021 → All [P]
```

**User Stories (After Phase 2)**:
```bash
# All 6 user stories can be worked on in parallel by different developers:
Phase 3 (US1) || Phase 4 (US2) || Phase 5 (US3) || Phase 6 (US4) || Phase 7 (US5) || Phase 8 (US6)

# Within each user story, all tests can be written in parallel:
US1 Tests: T022, T023, T024, T025, T026 → All [P]
US2 Tests: T035, T036, T037, T038 → All [P]
US3 Tests: T043, T044, T045 → All [P]
US4 Tests: T049, T050, T051, T052, T053 → All [P]
US5 Tests: T061, T062, T063, T064 → All [P]
US6 Tests: T070, T071, T072, T073, T074, T075, T076 → All [P]
```

**Final Phases**:
```bash
# Documentation and examples can run in parallel:
Phase 10 || Phase 11

# Within polish phase:
T104, T105, T107, T110, T111 → All [P]
```

---

## Implementation Strategy

### MVP First (Recommended)

**Minimum Viable Product** = User Story 1 only:

1. Complete Phase 1: Setup (T001-T005)
2. Complete Phase 2: Foundational (T006-T021)
3. Complete Phase 3: User Story 1 (T022-T034)
4. **STOP and VALIDATE**: Test US1 independently
5. Basic custom component deployment is now functional!

**Estimated effort**: ~30-40% of total feature (T001-T034)

### Incremental Delivery

Add user stories incrementally based on priority:

1. **MVP** (US1): Basic deployment (image, env vars) → ~40% effort
2. **+US2**: Secret references → +10% effort (cumulative 50%)
3. **+US3**: Command/args → +5% effort (cumulative 55%)
4. **+US4**: Port mappings → +15% effort (cumulative 70%)
5. **+US5**: Resource limits → +10% effort (cumulative 80%)
6. **+US6**: Config file mounting → +20% effort (cumulative 100%)

Each increment is independently testable and delivers user value!

### Parallel Team Strategy

With 3+ developers after Phase 2 completes:

- **Developer A**: User Story 1 (P1) - Core deployment
- **Developer B**: User Story 4 (P4) - Port mappings
- **Developer C**: User Story 6 (P6) - Config file mounting
- **Developer D**: User Story 2+3+5 (P2, P3, P5) - Smaller stories

Stories can merge independently as they complete.

---

## Task Count Summary

- **Phase 1 (Setup)**: 5 tasks
- **Phase 2 (Foundational)**: 16 tasks ← Blocking prerequisite
- **Phase 3 (US1)**: 13 tasks ← MVP
- **Phase 4 (US2)**: 8 tasks
- **Phase 5 (US3)**: 6 tasks
- **Phase 6 (US4)**: 12 tasks
- **Phase 7 (US5)**: 9 tasks
- **Phase 8 (US6)**: 17 tasks
- **Phase 9 (Integration)**: 6 tasks
- **Phase 10 (Example)**: 6 tasks
- **Phase 11 (Documentation)**: 5 tasks
- **Phase 12 (Polish)**: 10 tasks

**Total**: 113 tasks

**MVP Scope**: 34 tasks (Phases 1-3: T001-T034)  
**Full Feature**: 113 tasks

---

## File Change Summary

### New Files (10-12 files)
- internal/kindenv/customcomponent.go
- internal/kindenv/customcomponent_test.go
- internal/kindenv/configmap.go
- internal/kindenv/configmap_test.go
- internal/kindenv/volume.go
- internal/kindenv/volume_test.go
- internal/kindenv/validation.go
- internal/kindenv/validation_test.go
- examples/custom-component-app/* (6+ files)
- docs/CUSTOM_COMPONENTS.md

### Modified Files (6-8 files)
- internal/kindenv/config.go (CustomComponent structs)
- internal/kindenv/config_test.go (parsing tests)
- cmd/kindenv_start.go (deploy custom components)
- cmd/kindenv_start_test.go (deployment tests)
- cmd/kindenv_stop.go (cleanup)
- cmd/kindenv_stop_test.go (cleanup tests)
- cmd/kindenv_status.go (status reporting)
- cmd/kindenv_status_test.go (status tests)
- kindenv.yaml (examples)
- README.md (documentation)

---

## Next Steps

**Immediate**: Start with MVP implementation

```bash
# Recommended workflow:
1. Complete Phase 1 (Setup) - Create file structures
2. Complete Phase 2 (Foundational) - Core data structures and validation
3. Complete Phase 3 (US1) - Basic deployment capability
4. Test MVP independently
5. Then add US2, US3, US4, US5, US6 incrementally
```

**Alternative**: Run `/speckit.analyze` to validate consistency before implementation

**Ready for**: Implementation via `/speckit.implement` or manual task execution

**Independent Test Criteria per Story**:
- US1: Deploy with image + env vars, verify pod running
- US2: Deploy with secretKeyRef, verify secret values injected
- US3: Deploy with command override, verify via kubectl describe
- US4: Deploy with ports, access from localhost
- US5: Deploy with resources, verify limits via kubectl describe
- US6: Deploy with configFiles, verify file contents via kubectl exec
