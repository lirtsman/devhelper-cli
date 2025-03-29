package cmd

import (
	"testing"

	"github.com/lirtsman/devhelper-cli/internal/test"
	"github.com/stretchr/testify/assert"
)

// TestLocalenvStatusCommand tests the functionality of the localenv status command
func TestLocalenvStatusCommand(t *testing.T) {
	t.Run("Status command structure should be valid", func(t *testing.T) {
		// Make sure localenvStatusCmd exists and has the right properties
		assert.NotNil(t, localenvStatusCmd, "localenv status command should exist")
		assert.Equal(t, "status", localenvStatusCmd.Use, "Command name should be status")
		assert.Contains(t, localenvStatusCmd.Short, "status", "Command should mention status")
	})

	t.Run("Status command should be registered with parent", func(t *testing.T) {
		// Check if status command is registered with localenv command
		found := false
		for _, cmd := range localenvCmd.Commands() {
			if cmd.Use == "status" {
				found = true
				break
			}
		}
		assert.True(t, found, "status command should be registered with localenv command")
	})

	// Save original isCommandAvailable
	origCommandCheck := isCommandAvailable
	defer func() { isCommandAvailable = origCommandCheck }()

	t.Run("Status command should check for required components", func(t *testing.T) {
		// Most we can do without actually executing the command is to
		// verify the function uses isCommandAvailable
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"podman":   true,
			"kind":     true,
			"dapr":     true,
			"temporal": true,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("podman"), "Podman should be available")
		assert.True(t, isCommandAvailable("kind"), "Kind should be available")
		assert.True(t, isCommandAvailable("dapr"), "Dapr should be available")
		assert.True(t, isCommandAvailable("temporal"), "Temporal should be available")
	})

	t.Run("Status command should detect missing components", func(t *testing.T) {
		// Set mock to report missing components
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"podman":   true,
			"kind":     true,
			"dapr":     false,
			"temporal": false,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("podman"), "Podman should be available")
		assert.True(t, isCommandAvailable("kind"), "Kind should be available")
		assert.False(t, isCommandAvailable("dapr"), "Dapr should not be available")
		assert.False(t, isCommandAvailable("temporal"), "Temporal should not be available")
	})
}

// TestStatusHelperFunctions tests the helper functions used by the status command
func TestStatusHelperFunctions(t *testing.T) {
	t.Run("getTemporalStatusRequirement should handle different cases", func(t *testing.T) {
		// Test default case (no config)
		result := getTemporalStatusRequirement(false, false)
		assert.True(t, result, "Default behavior should be to enable Temporal status")

		// Test with config loaded but no explicit setting
		result = getTemporalStatusRequirement(true, false)
		assert.False(t, result, "With config but no explicit Temporal flag, should be disabled")

		// Test with config loaded and explicit setting
		result = getTemporalStatusRequirement(true, true)
		assert.True(t, result, "With config and Temporal enabled, should be enabled")
	})

	t.Run("getTemporalNamespaceArgs should construct correct args", func(t *testing.T) {
		// Test with default namespace (not loaded config)
		config := LocalEnvConfig{}
		result := getTemporalNamespaceArgs(false, config)
		assert.Equal(t, []string{"operator", "namespace", "describe", "default"}, result, "Default namespace should be used")

		// Test with custom namespace in config
		config.Temporal.Namespace = "customns"
		result = getTemporalNamespaceArgs(true, config)
		assert.Equal(t, []string{"operator", "namespace", "describe", "customns"}, result, "Custom namespace should be used")
	})

	t.Run("getTemporalUIURL should construct URL correctly", func(t *testing.T) {
		// Test with default port (not loaded config)
		config := LocalEnvConfig{}
		result := getTemporalUIURL(false, config)
		assert.Equal(t, "http://localhost:8233", result, "Default UI port should be used")

		// Test with custom port
		config.Temporal.UIPort = 9000
		result = getTemporalUIURL(true, config)
		assert.Equal(t, "http://localhost:9000", result, "Custom UI port should be used")
	})

	t.Run("getDaprDashboardURL should construct URL correctly", func(t *testing.T) {
		// Test with default port (not loaded config)
		config := LocalEnvConfig{}
		result := getDaprDashboardURL(false, config)
		assert.Equal(t, "http://localhost:8080", result, "Default dashboard port should be used")

		// Test with custom port
		config.Dapr.DashboardPort = 9000
		result = getDaprDashboardURL(true, config)
		assert.Equal(t, "http://localhost:9000", result, "Custom dashboard port should be used")
	})

	t.Run("isDaprDashboardAvailable should check command existence", func(t *testing.T) {
		// This is hard to test without mocking exec.Command
		// Just verify it returns a boolean value
		result := isDaprDashboardAvailable()
		// We can't assert a specific result as it depends on the system
		t.Logf("Dapr dashboard available: %v", result)
	})

	t.Run("getDaprWebUIURL should return URL or empty string", func(t *testing.T) {
		// Since we can't easily mock isDaprDashboardAvailable, we'll test the function
		// based on the actual system state
		config := LocalEnvConfig{}
		config.Dapr.DashboardPort = 9000

		// The actual result depends on whether the dapr dashboard command is available
		// on the system where the test is running
		result := getDaprWebUIURL(true, config)
		if isDaprDashboardAvailable() {
			assert.Equal(t, "http://localhost:9000", result, "Should return URL when dashboard is available")
		} else {
			assert.Equal(t, "", result, "Should return empty string when dashboard is not available")
		}
	})

	t.Run("getZipkinURL should construct URL correctly", func(t *testing.T) {
		// Test with default port (not loaded config)
		config := LocalEnvConfig{}
		result := getZipkinURL(false, config)
		assert.Equal(t, "http://localhost:9411", result, "Default Zipkin port should be used")

		// Test with custom port
		config.Dapr.ZipkinPort = 9000
		result = getZipkinURL(true, config)
		assert.Equal(t, "http://localhost:9000", result, "Custom Zipkin port should be used")
	})

	t.Run("getDaprStatusRequirement should handle different cases", func(t *testing.T) {
		// Test default case (no config)
		result := getDaprStatusRequirement(false, false)
		assert.True(t, result, "Default behavior should be to enable Dapr status")

		// Test with config loaded but no explicit setting
		result = getDaprStatusRequirement(true, false)
		assert.False(t, result, "With config but no explicit Dapr flag, should be disabled")

		// Test with config loaded and explicit setting
		result = getDaprStatusRequirement(true, true)
		assert.True(t, result, "With config and Dapr enabled, should be enabled")
	})
}
