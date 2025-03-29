# Homebrew Tap for DevHelper CLI

This repository contains the Homebrew formula for the [DevHelper CLI](https://github.com/lirtsman/devhelper-cli), a comprehensive command-line interface for ShieldDev operations.

## Installation

```bash
# Add the tap
brew tap lirtsman/devhelper-cli

# Install the CLI
brew install devhelper-cli

# Verify installation
devhelper-cli version
```

## Updating

To update to the latest version:

```bash
brew update
brew upgrade devhelper-cli
```

## Formula Details

The formula builds DevHelper CLI from source using Go. This ensures compatibility with all platforms supported by Homebrew.

## Development

### Updating the Formula

When a new version of DevHelper CLI is released:

1. Update the version number and URL in the formula
2. Update the SHA256 checksum:
   ```bash
   curl -sL https://github.com/lirtsman/devhelper-cli/archive/refs/tags/vX.Y.Z.tar.gz | shasum -a 256
   ```
3. Test the formula locally:
   ```bash
   brew uninstall devhelper-cli
   brew install --build-from-source ./Formula/devhelper-cli.rb
   ```
4. Commit and push the changes

## License

This tap is licensed under the Apache License 2.0, matching the DevHelper CLI project. 