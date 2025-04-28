#!/bin/bash

echo "Starting the Temporal sample worker locally..."

# Check if Redis is available
if redis-cli ping > /dev/null 2>&1; then
  echo "Redis is available. Worker will use configuration from Redis if found."
else
  echo "Redis is not available. Worker will use default configuration."
fi

# Check if Temporal server is available
if curl -s http://localhost:7233/health > /dev/null 2>&1; then
  echo "Temporal server is available."
else
  echo "Warning: Temporal server doesn't seem to be running at localhost:7233."
  echo "Make sure Temporal server is running before starting the worker."
fi

# Build and run the worker
echo "Building and running the worker..."
go run . 