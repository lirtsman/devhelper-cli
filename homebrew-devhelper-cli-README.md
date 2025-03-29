# Homebrew Tap for DevHelper CLI

This repository contains the Homebrew formula for the [DevHelper CLI](https://github.com/lirtsman/devhelper-cli), a comprehensive command-line interface for ShieldDev operations.

## Installation

```bash
# Add the tap
brew tap lirtsman/devhelper-cli

# Install the CLI
brew install devhelper-cli
```

## Updating

To update to the latest version:

```bash
brew update
brew upgrade devhelper-cli
```

## Formula Details

The formula automatically downloads the appropriate binary for your operating system and architecture:
- macOS (Intel and Apple Silicon)
- Linux (x86_64 and ARM64)

## Development

### Updating the Formula

When a new version of DevHelper CLI is released, the formula needs to be updated:

1. Update the version number in the formula
2. Update the SHA256 checksums for each binary

You can generate SHA256 checksums with:

```bash
shasum -a 256 /path/to/binary
```

## License

This tap is licensed under the Apache License 2.0, matching the DevHelper CLI project. 