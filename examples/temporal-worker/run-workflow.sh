#!/bin/bash

WORKFLOW_ID="sample-workflow-$(date +%s)"
TASK_QUEUE="sample-task-queue"
WORKFLOW="SampleWorkflow"

echo "Executing sample workflow with ID: $WORKFLOW_ID"

# Check if temporal CLI is installed
if ! command -v temporal &> /dev/null; then
  echo "Temporal CLI not found. Please install it first:"
  echo "See: https://docs.temporal.io/cli#install"
  exit 1
fi

# Check if Temporal server is available
if ! curl -s http://localhost:7233/health > /dev/null 2>&1; then
  echo "Error: Temporal server doesn't seem to be running at localhost:7233."
  echo "Make sure Temporal server is running before executing a workflow."
  exit 1
fi

# Execute the workflow
echo "Starting workflow execution..."
temporal workflow start \
  --workflow-id "$WORKFLOW_ID" \
  --task-queue "$TASK_QUEUE" \
  --type "$WORKFLOW" \
  --input '{"name": "Temporal"}'

echo "Workflow started! To check the status, run:"
echo "temporal workflow show --workflow-id $WORKFLOW_ID" 