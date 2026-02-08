# Tasks: RabbitMQ Support for KindEnv

**Input**: Design documents from `/specs/003-rabbitmq-kindenv/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Following TDD principles from the constitution - tests written before implementation

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Following existing devhelper-cli structure:
- **Configuration**: `internal/kindenv/` for core logic
- **Commands**: `cmd/` for CLI commands
- **Examples**: `examples/kindenv.yaml` for configuration examples
- **Docs**: `docs/KINDENV.md` for user documentation

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and RabbitMQ-specific structure setup

- [x] T001 Review existing MySQL implementation pattern in internal/kindenv/mysql.go for reference
- [x] T002 [P] Add Bitnami Helm repository configuration in cmd/kindenv_init.go
- [x] T003 [P] Create contracts directory structure for RabbitMQ interfaces

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core RabbitMQ configuration and validation infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Configuration Schema (TDD - Tests First)

- [x] T004 [P] Write table-driven tests for RabbitMQ configuration validation in internal/kindenv/config_test.go
- [x] T005 [P] Write tests for virtual host validation in internal/kindenv/rabbitmq_test.go
- [x] T006 [P] Write tests for NodePort validation (AMQP + Management) in internal/kindenv/rabbitmq_test.go
- [x] T007 [P] Write tests for resource format validation in internal/kindenv/rabbitmq_test.go

### Configuration Implementation (TDD - Make Tests Pass)

- [x] T008 Add RabbitMQ struct to Components section in internal/kindenv/config.go
- [x] T009 Add RabbitMQ secret struct to Secrets section in internal/kindenv/config.go
- [x] T010 [P] Add RabbitMQ defaults to LoadConfig() function in internal/kindenv/config.go
- [x] T011 [P] Add RabbitMQ to generateDefaultPortMappings() in internal/kindenv/config.go
- [x] T012 [P] Add RabbitMQ to processVariableSubstitutions() switch case in internal/kindenv/config.go

### RabbitMQ Manager Interface (TDD - Tests First)

- [x] T013 Create internal/kindenv/rabbitmq.go with interface definitions from contracts/rabbitmq-api-interface.go
- [x] T014 [P] Implement ValidateVirtualHost() function in internal/kindenv/rabbitmq.go
- [x] T015 [P] Implement ValidateNodePorts() function in internal/kindenv/rabbitmq.go
- [x] T016 [P] Implement ValidateResources() function (copy from mysql.go pattern) in internal/kindenv/rabbitmq.go
- [x] T017 [P] Implement ValidateChartVersion() function in internal/kindenv/rabbitmq.go
- [x] T018 Implement ValidateRabbitMQConfig() orchestrator function in internal/kindenv/rabbitmq.go

### Verify Foundation

- [x] T019 Run all configuration validation tests: go test ./internal/kindenv -run TestRabbitMQ -v
- [ ] T020 Verify code coverage meets 80% minimum: go test ./internal/kindenv -cover

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic RabbitMQ Installation (Priority: P1) 🎯 MVP

**Goal**: Enable developers to add RabbitMQ to Kind environment via configuration and have it automatically installed with AMQP (5672) and Management UI (15672) accessible

**Independent Test**: Enable RabbitMQ in kindenv.yaml, run `devhelper-cli kindenv start`, verify pod running and both ports accessible

### Tests for User Story 1 (TDD - Write Tests First)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T021 [P] [US1] Write integration test for RabbitMQ pod creation in cmd/kindenv_start_test.go
- [ ] T022 [P] [US1] Write integration test for AMQP port accessibility in cmd/kindenv_start_test.go
- [ ] T023 [P] [US1] Write integration test for Management UI accessibility in cmd/kindenv_start_test.go

### Implementation for User Story 1

#### Kubernetes Secret Creation

- [x] T024 [US1] Implement generateErlangCookie() helper function in cmd/kindenv_start.go
- [x] T025 [US1] Implement createRabbitMQSecret() function in cmd/kindenv_start.go

#### Helm Chart Installation

- [x] T026 [US1] Implement buildRabbitMQHelmValues() function in cmd/kindenv_start.go
- [x] T027 [US1] Implement installRabbitMQHelmChart() function in cmd/kindenv_start.go

#### Main Installation Flow

- [x] T028 [US1] Add RabbitMQ installation logic to kindenv start command in cmd/kindenv_start.go
- [x] T029 [US1] Add RabbitMQ health check and readiness wait logic in cmd/kindenv_start.go

#### Cleanup Logic

- [x] T030 [US1] Add RabbitMQ uninstall logic to kindenv stop command in cmd/kindenv_stop.go

### Verification for User Story 1

- [ ] T031 [US1] Run integration tests to verify basic installation works
- [ ] T032 [US1] Manual test: Enable RabbitMQ, run kindenv start, verify AMQP accessible at localhost:5672
- [ ] T033 [US1] Manual test: Verify Management UI accessible at http://localhost:15672

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - RabbitMQ Configuration Management (Priority: P2)

**Goal**: Enable developers to customize RabbitMQ settings (credentials, virtual host, ports, resources) through kindenv.yaml configuration

**Independent Test**: Modify RabbitMQ configuration in kindenv.yaml with custom values, start environment, verify settings applied correctly

### Tests for User Story 2 (TDD - Write Tests First)

- [ ] T034 [P] [US2] Write test for custom credentials configuration in internal/kindenv/config_test.go
- [ ] T035 [P] [US2] Write test for custom virtual host configuration in internal/kindenv/rabbitmq_test.go
- [ ] T036 [P] [US2] Write test for custom port mappings in internal/kindenv/config_test.go
- [ ] T037 [P] [US2] Write test for custom resource limits in internal/kindenv/rabbitmq_test.go

### Implementation for User Story 2

#### Configuration Enhancement

- [x] T038 [P] [US2] Update buildRabbitMQHelmValues() to use custom virtual host in cmd/kindenv_start.go
- [x] T039 [P] [US2] Update buildRabbitMQHelmValues() to use custom credentials in cmd/kindenv_start.go
- [x] T040 [P] [US2] Update buildRabbitMQHelmValues() to use custom resource limits in cmd/kindenv_start.go
- [ ] T041 [US2] Add configuration validation for custom settings in cmd/kindenv_start.go

#### Testing Custom Configuration

- [ ] T042 [US2] Integration test for custom credentials authentication
- [ ] T043 [US2] Integration test for custom virtual host creation
- [ ] T044 [US2] Integration test for resource limit enforcement

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - RabbitMQ Data Persistence (Priority: P3)

**Goal**: Enable optional data persistence for RabbitMQ queues, exchanges, and messages across kindenv restarts

**Independent Test**: Create queues/messages, enable persistence in config, restart kindenv, verify data preserved

### Tests for User Story 3 (TDD - Write Tests First)

- [ ] T045 [P] [US3] Write test for persistence enabled configuration in internal/kindenv/config_test.go
- [ ] T046 [P] [US3] Write test for persistence size validation in internal/kindenv/rabbitmq_test.go
- [ ] T047 [US3] Write integration test for data persistence across restarts in cmd/kindenv_start_test.go

### Implementation for User Story 3

#### Persistence Configuration

- [ ] T048 [P] [US3] Add persistence configuration to buildRabbitMQHelmValues() in cmd/kindenv_start.go
- [ ] T049 [US3] Add PersistentVolume validation logic in cmd/kindenv_start.go

#### Testing Persistence

- [ ] T050 [US3] Integration test: Create durable queue, restart, verify queue exists
- [ ] T051 [US3] Integration test: Publish persistent message, restart, verify message preserved
- [ ] T052 [US3] Integration test: Disable persistence, restart, verify clean state

**Checkpoint**: All user stories 1-3 should now be independently functional

---

## Phase 6: User Story 4 - RabbitMQ Health Monitoring (Priority: P3)

**Goal**: Provide RabbitMQ status and health information through kindenv status command

**Independent Test**: Run `devhelper-cli kindenv status`, verify accurate RabbitMQ status displayed with connection info

### Tests for User Story 4 (TDD - Write Tests First)

- [ ] T053 [P] [US4] Write test for getRabbitMQStatus() function in cmd/kindenv_status_test.go
- [ ] T054 [P] [US4] Write test for AMQP connectivity check in cmd/kindenv_status_test.go
- [ ] T055 [P] [US4] Write test for Management API health check in cmd/kindenv_status_test.go

### Implementation for User Story 4

#### Status Reporting Functions

- [x] T056 [P] [US4] Implement checkPodReady() helper function in cmd/kindenv_status.go
- [x] T057 [P] [US4] Implement checkServiceReady() helper function in cmd/kindenv_status.go
- [x] T058 [P] [US4] Implement testAMQPConnection() function in cmd/kindenv_status.go
- [x] T059 [P] [US4] Implement testManagementAPI() function in cmd/kindenv_status.go
- [x] T060 [US4] Implement getRabbitMQStatus() orchestrator function in cmd/kindenv_status.go

#### Status Display

- [x] T061 [US4] Add RabbitMQ status display to kindenv status command in cmd/kindenv_status.go
- [x] T062 [US4] Format connection information display (AMQP URL + Management URL) in cmd/kindenv_status.go

#### Testing Status Reporting

- [ ] T063 [US4] Integration test: Verify status shows "Running" when RabbitMQ is healthy
- [ ] T064 [US4] Integration test: Verify status shows "Failed" when RabbitMQ is down
- [ ] T065 [US4] Integration test: Verify connection info accuracy

**Checkpoint**: All user stories should now be independently functional with comprehensive health monitoring

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and final refinements across all user stories

### Documentation

- [ ] T066 [P] Update docs/KINDENV.md with RabbitMQ configuration section
- [ ] T067 [P] Update docs/KINDENV.md with RabbitMQ connection examples
- [ ] T068 [P] Update docs/KINDENV.md with RabbitMQ troubleshooting guide
- [ ] T069 [P] Add RabbitMQ configuration example to examples/kindenv.yaml

### Code Quality

- [ ] T070 [P] Run golangci-lint on all modified files
- [ ] T071 [P] Run go fmt on all Go files
- [ ] T072 [P] Verify godoc comments on all exported functions
- [ ] T073 Add inline code comments for complex RabbitMQ logic

### Final Integration Testing

- [ ] T074 Run all unit tests: go test ./internal/kindenv -v
- [ ] T075 Run all integration tests: go test ./cmd -v -run TestRabbitMQ
- [ ] T076 Verify code coverage: go test ./... -cover
- [ ] T077 Manual end-to-end test following quickstart.md guide

### Edge Case Handling

- [ ] T078 [P] Add error handling for RabbitMQ image pull failures
- [ ] T079 [P] Add error handling for port conflicts
- [ ] T080 [P] Add error handling for insufficient cluster resources
- [ ] T081 Add error handling for PersistentVolume creation failures

### Performance & Security

- [ ] T082 [P] Verify RabbitMQ starts within 2-minute performance target
- [ ] T083 [P] Verify resource limits are enforced correctly
- [ ] T084 Verify erlang cookie is generated securely with crypto/rand
- [ ] T085 Verify credentials are not logged in verbose output

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - No dependencies on other stories
  - User Story 2 (P2): Can start after Foundational - Builds on US1 but independently testable
  - User Story 3 (P3): Can start after Foundational - Extends US1 with persistence
  - User Story 4 (P3): Can start after Foundational - Adds monitoring to US1
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Extends US1 configuration, independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Adds persistence to US1, independently testable
- **User Story 4 (P3)**: Can start after Foundational (Phase 2) - Adds monitoring to US1, independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD cycle)
- Helper functions before main implementation
- Core functionality before edge case handling
- Integration tests after implementation complete
- Manual verification after automated tests pass

### Parallel Opportunities

#### Phase 2 (Foundational)
- All test writing (T004-T007) can run in parallel
- All validation functions (T014-T017) can run in parallel after tests
- Port mapping and substitution (T011-T012) can run in parallel

#### Phase 3 (User Story 1)
- All tests (T021-T023) can run in parallel
- Secret generation and Helm functions (T024-T027) can run in parallel

#### Phase 4 (User Story 2)
- All tests (T034-T037) can run in parallel
- Helm value updates (T038-T040) can run in parallel

#### Phase 5 (User Story 3)
- All tests (T045-T046) can run in parallel

#### Phase 6 (User Story 4)
- All tests (T053-T055) can run in parallel
- Helper functions (T056-T059) can run in parallel

#### Phase 7 (Polish)
- All documentation tasks (T066-T069) can run in parallel
- All code quality tasks (T070-T073) can run in parallel
- All edge case handling (T078-T081) can run in parallel
- All performance/security checks (T082-T085) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task T021: "Write integration test for RabbitMQ pod creation in cmd/kindenv_start_test.go"
Task T022: "Write integration test for AMQP port accessibility in cmd/kindenv_start_test.go"
Task T023: "Write integration test for Management UI accessibility in cmd/kindenv_start_test.go"

# Launch helper functions in parallel:
Task T024: "Implement generateErlangCookie() helper function in cmd/kindenv_start.go"
Task T025: "Implement createRabbitMQSecret() function in cmd/kindenv_start.go"
Task T026: "Implement buildRabbitMQHelmValues() function in cmd/kindenv_start.go"
Task T027: "Implement installRabbitMQHelmChart() function in cmd/kindenv_start.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T020) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 (T021-T033)
4. **STOP and VALIDATE**: Test User Story 1 independently
   - Manual test: Enable RabbitMQ, start kindenv, connect via AMQP
   - Manual test: Access Management UI at http://localhost:15672
   - Verify all acceptance criteria from spec.md
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational (T001-T020) → Foundation ready
2. Add User Story 1 (T021-T033) → Test independently → Deploy/Demo (MVP!)
   - **Delivers**: Basic RabbitMQ installation with default settings
3. Add User Story 2 (T034-T044) → Test independently → Deploy/Demo
   - **Delivers**: Customizable configuration (credentials, virtual hosts, resources)
4. Add User Story 3 (T045-T052) → Test independently → Deploy/Demo
   - **Delivers**: Optional data persistence across restarts
5. Add User Story 4 (T053-T065) → Test independently → Deploy/Demo
   - **Delivers**: Health monitoring and status reporting
6. Polish (T066-T085) → Final release quality

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T020)
2. Once Foundational is done, parallel development:
   - Developer A: User Story 1 (T021-T033) - Basic installation
   - Developer B: User Story 2 (T034-T044) - Configuration management
   - Developer C: User Story 3 (T045-T052) - Persistence
   - Developer D: User Story 4 (T053-T065) - Health monitoring
3. Stories complete and integrate independently
4. Team collaborates on Polish phase (T066-T085)

---

## Testing Checkpoints

### After Phase 2 (Foundational)
```bash
# Verify all validation works
go test ./internal/kindenv -run TestRabbitMQ -v
go test ./internal/kindenv -cover

# Expected: All config validation tests pass, 80%+ coverage
```

### After Phase 3 (User Story 1 - MVP)
```bash
# Enable RabbitMQ in kindenv.yaml
components:
  rabbitmq:
    enabled: true

# Start environment
devhelper-cli kindenv start

# Verify pod running
kubectl get pods -n rabbitmq

# Test AMQP connection
telnet localhost 5672

# Test Management UI
curl http://localhost:15672/api/overview -u user:password
open http://localhost:15672

# Expected: Pod running, both ports accessible, UI loads
```

### After Phase 4 (User Story 2)
```bash
# Update kindenv.yaml with custom settings
components:
  rabbitmq:
    enabled: true
    virtualHost: "/dev"
secrets:
  rabbitmq:
    username: "admin"
    password: "secure-password"

# Restart environment
devhelper-cli kindenv stop && devhelper-cli kindenv start

# Verify custom virtual host
curl http://localhost:15672/api/vhosts -u admin:secure-password

# Expected: Custom virtual host created, custom credentials work
```

### After Phase 5 (User Story 3)
```bash
# Enable persistence
components:
  rabbitmq:
    enabled: true
    persistence:
      enabled: true

# Create test queue, restart, verify persistence
devhelper-cli kindenv start
# (Create queue via Management UI or AMQP)
devhelper-cli kindenv stop
devhelper-cli kindenv start
# (Verify queue still exists)

# Expected: Queue and messages persist across restarts
```

### After Phase 6 (User Story 4)
```bash
# Check status
devhelper-cli kindenv status

# Expected output includes:
# - RabbitMQ status: Running
# - AMQP URL: amqp://user:***@localhost:5672/
# - Management UI: http://localhost:15672
```

---

## Success Metrics

### Code Quality
- ✅ 80%+ test coverage on all RabbitMQ code
- ✅ All golangci-lint checks pass
- ✅ Godoc comments on all exported functions
- ✅ TDD workflow followed (tests written first)

### Functionality
- ✅ All 4 user stories independently functional
- ✅ All acceptance criteria from spec.md met
- ✅ All edge cases handled gracefully
- ✅ Performance targets met (< 2 min startup)

### Documentation
- ✅ KINDENV.md updated with RabbitMQ section
- ✅ Configuration examples provided
- ✅ Troubleshooting guide complete
- ✅ Connection examples for major languages

---

## Notes

- [P] tasks = different files, no dependencies, can run in parallel
- [US1], [US2], [US3], [US4] labels map tasks to specific user stories for traceability
- Each user story should be independently completable and testable
- Follow TDD: Write tests first, watch them fail, implement, watch them pass
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Reference MySQL implementation (specs/001-mysql8-kindenv/) as pattern guide
- Use quickstart.md as detailed implementation reference
- Total tasks: 85 (excluding manual verification steps)
- Parallel opportunities: 40+ tasks can run concurrently within their phases
