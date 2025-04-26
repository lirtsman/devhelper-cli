package progress

import (
	"fmt"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Tracker represents a progress tracker for operations
type Tracker interface {
	// Start starts the progress tracking
	Start()

	// Step updates the progress with a new step
	Step(message string)

	// StepWithInfo updates the progress with a new step and additional info
	StepWithInfo(message, info string)

	// Success marks the operation as successful
	Success(message string)

	// Fail marks the operation as failed
	Fail(message string)

	// Done finalizes the progress tracking
	Done()
}

type trackerImpl struct {
	operation string
	spinner   *spinner.Spinner
	mu        sync.Mutex
	active    bool
	lastStep  string
}

// NewTracker creates a new progress tracker for an operation
func NewTracker(operation string) Tracker {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("%s ", color.BlueString("•"))
	s.Suffix = fmt.Sprintf(" %s", operation)

	return &trackerImpl{
		operation: operation,
		spinner:   s,
		active:    false,
	}
}

// Start starts the progress tracking
func (t *trackerImpl) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		t.spinner.Start()
		t.active = true
	}
}

// Step updates the progress with a new step
func (t *trackerImpl) Step(message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastStep = message

	if t.active {
		t.spinner.Suffix = fmt.Sprintf(" %s: %s", t.operation, message)
	} else {
		fmt.Printf("%s %s: %s\n", color.BlueString("•"), t.operation, message)
	}
}

// StepWithInfo updates the progress with a new step and additional info
func (t *trackerImpl) StepWithInfo(message, info string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastStep = message

	if t.active {
		t.spinner.Suffix = fmt.Sprintf(" %s: %s (%s)", t.operation, message, info)
	} else {
		fmt.Printf("%s %s: %s (%s)\n", color.BlueString("•"), t.operation, message, info)
	}
}

// Success marks the operation as successful
func (t *trackerImpl) Success(message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active {
		t.spinner.Stop()
		t.active = false
	}

	if message == "" {
		message = t.lastStep
	}

	fmt.Printf("%s %s: %s\n", color.GreenString("✓"), t.operation, message)
}

// Fail marks the operation as failed
func (t *trackerImpl) Fail(message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active {
		t.spinner.Stop()
		t.active = false
	}

	fmt.Printf("%s %s: %s\n", color.RedString("✗"), t.operation, message)
}

// Done finalizes the progress tracking
func (t *trackerImpl) Done() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active {
		t.spinner.Stop()
		t.active = false
		fmt.Printf("%s %s: %s\n", color.GreenString("✓"), t.operation, "Completed")
	}
}
