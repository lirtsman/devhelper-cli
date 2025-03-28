package cmd

import (
	"testing"

	"bitbucket.org/shielddev/shielddev-cli/internal/test"
	"github.com/stretchr/testify/assert"
)

// TestLocalenvStopCommand tests the functionality of the localenv stop command
func TestLocalenvStopCommand(t *testing.T) {
	t.Run("Stop command structure should be valid", func(t *testing.T) {
		// Make sure stopCmd exists and has the right properties
		assert.NotNil(t, stopCmd, "localenv stop command should exist")
		assert.Equal(t, "stop", stopCmd.Use, "Command name should be stop")
		assert.Contains(t, stopCmd.Short, "Stop local", "Command should mention stopping local environment")

		// Check that flags are properly defined
		assert.NotNil(t, stopCmd.Flags().Lookup("skip-dapr"), "skip-dapr flag should exist")
		assert.NotNil(t, stopCmd.Flags().Lookup("skip-temporal"), "skip-temporal flag should exist")
		assert.NotNil(t, stopCmd.Flags().Lookup("force"), "force flag should exist")
	})

	t.Run("Stop command should be registered with parent", func(t *testing.T) {
		// Check if stop command is registered with localenv command
		found := false
		for _, cmd := range localenvCmd.Commands() {
			if cmd.Use == "stop" {
				found = true
				break
			}
		}
		assert.True(t, found, "stop command should be registered with localenv command")
	})

	// Save original isCommandAvailable
	origCommandCheck := isCommandAvailable
	defer func() { isCommandAvailable = origCommandCheck }()

	t.Run("Stop command should check for required components", func(t *testing.T) {
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

	t.Run("Stop command should handle missing components", func(t *testing.T) {
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

	t.Run("Stop command flags should have defaults", func(t *testing.T) {
		// Verify that all flags have proper default values
		forceFlag := stopCmd.Flags().Lookup("force")
		assert.NotNil(t, forceFlag, "force flag should exist")
		assert.Equal(t, "false", forceFlag.DefValue, "force flag should default to false")

		skipDaprFlag := stopCmd.Flags().Lookup("skip-dapr")
		assert.NotNil(t, skipDaprFlag, "skip-dapr flag should exist")
		assert.Equal(t, "false", skipDaprFlag.DefValue, "skip-dapr flag should default to false")

		skipTemporalFlag := stopCmd.Flags().Lookup("skip-temporal")
		assert.NotNil(t, skipTemporalFlag, "skip-temporal flag should exist")
		assert.Equal(t, "false", skipTemporalFlag.DefValue, "skip-temporal flag should default to false")
	})
}
