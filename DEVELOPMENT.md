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

## Custom Components Development

When developing or testing custom components functionality, you can use the following examples:

### Minimal Example

```yaml
customComponents:
  - name: test-app
    image: nginx:latest
```

### With Environment Variables

```yaml
customComponents:
  - name: test-app
    image: nginx:latest
    env:
      - name: TEST_VAR
        value: "test-value"
```

### With Port Mapping

```yaml
customComponents:
  - name: test-app
    image: nginx:latest
    ports:
      - containerPort: 80
        protocol: TCP
        nodePort: 30088
```

### With Config Files

```yaml
customComponents:
  - name: test-app
    image: nginx:latest
    configFiles:
      - name: config.yaml
        path: /config/config.yaml
        contents: |
          key: value
```

### Complete Example

```yaml
customComponents:
  - name: test-app
    image: nginx:latest
    namespace: default
    replicas: 1
    env:
      - name: APP_ENV
        value: "development"
      - name: DB_HOST
        valueFrom:
          secretKeyRef:
            name: mysql-secret
            key: host
    ports:
      - containerPort: 80
        protocol: TCP
        nodePort: 30088
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
    configFiles:
      - name: application.yaml
        path: /config/application.yaml
        contents: |
          server:
            port: 8080
```

### Testing Custom Components

1. Add a custom component to your `kindenv.yaml`
2. Start the environment: `devhelper-cli kindenv start`
3. Check status: `devhelper-cli kindenv status`
4. Verify deployment: `kubectl get pods -l component-type=custom`
5. View logs: `kubectl logs -l app=test-app`

For more examples and detailed documentation, see [CUSTOM_COMPONENTS.md](CUSTOM_COMPONENTS.md) and [specs/002-custom-components/quickstart.md](specs/002-custom-components/quickstart.md). 