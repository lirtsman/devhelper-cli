# Custom Components for KindEnv

Deploy your own services alongside infrastructure components (MySQL, OpenSearch, etc.) in your Kind-based development environment.

## Quick Start

Add a custom component to your `kindenv.yaml`:

```yaml
customComponents:
  - name: my-app
    image: nginx:latest
```

Then deploy:

```bash
devhelper-cli kindenv start
```

That's it! Your component is now running in the cluster.

## Features

- **Container Images**: Deploy any container image (public or private registries)
- **Environment Variables**: Configure your app with direct values or secret references
- **Port Mapping**: Expose services to your local machine via NodePort
- **Resource Limits**: Control CPU and memory usage
- **Configuration Files**: Mount config files without rebuilding images
- **Multi-Component**: Deploy multiple services together
- **Service Discovery**: Components can communicate using Kubernetes DNS

## Configuration

### Minimal Configuration

Only two fields are required:

```yaml
customComponents:
  - name: my-app        # Required: DNS-compatible name
    image: nginx:latest  # Required: Container image
```

Everything else has sensible defaults:
- **Namespace**: `default`
- **Replicas**: `1`
- **Enabled**: `true`
- **Resources**: 100m CPU / 128Mi memory requests, 500m CPU / 512Mi memory limits

### Environment Variables

**Direct values** (for non-sensitive config):

```yaml
customComponents:
  - name: my-app
    image: myregistry/my-app:v1.0
    env:
      - name: APP_ENV
        value: "development"
      - name: LOG_LEVEL
        value: "debug"
```

**Secret references** (for credentials):

```yaml
customComponents:
  - name: my-app
    image: myregistry/my-app:v1.0
    env:
      - name: DB_USERNAME
        valueFrom:
          secretKeyRef:
            name: mysql-secret
            key: username
      - name: DB_PASSWORD
        valueFrom:
          secretKeyRef:
            name: mysql-secret
            key: password
```

### Port Mapping

Expose your service to localhost:

```yaml
customComponents:
  - name: api-service
    image: myregistry/api:latest
    ports:
      - containerPort: 8080
        protocol: TCP
        nodePort: 30088
```

Access your service at `http://localhost:30088`.

### Resource Limits

Control CPU and memory:

```yaml
customComponents:
  - name: data-processor
    image: myregistry/processor:latest
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2000m"
        memory: "4Gi"
```

### Configuration Files

Mount config files without rebuilding images:

```yaml
customComponents:
  - name: my-spring-app
    image: myregistry/my-app:latest
    configFiles:
      - name: application.yaml
        path: /config/application.yaml
        contents: |
          server:
            port: 8080
          database:
            host: mysql.mysql.svc.cluster.local
```

### Complete Example

A full-featured Spring Boot application:

```yaml
customComponents:
  - name: my-spring-app
    image: myregistry/my-app:v1.0
    namespace: default
    replicas: 1
    
    # Override container command
    command: ["java"]
    args: ["-jar", "/app/application.jar"]
    
    # Environment variables
    env:
      - name: SPRING_PROFILES_ACTIVE
        value: "local"
      - name: SPRING_DATASOURCE_URL
        value: "jdbc:mysql://mysql.mysql.svc.cluster.local:3306/mydb"
      - name: SPRING_DATASOURCE_USERNAME
        valueFrom:
          secretKeyRef:
            name: mysql-secret
            key: username
      - name: SPRING_DATASOURCE_PASSWORD
        valueFrom:
          secretKeyRef:
            name: mysql-secret
            key: password
    
    # Configuration files
    configFiles:
      - name: application.yaml
        path: /config/application.yaml
        contents: |
          server:
            port: 8080
          spring:
            application:
              name: my-spring-app
    
    # Port mappings
    ports:
      - containerPort: 8080
        protocol: TCP
        nodePort: 30088
    
    # Resource limits
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
```

## Connecting to Infrastructure

### MySQL

```yaml
env:
  - name: SPRING_DATASOURCE_URL
    value: "jdbc:mysql://mysql.mysql.svc.cluster.local:3306/mydb"
  - name: SPRING_DATASOURCE_USERNAME
    valueFrom:
      secretKeyRef:
        name: mysql-secret
        key: username
  - name: SPRING_DATASOURCE_PASSWORD
    valueFrom:
      secretKeyRef:
        name: mysql-secret
        key: password
```

### OpenSearch

```yaml
env:
  - name: OPENSEARCH_URL
    value: "https://opensearch.opensearch.svc.cluster.local:9200"
  - name: OPENSEARCH_USERNAME
    valueFrom:
      secretKeyRef:
        name: opensearch-secret
        key: username
  - name: OPENSEARCH_PASSWORD
    valueFrom:
      secretKeyRef:
        name: opensearch-secret
        key: password
```

### Monitoring Custom Application Metrics

When the monitoring stack is enabled (`components.monitoring.enabled: true`), your custom components can expose metrics for automatic scraping by Prometheus.

**Requirements**: Your component must serve a [Prometheus-compatible](https://prometheus.io/docs/instrumenting/exposition_formats/) `/metrics` endpoint.

**Create a ServiceMonitor** to tell Prometheus where to scrape:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-app-metrics
  namespace: default
  labels:
    release: monitoring           # Must match the Helm release label selector
spec:
  selector:
    matchLabels:
      app: my-app                 # Must match your Service's labels
  endpoints:
    - port: http                  # Named port on your Service
      path: /metrics
      interval: 15s
```

```bash
kubectl apply -f my-app-servicemonitor.yaml
```

Within 1–2 minutes, metrics appear in Grafana's **Explore** view (data source: Prometheus).

> **Note**: The monitoring stack auto-discovers ServiceMonitors across all namespaces (`serviceMonitorNamespaceSelector: {}`), so no additional configuration is needed beyond creating the ServiceMonitor with the `release: monitoring` label.

## Managing Components

### Check Status

```bash
# View all component statuses
devhelper-cli kindenv status

# View detailed status
devhelper-cli kindenv status --verbose

# Check with kubectl
kubectl get deployments
kubectl get pods -l component-type=custom
```

### View Logs

```bash
# Follow logs for a component
kubectl logs -l app=my-app -f

# View logs from all replicas
kubectl logs -l app=my-app --all-containers=true
```

### Update Configuration

1. Edit `kindenv.yaml`
2. Restart the cluster:

```bash
devhelper-cli kindenv stop
devhelper-cli kindenv start
```

### Disable a Component

Set `enabled: false`:

```yaml
customComponents:
  - name: my-app
    image: nginx:latest
    enabled: false  # Component will be skipped during deployment
```

## Troubleshooting

### Component Not Starting

```bash
# Check deployment status
kubectl get deployment my-app

# Check pod status
kubectl get pods -l app=my-app

# View pod events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>
```

### Common Issues

| Problem | Solution |
|---------|----------|
| Image pull failure | Verify image exists and registry credentials are configured |
| Port conflict | Ensure NodePort is unique (30000-32767) |
| Missing secret | Verify secret exists: `kubectl get secret mysql-secret -n mysql` |
| Invalid resource format | Use Kubernetes format: `"500m"` for CPU, `"1Gi"` for memory |
| Config file not found | Verify mount path is absolute and ConfigMap exists |

### Verify Config Files

```bash
# Check ConfigMap was created
kubectl get configmap my-app-config

# View file contents in pod
kubectl exec -it <pod-name> -- cat /config/application.yaml
```

## Best Practices

1. **Use Secret References**: Never put credentials in config files or direct env values
2. **Resource Limits**: Set appropriate limits to prevent resource exhaustion
3. **Unique Ports**: Ensure NodePorts are unique across all components
4. **Config File Size**: Keep config files under 100KB each
5. **Service Discovery**: Use Kubernetes DNS for inter-service communication
6. **Namespace Organization**: Use namespaces to organize related components

## Configuration Reference

### Required Fields

- `name` (string): DNS-compatible component name
- `image` (string): Container image reference

### Optional Fields

- `enabled` (boolean): Enable/disable component (default: `true`)
- `namespace` (string): Kubernetes namespace (default: `"default"`)
- `replicas` (int): Number of replicas (default: `1`)
- `command` ([]string): Override container entrypoint
- `args` ([]string): Container arguments
- `env` ([]EnvVar): Environment variables
- `ports` ([]PortMapping): Port mappings
- `resources` (ResourceRequirements): CPU and memory limits
- `configFiles` ([]ConfigFile): Configuration files to mount
- `labels` (map[string]string): Custom labels
- `annotations` (map[string]string): Custom annotations

For detailed configuration options, see the [Quick Start Guide](specs/002-custom-components/quickstart.md).

## Examples

See the [examples directory](examples/) for complete working examples.

## Related Documentation

- [KindEnv Setup Guide](KINDENV.md) - Setting up the Kind environment
- [Quick Start Guide](specs/002-custom-components/quickstart.md) - Detailed examples and troubleshooting
- [Development Guide](DEVELOPMENT.md) - Contributing to the project
