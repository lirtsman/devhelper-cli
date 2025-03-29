# Testing Guidelines for devhelper-cli

This document outlines best practices for organizing and writing tests for the devhelper-cli project.

## Test File Organization

1. **Test files should be located alongside the files they test**
   - Example: `cmd/localenv.go` → `cmd/localenv_test.go`

2. **Consolidate related tests**
   - Instead of creating multiple small test files, group tests by functionality
   - For example, merge `dashboard_test.go` into `localenv_helpers_test.go` if they test related functionality

3. **Naming conventions for test files**
   - Use `{filename}_test.go` naming format
   - For component tests that span multiple files, use the main component name (e.g., `localenv_test.go` for testing the localenv command)

## Test Structure

1. **Use descriptive test function names**
   - Format: `Test{ComponentName}{Functionality}`
   - Example: `TestLocalenvStartCommand`, `TestStatusHelperFunctions`

2. **Use table-driven tests wherever possible**
   ```go
   func TestSomething(t *testing.T) {
       testCases := []struct {
           name     string
           input    string
           expected bool
       }{
           {"valid case", "valid-input", true},
           {"invalid case", "invalid-input", false},
       }
       
       for _, tc := range testCases {
           t.Run(tc.name, func(t *testing.T) {
               result := functionToTest(tc.input)
               assert.Equal(t, tc.expected, result)
           })
       }
   }
   ```

3. **Use subtests for different test scenarios**
   - Creates clearer reports
   - Allows running specific subtests with `go test -run TestName/SubtestName`

## Mocking and Test Dependencies

1. **Use dependency injection for easy mocking**
   - Replace direct dependencies with interfaces that can be mocked
   - Inject dependencies into functions rather than creating them inside functions

2. **Keep mocks in a separate package**
   - Store mocks in `internal/test` or `mocks` directory
   - Reuse mocks across tests

3. **Avoid testing implementation details**
   - Test behavior, not internal workings
   - Focus on inputs and outputs, not how the function achieves its result

## Coverage

1. **Aim for at least 70% code coverage**
   - Focus on critical paths and error handling
   - Not all code needs to be covered, prioritize important functionality

2. **Check coverage with Make commands**
   - `make test-coverage` - Generate coverage data
   - `make test-coverage-html` - Generate HTML coverage report
   - `make test-coverage-func` - Show function-level coverage stats

## Refactoring Strategy

To improve the current test organization:

1. **Consolidate related test files**
   - Merge `dashboard_test.go` into `localenv_helpers_test.go`
   - Organize other test files by command functionality

2. **Standardize test naming and structure**
   - Ensure all tests follow the same patterns
   - Use consistent approaches for mocking and assertions

3. **Improve coverage of untested functions**
   - Prioritize functions with 0% coverage
   - Add tests for error handling and edge cases

4. **Make complex functions more testable**
   - Refactor functions that interact with the filesystem or execute commands
   - Add interfaces and dependency injection to make mocking easier 