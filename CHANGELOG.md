# Changelog

All notable changes to this project will be documented in this file.

## [v0.4.0] - 2025-04-27

* Merge pull request #1 from ShieldFC-RD/feature/kindenv (568fa30)
* Enhance README and Kind environment configuration (91b4e7b)
* Remove service account configuration from Kind environment setup (8857348)
* Implement component status checks in Kind environment (2fafc95)
* Update ECR credentials setup and namespace in Kind environment (b72bd65)
* Enhance Kind environment setup with Temporal Worker Operator and AWS ECR improvements (72e428f)
* Enhance Kind environment commands with cluster name flag support (a258809)
* Enhance Kind environment configuration and improve port mapping handling (8dab10a)
* Update Kind environment configuration to enable AWS ECR and improve Helm repository management (7f58665)
* Refactor Kind environment setup and enhance cluster management (00c2ed2)
* Refactor Kind environment configuration and enhance ECR setup (841cb51)
* Update dependencies and enhance Kind environment configuration (549d856)
* Add AWS ECR and Kind environment management functionality (3b3de17)
* Add installation step to Makefile for copying the built binary to /usr/local/bin (f2323bc)
* Update CHANGELOG.md for v0.3.1 (8fbf80f)

## [v0.3.1] - 2025-04-08

* Update Homebrew tap references to ShieldFC-RD organization (da195a4)
* Update module references to ShieldFC-RD organization (0648cb6)
* Update go.mod and go.sum to include github.com/stretchr/objx v0.5.2 as an indirect dependency (62ee118)
* Add unit tests for tool installation and update functions in local environment management (b9b8a34)
* Merge branch 'main' of https://github.com/lirtsman/devhelper-cli (e34278b)
* Update CHANGELOG.md for v0.3.0 (6feea91)
* Prepare for v0.3.0 release (27febca)
* Update CHANGELOG.md to document the addition of the `--clean-logs` flag for the `localenv stop` command, enhancing log management capabilities. Revise README.md to reflect changes in tool descriptions and features, including improved local environment management and log handling instructions. [skip ci] (e08a544)
* Enhance local environment configuration management by introducing structured tool version requirements and improving component initialization logic. Refactor validation functions to support version detection and auto-installation of required tools. Update configuration structure to better encapsulate tool paths and versions, ensuring a more robust setup process for Dapr, Temporal, and OpenSearch components. (906422a)
* Refactor Temporal namespace creation logic in localenv_start.go to occur after server startup, improving reliability and user feedback during setup. (3c058f9)
* Update CHANGELOG.md for v0.2.3 (cb36cfd)
* Update health check command in localenv_start.go to use default admin credentials for OpenSearch. (99d4744)
* Add localenv.yaml to .gitignore to exclude local environment configuration files from version control. (70efde8)
* Update CHANGELOG.md, remove localenv.yaml, and refine README.md documentation. Adjust environment variable prefix in code and documentation to DEVHELPER_. (640911f)
* Add OpenSearch and OpenSearch Dashboard support to local environment, including configuration management, health checks, and improved logging. Replace Docker with Podman for container management and update related commands and status checks. (6735981)
* Add commands to start and stop local development environment components including Dapr, Temporal, and OpenSearch. Implement configuration loading, process management, and logging for improved user experience. (1ea78d8)
* Add OpenSearch support to local environment (18b0e83)
* Update CHANGELOG.md for v0.2.2 (ba01dcd)
* Fix tar extraction issues in Format workflow and add troubleshooting documentation (a4af48c)
* Refactor project branding from ShieldDev to Shield across all files, including updates to documentation and command descriptions. (af50eec)
* Fix environment variable prefix in README (SHIELDDEV_ → DEVHELPER_) (429371a)
* Update README.md with improved documentation and latest features (f3b8717)
* Update coverage comment to trigger language statistics update (b35dfb5)
* Add .gitattributes to fix language statistics (8996383)
* Add coverage files to .gitignore and remove from repository (6b88ad4)
* Update CHANGELOG.md for v0.2.1 (3267e07)
* Fix CHANGELOG update to properly handle detached HEAD state (ad2f163)
* Add automatic CHANGELOG.md updates to release workflow (5cf50ac)
* Add Git commit information to Homebrew formula (bb2fe85)
* Fix version information in Homebrew formula (6821dbe)
* Replace workflow dispatch with direct formula update (24fd66b)
* Use HOMEBREW_TAP_TOKEN to trigger Homebrew update workflow (5a3548f)
* Force Homebrew formula update after release (664edfc)
* Add manual trigger for Homebrew formula updater (5c21ceb)
* Fix GitHub release creation permissions (c3a8954)
* Fix 'Cannot open: File exists' errors by completely disabling Go caching (a5e2175)
* Fix cache issues in GitHub Actions workflows and add automatic Homebrew formula updater (ecd3fbb)
* Update README to clarify installation instructions and improve user guidance (9e8f8a5)
* Remove Homebrew tap directory from Git tracking (9764456)
* Clean up repository and update .gitignore (88ace05)
* Fix formula file (7dc501c)
* Initial commit: Add devhelper-cli formula (b7b88fe)
* Enhance README and CI/CD workflows with Homebrew installation and ARM64 support (6c2b67d)
* Enhance README and CI/CD workflows for installation instructions and artifact handling (bd7fda2)
* Enhance CI/CD workflows to separate version retrieval and build steps (b4a52b5)
* Refactor CI/CD workflows to explicitly define cache paths (76517f0)
* Update CI/CD workflows to use latest action versions (834a36c)
* Update Go version and enhance linting process (4ce3723)
* Enhance README with CI/CD details and contribution guidelines (5ce6541)

## [v0.3.0] - 2025-03-30

* Prepare for v0.3.0 release (27febca)
* Update CHANGELOG.md to document the addition of the `--clean-logs` flag for the `localenv stop` command, enhancing log management capabilities. Revise README.md to reflect changes in tool descriptions and features, including improved local environment management and log handling instructions. [skip ci] (e08a544)
* Enhance local environment configuration management by introducing structured tool version requirements and improving component initialization logic. Refactor validation functions to support version detection and auto-installation of required tools. Update configuration structure to better encapsulate tool paths and versions, ensuring a more robust setup process for Dapr, Temporal, and OpenSearch components. (906422a)
* Refactor Temporal namespace creation logic in localenv_start.go to occur after server startup, improving reliability and user feedback during setup. (3c058f9)
* Update CHANGELOG.md for v0.2.3 (cb36cfd)

## [Unreleased]

### Added
- Add `--clean-logs` flag to `localenv stop` command to remove log files when stopping components

## [v0.2.3] - 2025-03-30

* Update health check command in localenv_start.go to use default admin credentials for OpenSearch. (99d4744)
* Add localenv.yaml to .gitignore to exclude local environment configuration files from version control. (70efde8)
* Update CHANGELOG.md, remove localenv.yaml, and refine README.md documentation. Adjust environment variable prefix in code and documentation to DEVHELPER_. (640911f)
* Add OpenSearch and OpenSearch Dashboard support to local environment, including configuration management, health checks, and improved logging. Replace Docker with Podman for container management and update related commands and status checks. (6735981)
* Add commands to start and stop local development environment components including Dapr, Temporal, and OpenSearch. Implement configuration loading, process management, and logging for improved user experience. (1ea78d8)
* Add OpenSearch support to local environment (18b0e83)
* Update CHANGELOG.md for v0.2.2 (ba01dcd)

## [v0.2.2] - 2025-03-29

* Fix tar extraction issues in Format workflow and add troubleshooting documentation (a4af48c)
* Refactor project branding from ShieldDev to Shield across all files, including updates to documentation and command descriptions. (af50eec)
* Fix environment variable prefix in README (DEVHELPER_ → DEVHELPER_) (429371a)
* Update README.md with improved documentation and latest features (f3b8717)
* Update coverage comment to trigger language statistics update (b35dfb5)
* Add .gitattributes to fix language statistics (8996383)
* Add coverage files to .gitignore and remove from repository (6b88ad4)
* Update CHANGELOG.md for v0.2.1 (3267e07)

## [v0.2.1] - 2025-03-29

* Fix CHANGELOG update to properly handle detached HEAD state (ad2f163)

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2023-03-29

### Added
- Initial release of the devhelper-cli
- Commands for managing local development environments
- Support for Dapr and Temporal
- Configuration handling
- Test framework

[Unreleased]: https://github.com/ShieldFC-RD/devhelper-cli/compare/v0.2.3...HEAD
[v0.2.3]: https://github.com/ShieldFC-RD/devhelper-cli/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/ShieldFC-RD/devhelper-cli/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/ShieldFC-RD/devhelper-cli/compare/v0.1.0...v0.2.1
[0.1.0]: https://github.com/ShieldFC-RD/devhelper-cli/releases/tag/v0.1.0 