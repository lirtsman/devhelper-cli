# Example Temporal Worker

This is a simple Temporal Worker example to demonstrate the usage of the `devhelper-cli tw` commands.

## Overview

This example includes:
- A sample Temporal workflow that processes a greeting
- Worker configuration via `tw.yaml`
- Redis configuration integration
- Dapr configuration integration
- Docker build support

## Prerequisites

- Temporal server running locally or in a Kind cluster
- Redis instance running locally or in a Kind cluster
- Go 1.21 or later
- Docker (for building the worker image)
- Dapr CLI (optional, for Dapr configuration)

## Usage with devhelper-cli tw commands

### Initialize the project (if starting from scratch)

```bash
# From an empty directory
devhelper-cli tw init

# Edit the generated tw.yaml to match your needs
```

### View the configuration

```bash
# View the current worker configuration
devhelper-cli tw config view

# View with a specific profile
devhelper-cli tw config view --profile dev
```

### Write configuration to Redis

```bash
# Write the worker settings to Redis
devhelper-cli tw config write

# Write to a specific Redis instance
devhelper-cli tw config write --redis localhost:6379
```

### Run the worker locally

```bash
# Run the worker using the Redis configuration
devhelper-cli tw run

# Run the worker without Redis configuration
devhelper-cli tw run --local
```

### Build the worker Docker image

```bash
# Build the worker Docker image with the default tag
devhelper-cli tw build

# Build with a specific tag
devhelper-cli tw build --tag v1.0.0

# Build with no cache
devhelper-cli tw build --no-cache

# Build and load into Kind cluster
devhelper-cli tw build --kind-load kind
```

### Deploy the worker to a Kind cluster

```bash
# Deploy to a Kind cluster
devhelper-cli tw deploy --kind

# Deploy to a specific Kind cluster and namespace
devhelper-cli tw deploy --kind --kind-cluster mykind --namespace default

# Build and deploy in one command
devhelper-cli tw deploy --kind --build
```

## Using without devhelper-cli

You can also use this worker without devhelper-cli:

### Running with Redis configuration

```bash
# Run the worker directly (will use Redis if available)
go run .

# Build the worker
go build -o worker .

# Run the built worker
./worker

# Use the provided script to run the worker
./run-local.sh
```

### Running with Dapr configuration

```bash
# Set up the Dapr configuration component and store settings in Redis
./create-dapr-config.sh

# Run the worker with Dapr
USE_DAPR_CONFIG=true dapr run --app-id temporal-worker --components-path ./components go run .

# Execute a sample workflow (separate terminal)
./run-workflow.sh
```

### Docker

```bash
# Build the Docker image
podman build -t temporal-worker .

# Run with Docker
podman run temporal-worker
```

## Configuration

The worker supports two configuration methods:

1. **Redis Configuration** (default): The worker reads configuration from Redis using the key format `configuration||{worker-name}`. The configuration is expected to be a JSON representation of the YAML defined in the `workerSettings` section of `tw.yaml`.

2. **Dapr Configuration** (optional): When enabled with `USE_DAPR_CONFIG=true`, the worker uses Dapr's configuration API to retrieve and subscribe to configuration updates. This requires a Dapr configuration component to be set up (see `create-dapr-config.sh`).

## Customization

You can customize this example by modifying:
1. Workflow logic in `workflow.go`
2. Worker configuration in `tw.yaml`
3. Configuration method in `main.go` 