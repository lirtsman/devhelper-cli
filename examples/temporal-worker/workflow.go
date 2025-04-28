package main

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// SampleWorkflow is a simple workflow that demonstrates the basic functionality
// of Temporal. It takes a name as input and returns a greeting.
func SampleWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Sample workflow started", "name", name)

	// Define workflow timeout
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	var result string
	err := workflow.ExecuteActivity(ctx, SampleActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity execution failed", "error", err)
		return "", err
	}

	logger.Info("Sample workflow completed", "result", result)
	return result, nil
}

// SampleActivity is a simple activity that takes a name as input and returns a greeting.
func SampleActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Sample activity started", "name", name)

	// Simulate some work by sleeping
	time.Sleep(time.Second)

	result := "Hello, " + name + "!"
	logger.Info("Sample activity completed", "result", result)
	return result, nil
}
