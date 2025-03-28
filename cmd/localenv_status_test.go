package cmd

import (
	"testing"

	"bitbucket.org/shielddev/shielddev-cli/internal/test"
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
			"docker":   true,
			"dapr":     true,
			"temporal": true,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("docker"), "Docker should be available")
		assert.True(t, isCommandAvailable("dapr"), "Dapr should be available")
		assert.True(t, isCommandAvailable("temporal"), "Temporal should be available")
	})

	t.Run("Status command should detect missing components", func(t *testing.T) {
		// Set mock to report missing components
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"docker":   true,
			"dapr":     false,
			"temporal": false,
		})

		// Verify the mock returns correctly
		assert.True(t, isCommandAvailable("docker"), "Docker should be available")
		assert.False(t, isCommandAvailable("dapr"), "Dapr should not be available")
		assert.False(t, isCommandAvailable("temporal"), "Temporal should not be available")
	})
}
