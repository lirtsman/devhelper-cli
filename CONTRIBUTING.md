# Contributing to devhelper-cli

Thank you for your interest in contributing to devhelper-cli! This document outlines the process for contributing to the project, including branching strategy, versioning approach, and contribution workflow.

## Branching Strategy

We follow a modified GitFlow branching strategy:

1. **`main`** - The production branch. All releases are merged into this branch and tagged with a version number.
2. **`develop`** - The primary integration branch where all features are merged before being released.
3. **Feature branches** - Created from `develop` for each new feature or bugfix, following the naming convention:
   - `feature/short-description` for new features
   - `bugfix/short-description` for bug fixes
4. **Release branches** - Created from `develop` when preparing a new release, named as `release/vX.Y.Z`
5. **Hotfix branches** - Created from `main` for critical fixes that need to be applied to production immediately, named as `hotfix/vX.Y.Z`

### Branch Workflow

1. Create a feature/bugfix branch from `develop`:
   ```bash
   git checkout develop
   git pull
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, commit, and push to your branch:
   ```bash
   git add .
   git commit -m "feat: your descriptive commit message"
   git push -u origin feature/your-feature-name
   ```

3. Create a Pull Request to merge into `develop`.

4. After code review and approval, your changes will be merged into `develop`.

## Versioning

We follow [Semantic Versioning](https://semver.org/) (SemVer) for version numbers:

- **MAJOR** version when you make incompatible API changes (X.y.z)
- **MINOR** version when you add functionality in a backward compatible manner (x.Y.z)
- **PATCH** version when you make backward compatible bug fixes (x.y.Z)

Additional labels for pre-release and build metadata are available as extensions to the MAJOR.MINOR.PATCH format.

### Version Tags

When creating releases, we tag the repository with the version number:

```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

Pre-release versions should be tagged with suffixes like `-alpha`, `-beta`, or `-rc`:

```bash
git tag -a v1.0.0-beta.1 -m "Beta release 1.0.0-beta.1"
git push origin v1.0.0-beta.1
```

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types include:
- **feat** - A new feature
- **fix** - A bug fix
- **docs** - Documentation only changes
- **style** - Changes that do not affect the meaning of the code (white-space, formatting, etc)
- **refactor** - A code change that neither fixes a bug nor adds a feature
- **perf** - A code change that improves performance
- **test** - Adding missing tests or correcting existing tests
- **chore** - Changes to the build process or auxiliary tools

Example:
```
feat(localenv): add support for custom Dapr configuration

This adds the ability to specify custom Dapr configurations in localenv.yaml.
It allows users to override default settings for individual components.

Closes #123
```

## Pull Requests

When creating a Pull Request, please:

1. Use a clear and descriptive title
2. Include a detailed description of changes
3. Reference any related issues using GitHub's keywords (`fixes`, `closes`, etc.)
4. Ensure all CI checks pass
5. Request a review from at least one maintainer

## Setting Up Development Environment

1. Clone the repository:
   ```bash
   git clone https://github.com/lirtsman/devhelper-cli.git
   cd devhelper-cli
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Run tests to ensure everything works:
   ```bash
   make test
   ```

4. Build the CLI:
   ```bash
   make build
   ```

## Testing

All code contributions should include appropriate tests:

1. Unit tests for new functionality
2. Updates to existing tests for modified functionality

Run tests with:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage-html
```

## Release Process

1. Create a release branch from `develop`:
   ```bash
   git checkout develop
   git pull
   git checkout -b release/vX.Y.Z
   ```

2. Update version information in code and documentation
3. Update CHANGELOG.md with details of changes
4. Create a Pull Request from the release branch to `main`
5. After approval and merge, tag the release:
   ```bash
   git checkout main
   git pull
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```
6. The GitHub Actions workflow will automatically create the release and build artifacts.

## License

By contributing to devhelper-cli, you agree that your contributions will be licensed under the same license as the project.

Thank you for your contributions! 