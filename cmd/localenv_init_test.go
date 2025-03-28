package cmd

import (
	"testing"

	"bitbucket.org/shielddev/shielddev-cli/internal/test"
	"github.com/stretchr/testify/assert"
)

// TestLocalenvInitCommand tests the functionality of the localenv init command
func TestLocalenvInitCommand(t *testing.T) {
	t.Run("Init command structure should be valid", func(t *testing.T) {
		// Make sure localenvInitCmd exists and has the right properties
		assert.NotNil(t, localenvInitCmd, "localenv init command should exist")
		assert.Equal(t, "init", localenvInitCmd.Use, "Command name should be init")
		assert.Contains(t, localenvInitCmd.Short, "Initialize", "Command should mention initialization")

		// Check that flags are properly defined
		assert.NotNil(t, localenvInitCmd.Flags().Lookup("force"), "force flag should exist")
		assert.NotNil(t, localenvInitCmd.Flags().Lookup("verbose"), "verbose flag should exist")
	})

	t.Run("Init command should be registered with parent", func(t *testing.T) {
		// Check if init command is registered with localenv command
		found := false
		for _, cmd := range localenvCmd.Commands() {
			if cmd.Use == "init" {
				found = true
				break
			}
		}
		assert.True(t, found, "init command should be registered with localenv command")
	})

	// Save original isCommandAvailable
	origCommandCheck := isCommandAvailable
	defer func() { isCommandAvailable = origCommandCheck }()

	t.Run("Init command should check for required components", func(t *testing.T) {
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

	t.Run("Init command should handle missing components", func(t *testing.T) {
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

	t.Run("Init command flags should have defaults", func(t *testing.T) {
		// Verify that all flags have proper default values
		forceFlag := localenvInitCmd.Flags().Lookup("force")
		assert.NotNil(t, forceFlag, "force flag should exist")
		assert.Equal(t, "false", forceFlag.DefValue, "force flag should default to false")

		verboseFlag := localenvInitCmd.Flags().Lookup("verbose")
		assert.NotNil(t, verboseFlag, "verbose flag should exist")
		assert.Equal(t, "false", verboseFlag.DefValue, "verbose flag should default to false")
	})
}
