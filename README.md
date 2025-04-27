# DevHelper CLI

A comprehensive command-line tool for managing development environments.

## Overview

DevHelper CLI is designed to streamline and automate the setup and management of local development environments. It provides developers with a unified interface to manage various tools and components needed for development, making it easier to maintain consistent environments across teams.

## Features

- **Local Environment Management**: Initialize, start, stop, and manage local development environments with a single command
- **Kind Environment Management**: Provision and manage Kubernetes clusters with Temporal, Dapr, and Redis using Kind
- **Component Integration**: Seamlessly work with Dapr, Temporal, and OpenSearch
- **Log Management**: View, follow, and clean logs for various components
- **Configuration**: Flexible configuration via YAML files for consistent environments
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Container Support**: Integration with Podman for container orchestration

## Installation

### Using Homebrew (macOS/Linux)

The easiest way to install DevHelper CLI is using Homebrew:

```bash
# Add the tap (first time only)
brew tap ShieldFC-RD/devhelper-cli

# Install the CLI
brew install devhelper-cli

# Verify installation
devhelper-cli version
```

### From Source

```bash
# Clone the repository
git clone https://github.com/ShieldFC-RD/devhelper-cli.git
cd devhelper-cli

# Build the CLI
go build -o devhelper-cli

# Move to a location in your PATH (optional)
sudo mv devhelper-cli /usr/local/bin/
```

### Using Go

```bash
go install github.com/ShieldFC-RD/devhelper-cli@latest
```

### From GitHub Releases

You can download pre-built binaries from the [GitHub Releases page](https://github.com/ShieldFC-RD/devhelper-cli/releases).

#### Linux

```bash
# Download the latest release
curl -L "https://github.com/ShieldFC-RD/devhelper-cli/releases/latest/download/devhelper-cli-linux-amd64" -o devhelper-cli

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
curl -L "https://github.com/ShieldFC-RD/devhelper-cli/releases/latest/download/devhelper-cli-darwin-amd64" -o devhelper-cli

# Make it executable
chmod +x devhelper-cli

# Move to a location in your PATH
sudo mv devhelper-cli /usr/local/bin/

# Verify installation
devhelper-cli version
```

#### Windows

1. Download `devhelper-cli-windows-amd64.exe` from the [latest release](https://github.com/ShieldFC-RD/devhelper-cli/releases/latest)
2. Rename it to `devhelper-cli.exe` (optional)
3. Add it to a location in your PATH or use it from any directory by providing the full path

### Upgrading

To upgrade to the latest version:

#### Using Homebrew

```bash
brew update
brew upgrade devhelper-cli
```

#### Manual Upgrade

```bash
# Linux/macOS
curl -L "https://github.com/ShieldFC-RD/devhelper-cli/releases/latest/download/devhelper-cli-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64" -o /usr/local/bin/devhelper-cli
chmod +x /usr/local/bin/devhelper-cli
```

For Windows, download the new version from the [latest release](https://github.com/ShieldFC-RD/devhelper-cli/releases/latest) and replace your existing executable.

## Usage

```bash
# Display help information
devhelper-cli --help

# Show version information (including build date and commit hash)
devhelper-cli version

# Initialize local development environment
devhelper-cli localenv init

# Start local development environment
devhelper-cli localenv start

# Start without OpenSearch
devhelper-cli localenv start --skip-opensearch

# Stream Temporal server logs to terminal
devhelper-cli localenv start --stream-logs

# Check local environment status
devhelper-cli localenv status

# Stop local development environment
devhelper-cli localenv stop

# Stop specific components
devhelper-cli localenv stop --skip-dapr --skip-temporal

# Stop and clean up log files
devhelper-cli localenv stop --clean-logs

# View component logs
devhelper-cli localenv logs temporal

# Follow component logs in real-time
devhelper-cli localenv logs temporal -f

# Initialize Kind-based environment configuration
devhelper-cli kindenv init

# Start Kind-based environment
devhelper-cli kindenv start

# Check Kind-based environment status
devhelper-cli kindenv status

# Stop/delete Kind-based environment
devhelper-cli kindenv stop --delete
```

## Configuration

DevHelper CLI can be configured using:

1. Configuration file: `$HOME/.devhelper-cli.yaml`
2. Environment variables: All environment variables should be prefixed with `DEVHELPER_`
3. Command-line flags

### Local Environment Configuration

The local development environment can be configured using `localenv.yaml` in your project directory:

```yaml
# localenv.yaml
tools:
  podman:
    path: /usr/local/bin/podman
    version: 5.4.1
  kind:
    path: /usr/local/bin/kind
    version: ""
  daprCli:
    path: /usr/local/bin/dapr
    version: 1.14.1
  temporalCli:
    path: /usr/local/bin/temporal
    version: 1.2.0
components:
  dapr:
    enabled: true
    dashboard: true
    dashboardPort: 8080
    zipkinPort: 9411
  temporal:
    enabled: true
    namespace: default
    uiPort: 8233
    grpcPort: 7233
  openSearch:
    enabled: true
    version: 2.17.1
    port: 9200
    dashboardPort: 5601
```

## Supported Components

DevHelper CLI supports several key components for local development:

### Dapr
- Runtime initialization and management
- Dashboard access and configuration
- Component integration

### Temporal
- Server initialization and management
- UI and gRPC endpoint configuration
- Namespace management

### OpenSearch
- Container management via Podman
- Dashboard configuration
- Security and access settings

## Kind-based Environment Management

DevHelper CLI includes the `kindenv` command for provisioning and managing a local Kind-based Kubernetes development environment. This environment comes with all necessary components for Shield application development.

### Prerequisites

- [Kind](https://kind.sigs.k8s.io/) - Kubernetes in Docker
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Kubernetes command-line tool
- [Helm](https://helm.sh/) - Kubernetes package manager
- Docker or Podman - Container engine
- AWS CLI (optional) - For AWS ECR integration

### Usage

```bash
# Initialize kindenv configuration and set up Helm repositories
devhelper-cli kindenv init

# Initialize with custom configuration file
devhelper-cli kindenv init --output custom-kindenv.yaml --force

# Start the Kind environment with all components
devhelper-cli kindenv start

# Start environment with specific options
devhelper-cli kindenv start --skip-temporal --skip-redis --use-aws-ecr

# Check the status of kindenv components
devhelper-cli kindenv status

# Stop/delete the Kind environment
devhelper-cli kindenv stop --delete
```

### Configuration

The `kindenv` command uses a YAML configuration file (by default `kindenv.yaml` in the current directory) that includes settings for:

- Tool paths and versions (Kind, kubectl, Helm, etc.)
- Cluster configuration (name, port mappings)
- Component settings (Temporal, Redis, Dapr)
- Image registry configuration (Docker Hub, AWS ECR)
- Secret management

### Detailed Command Reference

#### kindenv init

This command sets up the initial configuration for a Kind-based Kubernetes environment and prepares the necessary Helm repositories.

**What it does:**

1. Creates a default `kindenv.yaml` configuration file with:
   - Auto-detected paths and versions for required tools (Kind, kubectl, Helm, Docker/Podman, AWS CLI)
   - Default configuration for Temporal, Redis, and Dapr components
   - Port mappings for accessing services from the host machine
   - AWS ECR integration settings (if enabled)

2. Sets up required Helm repositories:
   - Temporal repository (https://go.temporal.io/helm-charts)
   - Dapr repository (https://dapr.github.io/helm-charts)
   - Bitnami repository (https://charts.bitnami.com/bitnami) for Redis
   - Shield repository (https://harbor.shieldfis.com/chartrepo/stable)

3. Updates Helm repositories to ensure the latest charts are available

**Options:**

- `--output`, `-o`: Specify a different output path for the configuration file (default: `kindenv.yaml`)
- `--force`, `-f`: Force overwrite if the configuration file already exists
- `--skip-repos`: Skip adding and updating Helm repositories
- `--name`: Set a specific cluster name (default: uses current directory name)

Example usage:
```bash
# Create configuration with default settings
devhelper-cli kindenv init

# Create configuration with custom output path and force overwrite
devhelper-cli kindenv init -o custom-config.yaml -f

# Create configuration with a specific cluster name
devhelper-cli kindenv init --name my-cluster

# Create configuration without setting up Helm repositories
devhelper-cli kindenv init --skip-repos
```

**Example Configuration File**

The `kindenv init` command generates a configuration file similar to the following:

```yaml
tools:
  podman:
    path: /opt/homebrew/bin/podman
    version: 5.4.1
  docker:
    path: ""
    version: ""
  kind:
    path: /opt/homebrew/bin/kind
    version: v0.20.0
  kubectl:
    path: /opt/homebrew/bin/kubectl
    version: v1.29.0
  helm:
    path: /opt/homebrew/bin/helm
    version: v3.13.1
  aws:
    path: /usr/local/bin/aws
    version: aws-cli/2.11.2
cluster:
  name: devhelper-cli
  createIfNotExists: true
  mapPorts:
    - containerPort: 30080
      hostPort: 8080
      protocol: TCP
    - containerPort: 30733
      hostPort: 7233
      protocol: TCP
    - containerPort: 30679
      hostPort: 6379
      protocol: TCP
components:
  temporal:
    enabled: true
    namespace: temporal
    chartVersion: 0.62.0
    nodePorts:
      web: 30080
      frontend: 30733
  redis:
    enabled: true
    nodePorts:
      redis: 30679
    chartVersion: 17.3.7
    auth:
      enabled: false
  dapr:
    enabled: true
    chartVersion: 1.15.3
    nodePorts:
      dashboard: 30479
    logLevel: debug
    mtls:
      enabled: false
    ha:
      enabled: false
  temporalWorkerOperator:
    enabled: true
    chartVersion: 0.1.46-dev
images:
  skipPull: false
  dockerHub:
    username: ""
    password: ""
  useAwsEcr: false
  aws:
    region: eu-west-1
    ecrRegistry: ""
    profile: ""
secrets:
  mysql:
    enabled: true
    name: mysql-credentials
    namespace: default
    username: root
    password: password
```

You can customize this file to fit your specific requirements before starting the Kind environment.

#### kindenv start

This command creates and configures a Kind cluster with all the required components for Shield application development.

**What it does:**

1. Creates a Kind cluster if it doesn't exist, configuring:
   - Port mappings based on your configuration
   - Node labels for proper service exposure

2. Sets up required infrastructure:
   - Configures Kubernetes context to point to the Kind cluster
   - Sets up AWS ECR credentials if enabled (for pulling container images)
   - Creates necessary MySQL credentials as Kubernetes secrets

3. Installs and configures components in the correct order:
   - Redis (with optional authentication)
   - Dapr runtime (with configurable log level and mTLS settings)
   - Temporal server (with web UI and frontend services exposed)
   - Temporal Worker Operator (with proper Redis integration)

4. Displays access information for installed services:
   - URLs and ports for Temporal Web UI
   - Connection details for Temporal Frontend
   - Connection details for Redis

**Options:**

- `--skip-temporal`: Skip installing Temporal server
- `--skip-dapr`: Skip installing Dapr runtime
- `--skip-redis`: Skip installing Redis
- `--skip-temporal-worker-operator`: Skip installing Temporal Worker Operator
- `--use-aws-ecr`: Enable AWS ECR integration for pulling images
- `--aws-profile`: Specify AWS profile to use for ECR access
- `--name`: Override cluster name from configuration file
- `--config`, `-f`: Path to configuration file (default: `kindenv.yaml`)
- `--verbose`, `-v`: Enable verbose output

Example usage:
```bash
# Start with default settings from kindenv.yaml
devhelper-cli kindenv start

# Start with AWS ECR integration and specific profile
devhelper-cli kindenv start --use-aws-ecr --aws-profile my-profile

# Start with only specific components
devhelper-cli kindenv start --skip-temporal --skip-temporal-worker-operator

# Start with custom configuration file and verbose output
devhelper-cli kindenv start -f custom-config.yaml -v

# Start with custom cluster name
devhelper-cli kindenv start --name my-custom-cluster
```

#### kindenv status

This command checks and displays the status of a Kind-based Kubernetes environment and its components.

**What it does:**

1. Verifies if the Kind cluster exists and is running
2. Checks the status of each component:
   - Temporal server (deployment and namespace status)
   - Redis (pod and namespace status)
   - Dapr runtime (deployment status)
   - Temporal Worker Operator (deployment status and CRD installation)

3. Displays access information for installed services:
   - URLs and ports for Temporal Web UI
   - Connection details for Temporal Frontend
   - Connection details for Redis

4. In verbose mode, provides additional information:
   - Available Kubernetes namespaces
   - Installed TemporalWorker resources
   - Installed TemporalNamespace resources
   - Detailed port mapping configuration

**Options:**

- `--name`: Override cluster name from configuration file
- `--config`, `-f`: Path to configuration file (default: `kindenv.yaml`)
- `--verbose`, `-v`: Enable verbose output with additional details

Example usage:
```bash
# Check status with default settings
devhelper-cli kindenv status

# Check status with verbose output
devhelper-cli kindenv status --verbose

# Check status for a specific cluster name
devhelper-cli kindenv status --name my-custom-cluster

# Check status using a custom configuration file
devhelper-cli kindenv status -f custom-config.yaml
```

#### kindenv stop

This command stops and deletes a Kind-based Kubernetes environment.

**What it does:**

1. Verifies if the Kind cluster exists
2. If the cluster exists, stops and deletes it completely
3. Provides appropriate feedback based on cluster status

**Options:**

- `--name`: Override cluster name from configuration file
- `--config`, `-f`: Path to configuration file (default: `kindenv.yaml`)
- `--verbose`, `-v`: Enable verbose output with additional details

Example usage:
```bash
# Stop the cluster with default settings
devhelper-cli kindenv stop

# Stop a specific cluster by name
devhelper-cli kindenv stop --name my-custom-cluster

# Stop with verbose output
devhelper-cli kindenv stop --verbose

# Stop using a custom configuration file
devhelper-cli kindenv stop -f custom-config.yaml
```

For detailed information on all `kindenv` options and configuration, see [KINDENV.md](KINDENV.md).

## Roadmap

DevHelper CLI is actively being developed with several planned features on the horizon:

### Remote Environment Management
- **Remote Deployment**: Ability to deploy applications to remote development environments
- **Environment Synchronization**: Keep local and remote environments in sync
- **Remote Debugging**: Tools for debugging applications in remote environments

### Project Initialization
- **Repository Templates**: Initialize new projects with proper repository structure and boilerplate
- **Framework Integration**: Quick setup for common frameworks and libraries
- **Best Practice Enforcement**: Built-in configurations that follow best practices

### Release Management
- **App-of-Apps Integration**: Prepare new components for releases through pull requests to app-of-apps
- **Deployment Configuration**: Generate proper deployment configurations automatically
- **Release Automation**: Streamline the process of creating and deploying new releases

These features are still in development and will be rolled out in future releases.

## Development

For detailed information about the development process, environment setup, and known issues, please refer to the [DEVELOPMENT.md](DEVELOPMENT.md) file.

### Prerequisites

- Go 1.21 or later
- Make
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/ShieldFC-RD/devhelper-cli.git
cd devhelper-cli

# Install dependencies
make deps

# Build the CLI
make build

# Run tests
make test
```

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

Coverage reports are generated in the `./coverage` directory and can be viewed in your browser.

For more details about testing guidelines, refer to the [TESTING.md](TESTING.md) document.

### CI/CD

This project uses GitHub Actions for continuous integration and delivery:

- **CI Workflow**: Runs on all pull requests to `main` and `develop` branches to ensure code quality
  - Builds the application on multiple platforms (Linux, macOS, Windows)
  - Runs all tests with coverage reporting
  - Performs code linting using `go vet` (note: full linting with golangci-lint is temporarily disabled due to compatibility issues with Go 1.24+)

- **Release Workflow**: Triggered when a tag is pushed
  - Creates a GitHub Release with detailed release notes
  - Builds binaries for multiple platforms (including ARM64 support)
  - Generates checksums for binary verification
  - Automatically updates the Homebrew formula
  - Updates the CHANGELOG.md with commit details

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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- The Shield team for their support and contributions
- The Go community for their excellent tools and libraries
- All contributors who have helped make this project better

## Troubleshooting

### Log Management

If you encounter issues with log files:

1. **View Logs**: Use `devhelper-cli localenv logs [component]` to view current logs
2. **Rotate Logs**: Use `--clean-logs` with the stop command to delete log files when stopping components
3. **Follow Logs**: Use the `-f` flag with the logs command to follow logs in real-time
4. **Log Locations**: Log files are stored in `~/.logs/devhelper-cli/`

### Component Issues

#### Dapr

If Dapr fails to start:
1. Check Dapr installation with `dapr --version`
2. Verify Podman is running with `podman ps`
3. Check logs with `devhelper-cli localenv logs dapr`

#### Temporal

If Temporal fails to start:
1. Check if ports are already in use (default 8233 for UI, 7233 for gRPC)
2. Verify installation with `temporal --version`
3. Check logs with `devhelper-cli localenv logs temporal`

#### OpenSearch

If OpenSearch fails to start:
1. Verify Podman is running with `podman ps`
2. Check if ports 9200 and 5601 are available
3. Check container logs with `podman logs opensearch-node`
This typically indicates an issue with archive extraction in the pipeline. Common fixes include:

1. **Check archive integrity**: Ensure the source archive isn't corrupted during upload
2. **Verify storage permissions**: Ensure the CI runner has proper permissions to read/write to storage
3. **Check disk space**: Make sure the runner has sufficient disk space for extraction
4. **Examine CI cache**: Try clearing the CI cache, as cached archives might be corrupted 