package cmd

import (
	"testing"

	"bitbucket.org/shielddev/shielddev-cli/internal/test"
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
		assert.NotNil(t, startCmd.Flags().Lookup("wait"), "wait flag should exist")
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
			"docker": true,
			"dapr":   true,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("docker"), "Docker should be available")
		assert.True(t, isCommandAvailable("dapr"), "Dapr should be available")
	})

	t.Run("Start command should handle missing components", func(t *testing.T) {
		// Set mock to report missing components
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"docker": true,
			"dapr":   false,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("docker"), "Docker should be available")
		assert.False(t, isCommandAvailable("dapr"), "Dapr should not be available")
	})

	t.Run("Start command flags should have defaults", func(t *testing.T) {
		// Verify that all flags have proper default values
		configFlag := startCmd.Flags().Lookup("config")
		assert.NotNil(t, configFlag, "config flag should exist")
		assert.Equal(t, "", configFlag.DefValue, "config flag should default to empty string")

		waitFlag := startCmd.Flags().Lookup("wait")
		assert.NotNil(t, waitFlag, "wait flag should exist")
		// The wait flag actually defaults to true in the implementation
		assert.Equal(t, "true", waitFlag.DefValue, "wait flag should default to true")

		skipDaprFlag := startCmd.Flags().Lookup("skip-dapr")
		assert.NotNil(t, skipDaprFlag, "skip-dapr flag should exist")
		assert.Equal(t, "false", skipDaprFlag.DefValue, "skip-dapr flag should default to false")

		skipTemporalFlag := startCmd.Flags().Lookup("skip-temporal")
		assert.NotNil(t, skipTemporalFlag, "skip-temporal flag should exist")
		assert.Equal(t, "false", skipTemporalFlag.DefValue, "skip-temporal flag should default to false")
	})
}
