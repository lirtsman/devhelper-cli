# ShieldDev CLI

A comprehensive command-line interface for ShieldDev operations.

## Overview

ShieldDev CLI is a powerful tool designed to streamline and automate ShieldDev operations. It provides various commands to help developers and operators manage ShieldDev resources efficiently from the command line.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/lirtsman/devhelper-cli.git
cd devhelper-cli

# Build the CLI
go build -o devhelper-cli

# Move to a location in your PATH (optional)
sudo mv devhelper-cli /usr/local/bin/
```

### Using Go

```bash
go install github.com/lirtsman/devhelper-cli@latest
```

### From GitHub Releases

You can download pre-built binaries from the [GitHub Releases page](https://github.com/lirtsman/devhelper-cli/releases).

#### Linux

```bash
# Download the latest release
curl -L "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-linux-amd64" -o devhelper-cli

# Make it executable
chmod +x devhelper-cli

# Move to a location in your PATH
sudo mv devhelper-cli /usr/local/bin/

# Verify installation
devhelper-cli version
```

#### macOS

```bash
# Download the latest release
curl -L "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-darwin-amd64" -o devhelper-cli

# Make it executable
chmod +x devhelper-cli

# Move to a location in your PATH
sudo mv devhelper-cli /usr/local/bin/

# Verify installation
devhelper-cli version
```

#### Windows

1. Download `devhelper-cli-windows-amd64.exe` from the [latest release](https://github.com/lirtsman/devhelper-cli/releases/latest)
2. Rename it to `devhelper-cli.exe` (optional)
3. Add it to a location in your PATH or use it from any directory by providing the full path

### Upgrading

To upgrade to the latest version:

```bash
# Linux/macOS
curl -L "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64" -o /usr/local/bin/devhelper-cli
chmod +x /usr/local/bin/devhelper-cli
```

For Windows, download the new version from the [latest release](https://github.com/lirtsman/devhelper-cli/releases/latest) and replace your existing executable.

## Development

For detailed information about the development process, environment setup, and known issues, please refer to the [DEVELOPMENT.md](DEVELOPMENT.md) file.

### Testing

DevHelper CLI uses Go's built-in testing framework with additional support from:
- [Testify](https://github.com/stretchr/testify) for assertions and mocking
- Go's standard testing coverage tools

To run tests:

```bash
# Run all tests
make test

# Run tests with coverage information
make test-coverage

# Generate coverage report in HTML format
make test-coverage-html

# Show function-level coverage stats
make test-coverage-func
```

Coverage reports are generated in the `./coverage` directory:
- `coverage.out`: Raw coverage data
- `coverage.html`: HTML report with color-coded coverage visualization

For more details about testing guidelines, refer to the [TESTING.md](TESTING.md) document.

### CI/CD

This project uses GitHub Actions for continuous integration and delivery:

- **CI Workflow**: Runs on all pull requests to `main` and `develop` branches to ensure code quality
  - Builds the application on multiple platforms
  - Runs all tests with coverage reporting
  - Performs code linting using `go vet` (note: full linting with golangci-lint is temporarily disabled due to compatibility issues with Go 1.24+)

- **Release Workflow**: Triggered when a tag is pushed
  - Creates a GitHub Release
  - Builds binaries for multiple platforms
  - Attaches binaries to the release

- **Format Workflow**: Automatically formats Go code on push

### Contributing

We welcome contributions to the DevHelper CLI! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed information on how to contribute to the project, including:

- Branching strategy
- Versioning approach
- Commit message format
- Pull request process
- Development setup
- Release process

#### Branching Strategy

We follow a modified GitFlow branching strategy:

- `main` - Production code, tagged with releases
- `develop` - Integration branch for features
- Feature branches - `feature/description` for new features
- Bugfix branches - `bugfix/description` for bug fixes
- Release branches - `release/vX.Y.Z` for release preparation
- Hotfix branches - `hotfix/vX.Y.Z` for urgent fixes

#### Versioning

We follow [Semantic Versioning](https://semver.org/) for version numbers:

- **MAJOR** version for incompatible API changes (X.y.z)
- **MINOR** version for backward-compatible functionality (x.Y.z)
- **PATCH** version for backward-compatible bug fixes (x.y.Z)

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

## Usage

```bash
# Display help information
devhelper-cli --help

# Show version
devhelper-cli version

# Deploy an application
devhelper-cli deploy app myapp --env prod --version 1.2.3

# Deploy a service
devhelper-cli deploy service myservice --env staging

# Check status of all resources
devhelper-cli status

# Check detailed status of applications in production
devhelper-cli status app --detailed --env prod

# Initialize local development environment
devhelper-cli localenv init

# Start local development environment
devhelper-cli localenv start

# Check local environment status
devhelper-cli localenv status

# Stop local development environment
devhelper-cli localenv stop
```

## Configuration

ShieldDev CLI can be configured using:

1. Configuration file: `$HOME/.devhelper-cli.yaml`
2. Environment variables: All environment variables should be prefixed with `SHIELDDEV_`
3. Command-line flags

Example configuration file:

```yaml
# ~/.devhelper-cli.yaml
verbose: true
api:
  endpoint: https://api.shielddev.example.com
  token: YOUR_API_TOKEN
```

### Local Environment Configuration

The local development environment can be configured using `localenv.yaml` in your project directory:

```yaml
# localenv.yaml
components:
  dapr: true
  temporal: true
paths:
  podman: /usr/local/bin/podman
  kind: /usr/local/bin/kind
  dapr: /usr/local/bin/dapr
  temporal: /usr/local/bin/temporal
clusterName: shielddev-local
temporal:
  namespace: default            # Default namespace to use
  uiPort: 8233                  # Web UI port
  frontendIP: localhost         # Frontend IP/hostname
  grpcPort: 7233                # gRPC port for client connections
dapr:
  dashboardPort: 8080           # Dashboard port (if dashboard is installed)
  dashboardIP: localhost        # Dashboard IP/hostname
  zipkinPort: 9411             # Zipkin port for distributed tracing
  zipkinIP: localhost          # Zipkin IP/hostname
```

This configuration is created when you run `devhelper-cli localenv init` and can be customized:

- `components` - Controls which components will be started (true) or disabled (false)
- `paths` - Stores the paths to required tools, useful for validation on subsequent starts
- `clusterName` - The default cluster name for Kind (useful for script automation)
- `temporal` - Configuration for Temporal server including namespace, ports, and addresses

## Environment Variables

- `SHIELDDEV_API_ENDPOINT`: API endpoint URL
- `SHIELDDEV_API_TOKEN`: API authentication token
- `SHIELDDEV_VERBOSE`: Enable verbose output (set to "true")

## Commands

### Global Flags

- `--config`: Path to config file (default is $HOME/.devhelper-cli.yaml)
- `--verbose`: Enable verbose output

### Available Commands

- `version`: Show the CLI version information
- `deploy`: Deploy ShieldDev resources
  - `app`: Deploy an application
  - `service`: Deploy a service
- `status`: Check the status of ShieldDev resources
  - Accepts optional resource type: `app`, `service`, `infra`, or `all` (default)
  - Flags:
    - `--detailed`, `-d`: Show detailed status information
    - `--env`, `-e`: Target environment (dev, staging, prod)
    - `--watch`, `-w`: Watch status updates in real-time
- `localenv`: Manage local development environment
  - `init`: Initialize the local environment configuration
    - Flags:
      - `--force`: Overwrite existing configuration
      - `--cluster-name`: Name for the Kind cluster (default: shielddev-local)
  - `start`: Start the local development environment with Dapr and Temporal
    - Flags:
      - `--skip-dapr`: Skip starting Dapr runtime
      - `--skip-temporal`: Skip starting Temporal server
      - `--config`, `-c`: Path to environment configuration file (default: localenv.yaml)
      - `--wait`: Wait for all components to be ready before exiting
  - `status`: Check status of the local development environment
    - Flags:
      - `--config`, `-c`: Path to environment configuration file (default: localenv.yaml)
  - `stop`: Stop the local development environment
    - Flags:
      - `--skip-dapr`: Skip stopping Dapr runtime
      - `--skip-temporal`: Skip stopping Temporal server
      - `--config`, `-c`: Path to environment configuration file (default: localenv.yaml)
      - `--force`: Force stop all components even if errors occur

## Prerequisites for Local Development

To use the `localenv` commands, the following tools must be installed and properly configured:

1. **Podman**: Required for running containerized services
   - Install from [Podman's official website](https://podman.io/getting-started/installation)
   - Podman is a binary tool that needs to be properly installed and configured to run containers

2. **Kind (Kubernetes in Docker)**: Required for local Kubernetes clusters
   - Install with: `curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/`
   - At least one cluster must be created before using devhelper-cli: `kind create cluster --name my-cluster`
   - Kind is a CLI tool that creates and manages Kubernetes clusters running in Podman containers
   - Or follow instructions at [Kind's documentation](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

3. **Dapr CLI**: Used for managing the Dapr runtime
   - Install with: `curl -fsSL https://raw.githubusercontent.com/dapr/cli/master/install/install.sh | /bin/bash`
   - The CLI will initialize and start Dapr runtime services automatically
   - Or follow instructions at [Dapr's documentation](https://docs.dapr.io/getting-started/install-dapr-cli/)

4. **Temporal CLI**: Used for managing Temporal server
   - Install with: `curl -sSf https://temporal.download/cli.sh | sh`
   - The CLI will start the Temporal server as a service
   - Or follow instructions at [Temporal's documentation](https://docs.temporal.io/cli#install)

## Local Development Environment

When you run `devhelper-cli localenv init`, the CLI:

1. Checks for required tools (Podman, Kind, Dapr CLI, Temporal CLI)
2. Validates their functionality
3. Creates a configuration file (localenv.yaml) that remembers the tools and enabled components

When you run `devhelper-cli localenv start`, the CLI:

1. Loads the configuration file if it exists
2. Verifies Podman is available and can run containers
3. Verifies Kind is available and has clusters configured
4. Initializes and starts the Dapr runtime (if enabled)
5. Starts the Temporal server (if enabled)

### Temporal Server

The Temporal server is started in development mode with `temporal server start-dev`. This provides:

- **Temporal Web UI**: Available at http://localhost:8233
  - Use this interface to monitor and manage workflows
  - Track workflow execution history
  - Visualize workflow dependencies
  
- **Temporal API**: Available at http://localhost:7233
  - Used by applications to interact with Temporal
  - Default namespace: "default"
  - gRPC port: 7233 (for client applications)

- **Configuration**: You can customize the connection settings in `localenv.yaml`:
  ```yaml
  temporal:
    namespace: default       # Namespace for workflows
    uiPort: 8233             # Web UI port
    frontendIP: localhost    # Server address
    grpcPort: 7233           # gRPC port for client connections
  ```
  
- **Working with Temporal**:
  ```bash
  # List workflows
  temporal workflow list
  
  # View details of a specific workflow
  temporal workflow describe -w <workflow-id>
  
  # Start a new workflow
  temporal workflow start -w <workflow-id> -t <task-queue> -wt <workflow-type>
  ```

### Dapr Runtime

Dapr is initialized with `dapr init --container-runtime podman`, which provides a self-hosted mode using Podman as the container runtime:

- **Self-hosted Mode**: Dapr runs as a set of containers in your local Podman environment
  - Includes Redis for state management and pub/sub
  - Includes Zipkin for distributed tracing
  - Includes placement and scheduler services
  
- **Zipkin UI**: 
  - Available at http://localhost:9411
  - Provides distributed tracing for monitoring application performance
  - Visualize traces to identify bottlenecks and troubleshoot issues
  - Configure port and host in `localenv.yaml` under the `dapr` section
  
- **Dapr Dashboard** (Optional): 
  - Not installed by default, but can be installed with: `dapr dashboard install`
  - When installed, it's available at http://localhost:8080
  - Provides UI for monitoring and managing Dapr applications and components
  - Configure port and host in `localenv.yaml` under the `dapr` section
  
- **Podman Integration**: The CLI automatically uses Podman as the container runtime for Dapr
  - When Docker is not available, Podman is used for all container operations
  - Dapr version 1.8+ is required for Podman support
  
- **Working with Dapr**:
  ```bash
  # Check Dapr status
  dapr status
  
  # List running Dapr applications
  dapr list
  
  # Run an application with Dapr
  dapr run --app-id myapp --app-port 3000 -- node app.js
  
  # Stop a running Dapr application
  dapr stop --app-id myapp
  ```

- **Container Services**:
  - **dapr_redis**: Redis instance for state management and pub/sub
  - **dapr_placement**: Service for actor placement
  - **dapr_zipkin**: Zipkin instance for distributed tracing
  - **dapr_scheduler**: Scheduler service for distributed coordination

- **Future Kubernetes Integration**: When ready to deploy to Kubernetes, you can use:
  ```bash
  # Initialize Dapr in a Kubernetes cluster
  dapr init -k
  
  # Check Dapr status in Kubernetes
  dapr status -k
  ```

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details. 