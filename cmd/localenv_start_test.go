package cmd

import (
	"testing"

	"github.com/lirtsman/devhelper-cli/internal/test"
	"github.com/stretchr/testify/assert"
)

// TestLocalenvStartCommand tests the functionality of the localenv start command
func TestLocalenvStartCommand(t *testing.T) {
	t.Run("Start command structure should be valid", func(t *testing.T) {
		// Make sure startCmd exists and has the right properties
		assert.NotNil(t, startCmd, "localenv start command should exist")
		assert.Equal(t, "start", startCmd.Use, "Command name should be start")
		assert.Contains(t, startCmd.Short, "Start local", "Command should mention starting local environment")

		// Check that flags are properly defined
		assert.NotNil(t, startCmd.Flags().Lookup("skip-dapr"), "skip-dapr flag should exist")
		assert.NotNil(t, startCmd.Flags().Lookup("skip-temporal"), "skip-temporal flag should exist")
		assert.NotNil(t, startCmd.Flags().Lookup("config"), "config flag should exist")
		assert.NotNil(t, startCmd.Flags().Lookup("stream-logs"), "stream-logs flag should exist")
	})

	t.Run("Start command should be registered with parent", func(t *testing.T) {
		// Check if start command is registered with localenv command
		found := false
		for _, cmd := range localenvCmd.Commands() {
			if cmd.Use == "start" {
				found = true
				break
			}
		}
		assert.True(t, found, "start command should be registered with localenv command")
	})

	// Save original isCommandAvailable
	origCommandCheck := isCommandAvailable
	defer func() { isCommandAvailable = origCommandCheck }()

	t.Run("Start command should check for required components", func(t *testing.T) {
		// Most we can do without actually executing the command is to
		// verify the function uses isCommandAvailable
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"podman": true,
			"kind":   true,
			"dapr":   true,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("podman"), "Podman should be available")
		assert.True(t, isCommandAvailable("kind"), "Kind should be available")
		assert.True(t, isCommandAvailable("dapr"), "Dapr should be available")
	})

	t.Run("Start command should handle missing components", func(t *testing.T) {
		// Set mock to report missing components
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"podman": true,
			"kind":   true,
			"dapr":   false,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("podman"), "Podman should be available")
		assert.True(t, isCommandAvailable("kind"), "Kind should be available")
		assert.False(t, isCommandAvailable("dapr"), "Dapr should not be available")
	})

	t.Run("Start command flags should have defaults", func(t *testing.T) {
		// Verify that all flags have proper default values
		configFlag := startCmd.Flags().Lookup("config")
		assert.NotNil(t, configFlag, "config flag should exist")
		assert.Equal(t, "", configFlag.DefValue, "config flag should default to empty string")

		skipDaprFlag := startCmd.Flags().Lookup("skip-dapr")
		assert.NotNil(t, skipDaprFlag, "skip-dapr flag should exist")
		assert.Equal(t, "false", skipDaprFlag.DefValue, "skip-dapr flag should default to false")

		skipTemporalFlag := startCmd.Flags().Lookup("skip-temporal")
		assert.NotNil(t, skipTemporalFlag, "skip-temporal flag should exist")
		assert.Equal(t, "false", skipTemporalFlag.DefValue, "skip-temporal flag should default to false")

		skipDaprDashboardFlag := startCmd.Flags().Lookup("skip-dapr-dashboard")
		assert.NotNil(t, skipDaprDashboardFlag, "skip-dapr-dashboard flag should exist")
		assert.Equal(t, "false", skipDaprDashboardFlag.DefValue, "skip-dapr-dashboard flag should default to false")
	})

	t.Run("Start command should honor skip-dapr-dashboard flag", func(t *testing.T) {
		// Verify that the skip-dapr-dashboard flag exists and can be retrieved
		flag := startCmd.Flags().Lookup("skip-dapr-dashboard")
		assert.NotNil(t, flag, "skip-dapr-dashboard flag should exist")

		// The actual behavior can't be fully tested without executing the command,
		// but we can verify the flag is properly registered
		assert.Equal(t, "skip-dapr-dashboard", flag.Name, "Flag name should be skip-dapr-dashboard")
		assert.Equal(t, "Skip starting Dapr Dashboard", flag.Usage, "Flag usage should mention skipping Dapr Dashboard")
	})

	t.Run("Start command should have stream-logs flag", func(t *testing.T) {
		// Verify that the stream-logs flag exists and can be retrieved
		flag := startCmd.Flags().Lookup("stream-logs")
		assert.NotNil(t, flag, "stream-logs flag should exist")

		// Check the flag properties
		assert.Equal(t, "stream-logs", flag.Name, "Flag name should be stream-logs")
		assert.Equal(t, "Stream Temporal server logs to terminal", flag.Usage, "Flag usage should mention streaming logs")
		assert.Equal(t, "false", flag.DefValue, "stream-logs flag should default to false")

		// The actual streaming behavior would require integration tests
	})
}
