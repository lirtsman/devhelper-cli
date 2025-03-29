# Development Notes

## Go Version Compatibility

This project uses Go 1.21 in order to maintain compatibility with development tools and older environments. The go.mod file specifies this version explicitly.

## Known Issues

### golangci-lint Compatibility

There is currently a compatibility issue between Go 1.24 and golangci-lint. If you're developing with Go 1.24+ locally, you may encounter errors when running `golangci-lint`.

Error message:
```
ERRO Running error: can't run linter goanalysis_metalinter
buildir: failed to load package goarch: could not load export data: internal error in importing "internal/goarch" (unsupported version: 2); please report an issue
```

As a workaround, in the Makefile we've temporarily disabled the `golangci-lint` step and are only running `go vet`. Once golangci-lint releases a version compatible with Go 1.24+, we can re-enable it.

## Building Locally

To build the project locally:

```bash
make build
```

## Running Tests

To run all tests:

```bash
make test
```

## Linting

Currently, linting is limited to running `go vet`:

```bash
make lint
```

## YAML Package Usage

Throughout the codebase, we use `gopkg.in/yaml.v3` imported with the alias `yamlv3` to avoid potential namespace conflicts. 