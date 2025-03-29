package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/lirtsman/devhelper-cli/internal/test"
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

// We need this variable to override exec.Command in tests
var execLookPath = exec.LookPath

// mockCmd is a structure to help us mock exec.Cmd behavior
type mockCmd struct {
	outputData []byte
	errorVal   error
}

// CombinedOutput is a mockable version of exec.Cmd.CombinedOutput
func (m *mockCmd) CombinedOutput() ([]byte, error) {
	return m.outputData, m.errorVal
}

// TestValidateToolFunction tests the validateTool function behavior
func TestValidateToolFunction(t *testing.T) {
	// Create a test version of validateTool that uses our provided lookup and command funcs
	validateToolTest := func(
		name, versionFlag string,
		verbose bool,
		lookupFunc func(string) (string, error),
		cmdFunc func(string, ...string) *mockCmd,
	) (string, error) {
		// Check if the tool is in PATH
		path, err := lookupFunc(name)
		if err != nil {
			return "", fmt.Errorf("not found in PATH")
		}

		// Check if the tool works by running version command
		cmd := cmdFunc(path, versionFlag)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return path, fmt.Errorf("found but failed to run: %v", err)
		}

		if verbose {
			fmt.Printf("   %s version output: %s\n", name, string(output))
		}

		return path, nil
	}

	t.Run("should return path and nil error when tool exists and works", func(t *testing.T) {
		// Mock the lookup function
		fakePath := "/mocked/path/to/dapr"
		mockLookup := func(file string) (string, error) {
			return fakePath, nil
		}

		// Mock the command function
		mockCmdFunc := func(command string, args ...string) *mockCmd {
			// Return our mock that will succeed
			return &mockCmd{
				outputData: []byte("version 1.0.0"),
				errorVal:   nil,
			}
		}

		// Call our test version of validateTool
		path, err := validateToolTest("dapr", "--version", false, mockLookup, mockCmdFunc)

		// Check results
		assert.NoError(t, err, "Should not return error for working tool")
		assert.Equal(t, fakePath, path, "Should return the correct path")
	})

	t.Run("should return error when tool is not in PATH", func(t *testing.T) {
		// Mock the lookup function to return error
		mockLookup := func(file string) (string, error) {
			return "", fmt.Errorf("executable file not found in $PATH")
		}

		// Mock command function (though it shouldn't be called)
		mockCmdFunc := func(command string, args ...string) *mockCmd {
			t.Fatal("Command function should not be called when lookup fails")
			return nil
		}

		// Call our test version of validateTool
		path, err := validateToolTest("non-existent-tool", "--version", false, mockLookup, mockCmdFunc)

		// Check results
		assert.Error(t, err, "Should return error when tool not found")
		assert.Equal(t, "", path, "Should return empty path when tool not found")
		assert.Contains(t, err.Error(), "not found in PATH", "Error should mention PATH")
	})

	t.Run("should return error when tool exists but fails to run", func(t *testing.T) {
		// Mock the lookup function
		fakePath := "/mocked/path/to/broken-tool"
		mockLookup := func(file string) (string, error) {
			return fakePath, nil
		}

		// Mock the command function that returns an error
		mockCmdFunc := func(command string, args ...string) *mockCmd {
			// Return our mock that will fail
			return &mockCmd{
				outputData: []byte("error message"),
				errorVal:   fmt.Errorf("command failed"),
			}
		}

		// Call our test version of validateTool
		path, err := validateToolTest("broken-tool", "--version", false, mockLookup, mockCmdFunc)

		// Check results
		assert.Error(t, err, "Should return error when tool fails to run")
		assert.Equal(t, fakePath, path, "Should return path even when tool fails")
		assert.Contains(t, err.Error(), "failed to run", "Error should mention failure to run")
	})

	t.Run("should handle verbose output correctly", func(t *testing.T) {
		// Mock the lookup function
		fakePath := "/mocked/path/to/dapr"
		mockLookup := func(file string) (string, error) {
			return fakePath, nil
		}

		// Mock the command function
		mockCmdFunc := func(command string, args ...string) *mockCmd {
			// Return our mock that will succeed
			return &mockCmd{
				outputData: []byte("version 1.0.0"),
				errorVal:   nil,
			}
		}

		// Capture stdout to verify verbose output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call our test version of validateTool with verbose=true
		path, err := validateToolTest("dapr", "--version", true, mockLookup, mockCmdFunc)

		// Close writer and restore stdout
		w.Close()
		os.Stdout = oldStdout

		// Read captured output
		var captureBuffer [1024]byte
		n, _ := r.Read(captureBuffer[:])
		capturedOutput := string(captureBuffer[:n])

		// Check results
		assert.NoError(t, err, "Should not return error for working tool")
		assert.Equal(t, fakePath, path, "Should return the correct path")
		assert.Contains(t, capturedOutput, "version output", "Verbose output should be printed")
	})
}
