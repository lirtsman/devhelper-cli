package cmd

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Create a variable for exec.Command that we can override in tests
var execCommand = exec.Command

// mockCommand creates a mock command execution function
type mockCommandFunc func(command string, args ...string) *exec.Cmd

// mockableExec creates a mockable version of the exec.Command function
func mockableExec(output string, err error) mockCommandFunc {
	return func(command string, args ...string) *exec.Cmd {
		// Create a fake command that does nothing but return the provided output/error
		return &exec.Cmd{}
	}
}

// TestLocalenvHelpers tests the helper functions in localenv_helpers.go
func TestLocalenvHelpers(t *testing.T) {
	// Table-driven test for requirement functions
	requirementTests := []struct {
		name          string
		function      func(bool, bool, bool) bool
		configLoaded  bool
		configValue   bool
		skipFlag      bool
		expectedValue bool
		description   string
	}{
		// getDaprRequirement tests
		{"getDaprRequirement with config loaded and true", getDaprRequirement, true, true, false, true, "When config loaded and flag is true, should return true"},
		{"getDaprRequirement with config loaded and false", getDaprRequirement, true, false, false, false, "When config loaded and flag is false, should return false"},
		{"getDaprRequirement with config loaded and skip flag", getDaprRequirement, true, false, true, false, "When config loaded, should ignore skip flag"},
		{"getDaprRequirement with config loaded, true and skip flag", getDaprRequirement, true, true, true, true, "When config loaded, should ignore skip flag"},
		{"getDaprRequirement without config, no skip", getDaprRequirement, false, false, false, true, "Without config, should return !skipFlag"},
		{"getDaprRequirement without config, with skip", getDaprRequirement, false, false, true, false, "Without config, should return !skipFlag"},

		// getTemporalRequirement tests
		{"getTemporalRequirement with config loaded and true", getTemporalRequirement, true, true, false, true, "When config loaded and flag is true, should return true"},
		{"getTemporalRequirement with config loaded and false", getTemporalRequirement, true, false, false, false, "When config loaded and flag is false, should return false"},
		{"getTemporalRequirement with config loaded and skip flag", getTemporalRequirement, true, false, true, false, "When config loaded, should ignore skip flag"},
		{"getTemporalRequirement with config loaded, true and skip flag", getTemporalRequirement, true, true, true, true, "When config loaded, should ignore skip flag"},
		{"getTemporalRequirement without config, no skip", getTemporalRequirement, false, false, false, true, "Without config, should return !skipFlag"},
		{"getTemporalRequirement without config, with skip", getTemporalRequirement, false, false, true, false, "Without config, should return !skipFlag"},

		// getDaprDashboardRequirement tests
		{"getDaprDashboardRequirement with config loaded and true", getDaprDashboardRequirement, true, true, false, true, "When config loaded and flag is true, should return true"},
		{"getDaprDashboardRequirement with config loaded and false", getDaprDashboardRequirement, true, false, false, false, "When config loaded and flag is false, should return false"},
		{"getDaprDashboardRequirement with config loaded and skip flag", getDaprDashboardRequirement, true, false, true, false, "When config loaded, should ignore skip flag"},
		{"getDaprDashboardRequirement with config loaded, true and skip flag", getDaprDashboardRequirement, true, true, true, true, "When config loaded, should ignore skip flag"},
		{"getDaprDashboardRequirement without config, no skip", getDaprDashboardRequirement, false, false, false, true, "Without config, should return !skipFlag"},
		{"getDaprDashboardRequirement without config, with skip", getDaprDashboardRequirement, false, false, true, false, "Without config, should return !skipFlag"},
	}

	for _, tc := range requirementTests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.function(tc.configLoaded, tc.configValue, tc.skipFlag)
			assert.Equal(t, tc.expectedValue, result, tc.description)
		})
	}

	t.Run("isTemporalNamespaceExist should handle edge cases", func(t *testing.T) {
		// This function just calls exec.Command, so we can't easily mock it
		// without more advanced test infrastructure. Just call with an invalid
		// name to verify it handles errors correctly
		assert.False(t, isTemporalNamespaceExist("this-namespace-does-not-exist-for-test"),
			"Should return false for non-existent namespace")
	})
}

// TestIsPortInUse tests the port checking functionality
func TestIsPortInUse(t *testing.T) {
	portTests := []struct {
		name        string
		port        int
		expectation string
	}{
		{"negative port", -1, "Negative port should return false"},
		{"port greater than max", 65536, "Port > 65535 should return false"},
		{"unusual port", 54321, "Testing if unusual port is in use"},
	}

	for _, tc := range portTests {
		t.Run(tc.name, func(t *testing.T) {
			result := isPortInUse(tc.port)
			if tc.port == 54321 {
				// We can't assert a specific result for the unusual port
				t.Logf("Port %d is in use: %v", tc.port, result)
			} else {
				assert.False(t, result, tc.expectation)
			}
		})
	}
}

// TestDashboardFunctions tests all dashboard-related functionality in one place
func TestDashboardFunctions(t *testing.T) {
	// Define test cases for both functions
	dashboardTests := []struct {
		name           string
		testFunc       func() bool
		expectedResult bool
		description    string
		skipTest       bool
		skipReason     string
	}{
		{
			name: "tryStartDashboard with non-existent command",
			testFunc: func() bool {
				return tryStartDashboard("non-existent-command", 8080, nil)
			},
			expectedResult: false,
			description:    "Should return false for non-existent command",
			skipTest:       true,
			skipReason:     "Skipping this test as it takes 3 seconds to run. The functionality is covered by the timeout variant test.",
		},
		{
			name: "tryStartDashboardWithTimeout with non-existent command",
			testFunc: func() bool {
				return tryStartDashboardWithTimeout("non-existent-command", 8080, nil, 100*time.Millisecond)
			},
			expectedResult: false,
			description:    "Should return false for non-existent command with custom timeout",
			skipTest:       false,
		},
	}

	for _, tc := range dashboardTests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipTest {
				t.Skip(tc.skipReason)
			}
			result := tc.testFunc()
			assert.Equal(t, tc.expectedResult, result, tc.description)
		})
	}
}

// TestHelperProcess isn't a real test. It's used to mock exec.Command behavior in other tests.
// This pattern is inspired by the approach used in Go's os/exec tests.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// If GO_HELPER_ERROR is set, exit with error
	if os.Getenv("GO_HELPER_ERROR") == "true" {
		os.Exit(1)
	}

	// If GO_HELPER_TIMEOUT is set, sleep to simulate a process that doesn't complete quickly
	if os.Getenv("GO_HELPER_TIMEOUT") == "true" {
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	// Default behavior
	os.Exit(0)
}
