package test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
)

// CommandRunner allows testing of Cobra commands
type CommandRunner struct {
	t         *testing.T
	cmd       *cobra.Command
	args      []string
	outputBuf *bytes.Buffer
	errorBuf  *bytes.Buffer
}

// NewCommandRunner creates a new command runner for testing
func NewCommandRunner(t *testing.T, cmd *cobra.Command) *CommandRunner {
	return &CommandRunner{
		t:         t,
		cmd:       cmd,
		outputBuf: new(bytes.Buffer),
		errorBuf:  new(bytes.Buffer),
	}
}

// WithArgs adds arguments to the command
func (r *CommandRunner) WithArgs(args ...string) *CommandRunner {
	r.args = args
	return r
}

// Run executes the command and captures output
func (r *CommandRunner) Run() (string, string, error) {
	// Clone the command to avoid modifying the original
	cmd := &cobra.Command{}
	*cmd = *r.cmd

	// Redirect to our buffers
	cmd.SetOut(r.outputBuf)
	cmd.SetErr(r.errorBuf)

	// Reset buffers
	r.outputBuf.Reset()
	r.errorBuf.Reset()

	// Set command args
	cmd.SetArgs(r.args)

	// Run the command
	err := cmd.Execute()

	// Return the output and error
	return r.outputBuf.String(), r.errorBuf.String(), err
}

// ExecutableTester is used to test the actual executable
type ExecutableTester struct {
	t            *testing.T
	execPath     string
	envVars      []string
	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer
}

// NewExecutableTester creates a new tester for CLI executables
func NewExecutableTester(t *testing.T, execPath string) *ExecutableTester {
	return &ExecutableTester{
		t:            t,
		execPath:     execPath,
		stdoutBuffer: new(bytes.Buffer),
		stderrBuffer: new(bytes.Buffer),
	}
}

// WithEnv adds environment variables to the command
func (e *ExecutableTester) WithEnv(envVars []string) *ExecutableTester {
	e.envVars = envVars
	return e
}

// Run executes the CLI with given arguments
func (e *ExecutableTester) Run(args ...string) (string, string, error) {
	cmd := exec.Command(e.execPath, args...)

	// Set environment variables
	if len(e.envVars) > 0 {
		cmd.Env = append(os.Environ(), e.envVars...)
	}

	// Clear buffers
	e.stdoutBuffer.Reset()
	e.stderrBuffer.Reset()

	// Capture stdout and stderr
	cmd.Stdout = e.stdoutBuffer
	cmd.Stderr = e.stderrBuffer

	// Run command
	err := cmd.Run()

	return e.stdoutBuffer.String(), e.stderrBuffer.String(), err
}

// CommandAvailabilityChecker defines a function type for checking if commands are available
type CommandAvailabilityChecker func(string) bool

// CommandExistsMock creates a mock for isCommandAvailable function
func CommandExistsMock(cmds map[string]bool) CommandAvailabilityChecker {
	return func(cmd string) bool {
		available, exists := cmds[cmd]
		if !exists {
			return false
		}
		return available
	}
}

// ExecCommandMock holds the data for mocking exec commands
type ExecCommandMock struct {
	// For tracking calls
	CmdCalls []string
	ArgCalls [][]string

	// For configuring responses
	OutputToReturn map[string][]byte
	ErrorToReturn  map[string]error
}

// NewExecCommandMock creates a new mock for exec commands
func NewExecCommandMock() *ExecCommandMock {
	return &ExecCommandMock{
		CmdCalls:       make([]string, 0),
		ArgCalls:       make([][]string, 0),
		OutputToReturn: make(map[string][]byte),
		ErrorToReturn:  make(map[string]error),
	}
}

// GetCommandKey creates a key for lookup in the output/error maps
func (e *ExecCommandMock) GetCommandKey(cmd string, args ...string) string {
	key := cmd
	for _, arg := range args {
		key += ":" + arg
	}
	return key
}

// SetOutput sets the output for a command
func (e *ExecCommandMock) SetOutput(output []byte, cmd string, args ...string) {
	key := e.GetCommandKey(cmd, args...)
	e.OutputToReturn[key] = output
}

// SetError sets the error for a command
func (e *ExecCommandMock) SetError(err error, cmd string, args ...string) {
	key := e.GetCommandKey(cmd, args...)
	e.ErrorToReturn[key] = err
}

// CaptureOutput captures stdout and stderr during a function execution
func CaptureOutput(f func()) (string, string) {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Replace stdout and stderr with pipes
	os.Stdout = wOut
	os.Stderr = wErr

	// Call the function
	f()

	// Close the write end of the pipes
	wOut.Close()
	wErr.Close()

	// Read the output
	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	// Restore original stdout and stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return bufOut.String(), bufErr.String()
}
