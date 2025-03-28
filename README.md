# ShieldDev CLI

A comprehensive command-line interface for ShieldDev operations.

## Overview

ShieldDev CLI is a powerful tool designed to streamline and automate ShieldDev operations. It provides various commands to help developers and operators manage ShieldDev resources efficiently from the command line.

## Installation

### From Source

```bash
# Clone the repository
git clone https://bitbucket.org/shielddev/shielddev-cli.git
cd shielddev-cli

# Build the CLI
go build -o shielddev-cli

# Move to a location in your PATH (optional)
sudo mv shielddev-cli /usr/local/bin/
```

### Using Go

```bash
go install bitbucket.org/shielddev/shielddev-cli@latest
```

## Usage

```bash
# Display help information
shielddev-cli --help

# Show version
shielddev-cli version

# Deploy an application
shielddev-cli deploy app myapp --env prod --version 1.2.3

# Deploy a service
shielddev-cli deploy service myservice --env staging

# Check status of all resources
shielddev-cli status

# Check detailed status of applications in production
shielddev-cli status app --detailed --env prod

# Start local development environment
shielddev-cli localenv start

# Check local environment status
shielddev-cli localenv status

# Stop local development environment
shielddev-cli localenv stop
```

## Configuration

ShieldDev CLI can be configured using:

1. Configuration file: `$HOME/.shielddev-cli.yaml`
2. Environment variables: All environment variables should be prefixed with `SHIELDDEV_`
3. Command-line flags

Example configuration file:

```yaml
# ~/.shielddev-cli.yaml
verbose: true
api:
  endpoint: https://api.shielddev.example.com
  token: YOUR_API_TOKEN
```

## Environment Variables

- `SHIELDDEV_API_ENDPOINT`: API endpoint URL
- `SHIELDDEV_API_TOKEN`: API authentication token
- `SHIELDDEV_VERBOSE`: Enable verbose output (set to "true")

## Commands

### Global Flags

- `--config`: Path to config file (default is $HOME/.shielddev-cli.yaml)
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
  - `start`: Start the local development environment with Dapr and Temporal
    - Flags:
      - `--skip-dapr`: Skip starting Dapr runtime
      - `--skip-temporal`: Skip starting Temporal server
      - `--config`, `-c`: Path to environment configuration file
      - `--wait`: Wait for all components to be ready before exiting
  - `status`: Check status of the local development environment
  - `stop`: Stop the local development environment
    - Flags:
      - `--skip-dapr`: Skip stopping Dapr runtime
      - `--skip-temporal`: Skip stopping Temporal server
      - `--force`: Force stop all components even if errors occur

## Prerequisites for Local Development

To use the `localenv` commands, the following tools should be installed:

1. **Docker**: Required for running containerized services
   - Install from [Docker's official website](https://www.docker.com/get-started)

2. **Dapr CLI**: Used for managing the Dapr runtime
   - Install with: `curl -fsSL https://raw.githubusercontent.com/dapr/cli/master/install/install.sh | /bin/bash`
   - Or follow instructions at [Dapr's documentation](https://docs.dapr.io/getting-started/install-dapr-cli/)

3. **Temporal CLI**: Used for managing Temporal server
   - Install with: `curl -sSf https://temporal.download/cli.sh | sh`
   - Or follow instructions at [Temporal's documentation](https://docs.temporal.io/cli#install)

## Development

### Testing

The CLI includes a comprehensive test suite to ensure functionality works as expected. Tests are written using Go's standard testing package along with the Testify assertion library.

To run tests:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific tests
go test -v ./cmd -run TestLocalenv
```

The test suite includes:

- **Command structure tests**: Verify commands and subcommands exist with correct names and descriptions.
- **Flag validation tests**: Ensure all command flags are properly defined with correct default values.
- **Dependency validation tests**: Check that commands properly validate required external tools.
- **Test utilities**: The `internal/test` package provides helper functions for testing CLI commands.

For testing the CLI, we use a combination of:
- Unit tests for individual function behavior
- Integration tests for command structure and flag handling
- Mocking for external tools and command execution

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details. 