# Changelog

All notable changes to this project will be documented in this file.

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

## [Unreleased]

### Added
- GitHub Actions workflow for CI/CD
- Test coverage reporting
- Automated releases
- Code formatting

## [0.1.0] - 2023-03-29

### Added
- Initial release of the devhelper-cli
- Commands for managing local development environments
- Support for Dapr and Temporal
- Configuration handling
- Test framework

[Unreleased]: https://github.com/lirtsman/devhelper-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/lirtsman/devhelper-cli/releases/tag/v0.1.0 