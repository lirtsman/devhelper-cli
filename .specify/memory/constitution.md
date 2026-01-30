# DevHelper CLI Constitution

## Core Principles

### I. Code Quality Standards (NON-NEGOTIABLE)
All code must meet strict quality standards:
- Go code follows official Go style guide and passes `go fmt`, `go vet`, and `golangci-lint`
- Clear, self-documenting code with meaningful variable and function names
- Comprehensive godoc comments for all exported functions, types, and packages
- Error handling must be explicit with wrapped errors using `fmt.Errorf` or `errors.Wrap`
- No code duplication - extract common functionality into reusable packages

### II. Test-Driven Development (NON-NEGOTIABLE)
Testing is mandatory and follows strict TDD practices:
- Tests written before implementation (Red-Green-Refactor cycle)
- Minimum 80% code coverage for all new features
- Unit tests for all business logic using Go's testing package
- Table-driven tests for multiple input scenarios
- Integration tests for external dependencies and system interactions
- Benchmark tests for performance-critical code paths

### III. Cobra CLI Best Practices (NON-NEGOTIABLE)
All CLI commands must follow Cobra framework standards:
- Consistent command structure: `devhelper-cli <command> <subcommand> [flags]`
- Clear, concise command descriptions and usage examples
- Proper flag definitions with short and long forms where appropriate
- Input validation with meaningful error messages before command execution
- Consistent output formatting with support for `--output` flag (json, yaml, table)
- Progress indicators for long-running operations using appropriate libraries

### IV. Command Design Standards
CLI commands must provide excellent user experience:
- Commands are idempotent where possible (safe to run multiple times)
- Dry-run mode (`--dry-run`) for destructive operations
- Verbose mode (`--verbose`) for detailed operation logging
- Confirmation prompts for destructive actions (with `--force` override)
- Consistent exit codes: 0 for success, 1 for user errors, 2 for system errors
- Help text includes practical examples and common use cases

### V. Error Handling & User Feedback
All user interactions must be clear and actionable:
- Error messages include context and suggested solutions
- Structured logging with appropriate levels (debug, info, warn, error)
- Color-coded output for better readability (using libraries like fatih/color)
- Progress bars and spinners for operations taking more than 2 seconds
- Clear success/failure indicators with next steps guidance

## Cobra CLI Implementation Standards

### Command Structure Requirements
- Root command (`devhelper-cli`) provides global flags and version info
- Subcommands organized by functional area (e.g., `kindenv`, `localenv`, `tw`)
- Command files follow naming pattern: `<command>_<subcommand>.go`
- Each command has dedicated test file: `<command>_<subcommand>_test.go`
- Shared functionality extracted to helper packages in `internal/`

### Flag and Configuration Management
- Global flags defined in root command and inherited by subcommands
- Configuration files supported in multiple formats (YAML, JSON, TOML)
- Environment variable support with `DEVHELPER_` prefix
- Flag precedence: CLI flags > environment variables > config file > defaults
- Sensitive data (passwords, tokens) never logged or displayed in help

### Command Testing Standards
- Each command has unit tests covering success and error scenarios
- Integration tests for commands that interact with external systems
- Mock external dependencies using interfaces and dependency injection
- Test CLI output formatting and error message content
- Performance tests for commands with specific timing requirements

## Development Workflow

### Code Review Requirements
- All code changes require peer review before merge
- Reviewers must verify Cobra CLI patterns and constitutional compliance
- CLI usability review required for new commands or flag changes
- Documentation updates required for user-facing changes
- Breaking changes require migration guide and deprecation notices

### Quality Gates
- All tests must pass (unit, integration, and CLI tests)
- Code coverage must not decrease below current baseline
- `golangci-lint` must pass with zero issues
- CLI commands must be manually tested in clean environment
- Performance benchmarks must not regress beyond 10% of baseline

## Governance

### Constitutional Authority
This constitution supersedes all other development practices:
- All pull requests must demonstrate compliance with Cobra CLI standards
- Code complexity must be justified with clear user benefit
- CLI design changes require user experience consideration
- Regular constitutional review based on CLI usage patterns and feedback

### Cobra CLI Compliance Checklist
Every new command must verify:
- [ ] Follows standard command naming and structure patterns
- [ ] Includes comprehensive help text with examples
- [ ] Implements proper flag validation and error handling
- [ ] Supports standard output formats (--output flag)
- [ ] Has corresponding unit and integration tests
- [ ] Follows consistent error handling and user feedback patterns

### Amendment Process
- Constitutional changes require team consensus and CLI impact assessment
- Breaking changes to CLI patterns require deprecation period and migration guide
- User experience changes must be validated with actual usage scenarios
- All amendments include updated examples and documentation

**Version**: 2.0.0 | **Ratified**: 2026-01-30 | **Last Amended**: 2026-01-30
