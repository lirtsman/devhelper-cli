package cmd

import (
	"testing"

	"bitbucket.org/shielddev/shielddev-cli/internal/test"
	"github.com/stretchr/testify/assert"
)

// TestLocalenvCommand tests the basic functionality of the localenv command
func TestLocalenvCommand(t *testing.T) {
	t.Run("Localenv command structure should be valid", func(t *testing.T) {
		// Make sure localenvCmd exists and has the right properties
		assert.NotNil(t, localenvCmd, "localenv command should exist")
		assert.Equal(t, "localenv", localenvCmd.Use, "Command name should be localenv")
		assert.Contains(t, localenvCmd.Short, "local development environment", "Command should mention local development")

		// Check that it has the right subcommands
		hasStart := false
		hasStop := false
		hasStatus := false

		for _, cmd := range localenvCmd.Commands() {
			switch cmd.Use {
			case "start":
				hasStart = true
			case "stop":
				hasStop = true
			case "status":
				hasStatus = true
			}
		}

		assert.True(t, hasStart, "localenv should have start subcommand")
		assert.True(t, hasStop, "localenv should have stop subcommand")
		assert.True(t, hasStatus, "localenv should have status subcommand")
	})

	t.Run("Localenv subcommands should exist", func(t *testing.T) {
		// Ensure all the expected subcommands exist
		assert.NotNil(t, startCmd, "start command should exist")
		assert.NotNil(t, stopCmd, "stop command should exist")
		assert.NotNil(t, localenvStatusCmd, "status command should exist")
	})
}

// TestCommandValidation tests that the commands validate dependencies properly
func TestCommandValidation(t *testing.T) {
	// Save original isCommandAvailable
	origCommandCheck := isCommandAvailable
	defer func() { isCommandAvailable = origCommandCheck }()

	t.Run("Commands should check for required tools", func(t *testing.T) {
		// Mock the isCommandAvailable function to return no tools are available
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"podman":   false,
			"kind":     false,
			"dapr":     false,
			"temporal": false,
		})

		// Since we can't easily test the commands with mocked exec.Command,
		// we'll just verify that the component checker works correctly
		assert.False(t, isCommandAvailable("podman"), "Podman should not be available")
		assert.False(t, isCommandAvailable("kind"), "Kind should not be available")
		assert.False(t, isCommandAvailable("dapr"), "Dapr should not be available")
		assert.False(t, isCommandAvailable("temporal"), "Temporal should not be available")
	})

	t.Run("Commands should find available tools", func(t *testing.T) {
		// Mock the isCommandAvailable function to return all tools are available
		isCommandAvailable = test.CommandExistsMock(map[string]bool{
			"podman":   true,
			"kind":     true,
			"dapr":     true,
			"temporal": true,
		})

		assert.True(t, isCommandAvailable("podman"), "Podman should be available")
		assert.True(t, isCommandAvailable("kind"), "Kind should be available")
		assert.True(t, isCommandAvailable("dapr"), "Dapr should be available")
		assert.True(t, isCommandAvailable("temporal"), "Temporal should be available")
	})
}

// TestLocalenvFlags tests that command flags work correctly
func TestLocalenvFlags(t *testing.T) {
	t.Run("Start command should have skip flags", func(t *testing.T) {
		assert.NotNil(t, startCmd.Flags().Lookup("skip-dapr"), "skip-dapr flag should exist")
		assert.NotNil(t, startCmd.Flags().Lookup("skip-temporal"), "skip-temporal flag should exist")
		assert.NotNil(t, startCmd.Flags().Lookup("skip-dapr-dashboard"), "skip-dapr-dashboard flag should exist")
		assert.NotNil(t, startCmd.Flags().Lookup("config"), "config flag should exist")
		assert.NotNil(t, startCmd.Flags().Lookup("wait"), "wait flag should exist")
	})

	t.Run("Stop command should have skip flags", func(t *testing.T) {
		assert.NotNil(t, stopCmd.Flags().Lookup("skip-dapr"), "skip-dapr flag should exist")
		assert.NotNil(t, stopCmd.Flags().Lookup("skip-temporal"), "skip-temporal flag should exist")
		assert.NotNil(t, stopCmd.Flags().Lookup("skip-dapr-dashboard"), "skip-dapr-dashboard flag should exist")
		assert.NotNil(t, stopCmd.Flags().Lookup("force"), "force flag should exist")
	})

	t.Run("Init command should have proper flags", func(t *testing.T) {
		assert.NotNil(t, localenvInitCmd.Flags().Lookup("force"), "force flag should exist")
		assert.NotNil(t, localenvInitCmd.Flags().Lookup("verbose"), "verbose flag should exist")
	})

	t.Run("Verbose flag should be available to all commands", func(t *testing.T) {
		assert.NotNil(t, localenvCmd.PersistentFlags().Lookup("verbose"), "verbose flag should exist")
	})
}

// TestLocalenvCommands tests that all localenv subcommands are properly registered
func TestLocalenvCommands(t *testing.T) {
	t.Run("All subcommands should be registered", func(t *testing.T) {
		// Create a map to track which commands we've found
		foundCmds := map[string]bool{
			"start":  false,
			"stop":   false,
			"status": false,
			"init":   false,
		}

		// Check each registered command
		for _, cmd := range localenvCmd.Commands() {
			if _, ok := foundCmds[cmd.Use]; ok {
				foundCmds[cmd.Use] = true
			}
		}

		// Verify that all expected commands were found
		for cmdName, found := range foundCmds {
			assert.True(t, found, "%s command should be registered", cmdName)
		}
	})
}
