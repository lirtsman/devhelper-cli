# DevHelper CLI

A comprehensive command-line interface for Shield operations.

## Overview

DevHelper CLI is a powerful tool designed to streamline and automate Shield operations. It provides various commands to help developers and operators manage Shield resources efficiently from the command line.

## Features

- **Deployment Management**: Deploy applications and services to different environments
- **Local Development**: Initialize, start, and manage local development environments
- **Configuration**: Flexible configuration via files, environment variables, or flags
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Resource Status**: Check the status of all your resources in one place
- **Integrated Search**: OpenSearch integration for powerful search and analytics capabilities

## Installation

### Using Homebrew (macOS/Linux)

The easiest way to install DevHelper CLI is using Homebrew:

```bash
# Add the tap (first time only)
brew tap lirtsman/devhelper-cli

# Install the CLI
brew install devhelper-cli

# Verify installation
devhelper-cli version
```

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

#### Using Homebrew

```bash
brew update
brew upgrade devhelper-cli
```

#### Manual Upgrade

```bash
# Linux/macOS
curl -L "https://github.com/lirtsman/devhelper-cli/releases/latest/download/devhelper-cli-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64" -o /usr/local/bin/devhelper-cli
chmod +x /usr/local/bin/devhelper-cli
```

For Windows, download the new version from the [latest release](https://github.com/lirtsman/devhelper-cli/releases/latest) and replace your existing executable.

## Usage

```bash
# Display help information
devhelper-cli --help

# Show version information (including build date and commit hash)
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

# Start without OpenSearch
devhelper-cli localenv start --skip-opensearch

# Check local environment status
devhelper-cli localenv status

# Stop local development environment
devhelper-cli localenv stop

# Stop only OpenSearch
devhelper-cli localenv stop --skip-dapr --skip-temporal
```

## Configuration

DevHelper CLI can be configured using:

1. Configuration file: `$HOME/.devhelper-cli.yaml`
2. Environment variables: All environment variables should be prefixed with `DEVHELPER_`
3. Command-line flags

Example configuration file:

```yaml
# ~/.devhelper-cli.yaml
verbose: true
api:
  endpoint: https://api.devhelper.example.com
  token: YOUR_API_TOKEN
```

### Local Environment Configuration

The local development environment can be configured using `localenv.yaml` in your project directory:

```yaml
# localenv.yaml
components:
  dapr: true
  temporal: true
  openSearch: true
paths:
  podman: /usr/local/bin/podman
  kind: /usr/local/bin/kind
  dapr: /usr/local/bin/dapr
  temporal: /usr/local/bin/temporal
temporal:
  namespace: default            # Default namespace to use
  uiPort: 8233                  # Web UI port
  frontendIP: localhost         # Frontend IP/hostname
opensearch:
  port: 9200                    # OpenSearch API port
  dashboardPort: 5601           # OpenSearch Dashboards port
```

## Development

For detailed information about the development process, environment setup, and known issues, please refer to the [DEVELOPMENT.md](DEVELOPMENT.md) file.

### Prerequisites

- Go 1.21 or later
- Make
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/lirtsman/devhelper-cli.git
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

- **Homebrew Update Workflow**: Automatically updates the Homebrew formula when a new release is published

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

### CI/CD Pipeline Issues

#### Tar Extraction Failures

If you encounter errors like:
```
Failed to restore: "/usr/bin/tar" failed with error: The process '/usr/bin/tar' failed with exit code 2
```

This typically indicates an issue with archive extraction in the pipeline. Common fixes include:

1. **Check archive integrity**: Ensure the source archive isn't corrupted during upload
2. **Verify storage permissions**: Ensure the CI runner has proper permissions to read/write to storage
3. **Check disk space**: Make sure the runner has sufficient disk space for extraction
4. **Examine CI cache**: Try clearing the CI cache, as cached archives might be corrupted

#### Format Workflow Specific Issues

If you see this error in the Format workflow:

```
Annotations
1 warning
Format
Failed to restore: "/usr/bin/tar" failed with error: The process '/usr/bin/tar' failed with exit code 2
```

This is likely related to Go's module cache. To fix this:

1. **Disable Go caching in the Format workflow**:
   - Edit `.github/workflows/format.yml`
   - Change the `setup-go` action configuration:
   ```yaml
   - name: Set up Go
     uses: actions/setup-go@v4
     with:
       go-version: '1.21'
       cache: false  # Change from true to false
   ```

2. **Alternative: Update the Format workflow to use a newer actions/setup-go version**:
   ```yaml
   - name: Set up Go
     uses: actions/setup-go@v4
     with:
       go-version: '1.21'
       cache-dependency-path: go.sum
   ``` 