# Quick Start: Custom Components in KindEnv

**Version**: 1.0  
**Date**: 2026-01-30  
**Feature**: Custom Components for KindEnv

## Overview

This guide shows you how to deploy your own custom services alongside the standard kindenv components (MySQL, OpenSearch, etc.). You'll learn to configure custom containers with environment variables, connect to existing services, and expose ports for local access.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Minimal Example](#minimal-example)
3. [Environment Variables](#environment-variables)
4. [Connecting to Databases](#connecting-to-databases)
5. [Mounting Configuration Files](#mounting-configuration-files)
6. [Port Mapping](#port-mapping)
7. [Resource Limits](#resource-limits)
8. [Multiple Components](#multiple-components)
9. [Complete Example](#complete-example)
10. [Troubleshooting](#troubleshooting)

## Prerequisites

- Kindenv installed and configured
- Kind cluster running (`kindenv start` must work)
- kubectl configured and accessible
- Container images available (public or in configured registry)

**Verify setup**:

```bash
# Check kindenv is installed
devhelper-cli kindenv --help

# Check kubectl access
kubectl cluster-info

# Verify Kind cluster
kind get clusters
```

## Minimal Example

Start with the simplest possible custom component - **just name and image** (everything else is optional).

**1. Edit your `kindenv.yaml`:**

```yaml
# ... existing configuration ...

customComponents:
  - name: nginx-test    # Required
    image: nginx:latest # Required
    # That's it! Everything else is optional
```

**2. Deploy**:

```bash
devhelper-cli kindenv start
```

**3. Verify**:

```bash
# Check deployment status
devhelper-cli kindenv status

# Check pod is running
kubectl get pods -l app=nginx-test

# Expected output:
# NAME                          READY   STATUS    RESTARTS   AGE
# nginx-test-7d8f9c5b6d-abcde   1/1     Running   0          30s
```

**What happened automatically?**
- ✅ Created Kubernetes Deployment with 1 replica (default)
- ✅ Applied resource limits: 100m CPU / 128Mi memory requests, 500m CPU / 512Mi memory limits (defaults)
- ✅ Deployed to `default` namespace (default)
- ✅ Auto-generated labels: `app: nginx-test`, `managed-by: kindenv`, `component-type: custom`
- ✅ Component enabled by default (`enabled: true`)

**Only 2 fields required**: `name` and `image`. Everything else has sensible defaults or is optional!

## Environment Variables

Configure your application with environment variables.

**Direct Values** (for non-sensitive configuration):

```yaml
customComponents:
  - name: my-app
    image: myregistry/my-app:v1.0
    env:
      - name: APP_ENV
        value: "development"
      - name: LOG_LEVEL
        value: "debug"
      - name: API_URL
        value: "http://api-service:3000"
```

**What this does**:
- Sets environment variables inside the container
- Values are visible in `kubectl describe pod`
- Suitable for non-sensitive config (URLs, feature flags, etc.)

## Connecting to Databases

Use secret references to securely connect to MySQL and OpenSearch.

### MySQL Connection

```yaml
customComponents:
  - name: my-spring-app
    image: myregistry/my-spring-app:latest
    env:
      # Direct value for database URL
      - name: SPRING_DATASOURCE_URL
        value: "jdbc:mysql://mysql.mysql.svc.cluster.local:3306/mydb"
      
      # Secret reference for credentials
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

**Important Notes**:
- MySQL service DNS: `mysql.mysql.svc.cluster.local` (namespace: mysql)
- MySQL port: `3306`
- Secret `mysql-secret` is created automatically by kindenv
- Available keys: `username`, `password`

### OpenSearch Connection

```yaml
customComponents:
  - name: search-indexer
    image: myregistry/search-indexer:latest
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

**Important Notes**:
- OpenSearch service DNS: `opensearch.opensearch.svc.cluster.local` (namespace: opensearch)
- OpenSearch port: `9200`
- Protocol: `https` (even though security is disabled in dev, the endpoint uses HTTPS)
- Secret `opensearch-secret` contains credentials

### Verify Secret Availability

```bash
# Check MySQL secret
kubectl get secret mysql-secret -n mysql

# Check OpenSearch secret
kubectl get secret opensearch-secret -n opensearch

# View secret keys (values are base64 encoded)
kubectl get secret mysql-secret -n mysql -o jsonpath='{.data}'
```

## Mounting Configuration Files

Mount custom configuration files into your containers without rebuilding images.

### Basic Config File

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
          spring:
            application:
              name: my-app
          database:
            host: mysql.mysql.svc.cluster.local
            port: 3306
```

**What this does**:
- Creates a Kubernetes ConfigMap named `my-spring-app-config`
- Mounts `application.yaml` at `/config/application.yaml` inside the container
- File has read-only permissions (0644)
- Formatting and indentation are preserved exactly

**Verify the mount**:

```bash
# Check ConfigMap was created
kubectl get configmap my-spring-app-config

# Verify file contents in pod
kubectl exec -it <pod-name> -- cat /config/application.yaml
```

### Multiple Config Files

Mount multiple configuration files for complex applications:

```yaml
customComponents:
  - name: my-spring-app
    image: myregistry/my-app:latest
    configFiles:
      # Application configuration
      - name: application.yaml
        path: /config/application.yaml
        contents: |
          server:
            port: 8080
          management:
            endpoints:
              web:
                exposure:
                  include: health,info,metrics
      
      # Logging configuration
      - name: logback.xml
        path: /config/logback.xml
        contents: |
          <configuration>
            <appender name="STDOUT" class="ch.qos.logback.core.ConsoleAppender">
              <encoder>
                <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
              </encoder>
            </appender>
            <root level="INFO">
              <appender-ref ref="STDOUT" />
            </root>
          </configuration>
      
      # Database properties
      - name: database.properties
        path: /config/database.properties
        contents: |
          db.connection.pool.size=10
          db.connection.timeout=30000
          db.query.timeout=60000
```

**Access in application**:

For Spring Boot, set the config location:

```yaml
env:
  - name: SPRING_CONFIG_LOCATION
    value: "file:/config/application.yaml"
  - name: LOGGING_CONFIG
    value: "file:/config/logback.xml"
```

### Config File with Database Credentials

Combine config files with secret references:

```yaml
customComponents:
  - name: my-app
    image: myregistry/my-app:latest
    configFiles:
      - name: app-config.yaml
        path: /config/app-config.yaml
        contents: |
          database:
            host: mysql.mysql.svc.cluster.local
            port: 3306
            database: mydb
          cache:
            enabled: true
            ttl: 3600
    env:
      # Credentials from secrets
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
      # Config file location
      - name: CONFIG_FILE
        value: "/config/app-config.yaml"
```

**Application reads**:
1. Config structure from `/config/app-config.yaml`
2. Credentials from environment variables (`DB_USERNAME`, `DB_PASSWORD`)

### Updating Config Files

To update a config file:

1. **Edit kindenv.yaml** - modify the `contents` field
2. **Restart cluster**:

```bash
devhelper-cli kindenv stop
devhelper-cli kindenv start
```

**What happens**:
- ConfigMap is deleted and recreated with new contents
- Pods are restarted automatically
- New config takes effect within 30 seconds

**Verify update**:

```bash
# Check ConfigMap has new contents
kubectl get configmap my-app-config -o yaml

# Verify in running pod
kubectl exec -it <pod-name> -- cat /config/application.yaml
```

### Config File Best Practices

**File Naming**:
- Use descriptive names: `application.yaml`, not `config.yaml`
- Include extension: `.yaml`, `.xml`, `.properties`, `.json`
- Avoid special characters or spaces

**Mount Paths**:
- Use absolute paths: `/config/application.yaml`
- Avoid system directories: `/etc`, `/var`, `/usr`
- Common patterns:
  - `/config/` - application config
  - `/app/config/` - app-specific config
  - `/etc/app/` - alternative location

**File Size**:
- Keep individual files small (<100KB)
- Total size limit: 1MB per component
- Split large configs into multiple files

**Security**:
- **Never put secrets in config files**
- Use environment variables with `secretKeyRef` for credentials
- Config files are stored in ConfigMaps (not encrypted)
- Suitable for: URLs, feature flags, timeouts, cache settings
- Not suitable for: passwords, API keys, tokens

**Example - What NOT to do**:

```yaml
# ❌ BAD: Hardcoded credentials in config file
configFiles:
  - name: database.yaml
    path: /config/database.yaml
    contents: |
      username: root
      password: my-secret-password  # ← NEVER DO THIS

# ✅ GOOD: Credentials via secrets, config via file
configFiles:
  - name: database.yaml
    path: /config/database.yaml
    contents: |
      host: mysql.mysql.svc.cluster.local
      port: 3306
      pool_size: 10
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

## Port Mapping

Expose your application to your local machine.

### Basic Port Mapping

```yaml
customComponents:
  - name: api-service
    image: myregistry/api:latest
    ports:
      - containerPort: 8080
        hostPort: 8080
        protocol: TCP
```

**Access your service**:

```bash
# From your local machine
curl http://localhost:8080/health

# Or open in browser
open http://localhost:8080
```

### Multiple Ports

```yaml
customComponents:
  - name: web-app
    image: myregistry/web-app:latest
    ports:
      - containerPort: 8080
        hostPort: 8080
        protocol: TCP
      - containerPort: 9090
        hostPort: 9090
        protocol: TCP
```

**Access different services**:
- Web UI: `http://localhost:8080`
- Metrics: `http://localhost:9090/metrics`

### Custom NodePort

```yaml
customComponents:
  - name: grpc-service
    image: myregistry/grpc-service:latest
    ports:
      - containerPort: 50051
        hostPort: 50051
        protocol: TCP
        nodePort: 30051  # Explicitly specify NodePort (30000-32767)
```

**Port Assignment Rules**:
- `containerPort`: Port your app listens on (required)
- `hostPort`: Port on your local machine (defaults to containerPort)
- `nodePort`: Kubernetes NodePort (auto-assigned if not specified)
- NodePort must be in range 30000-32767
- All ports must be unique across all components

## Resource Limits

Control CPU and memory usage.

### Default Resources

If you don't specify resources, kindenv applies sensible defaults:

```yaml
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

### Custom Resources

For resource-intensive applications:

```yaml
customComponents:
  - name: data-processor
    image: myregistry/data-processor:latest
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2000m"
        memory: "4Gi"
```

**Resource Formats**:
- CPU: `"100m"` (100 millicores), `"1"` (1 core), `"2000m"` (2 cores)
- Memory: `"128Mi"` (128 mebibytes), `"1Gi"` (1 gibibyte), `"512M"` (512 megabytes)

**Best Practices**:
- Set `requests` to the minimum your app needs
- Set `limits` to prevent resource exhaustion
- Ensure `limits` >= `requests`
- Consider your laptop's total resources (don't exceed 50-70% of available)

## Multiple Components

Deploy several custom components together.

```yaml
customComponents:
  # API Service
  - name: api-service
    image: myregistry/api:v2.0
    replicas: 2  # Run 2 instances for testing load balancing
    env:
      - name: DATABASE_URL
        value: "postgresql://postgres:5432/mydb"
      - name: REDIS_URL
        value: "redis://redis:6379"
    ports:
      - containerPort: 3000
        hostPort: 3000
    resources:
      requests:
        cpu: "200m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
  
  # Background Worker (no port exposure)
  - name: worker-service
    image: myregistry/worker:v2.0
    replicas: 3
    env:
      - name: QUEUE_URL
        value: "redis://redis:6379"
      - name: API_URL
        value: "http://api-service:3000"
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
  
  # Admin Dashboard
  - name: admin-dashboard
    image: myregistry/admin-dashboard:latest
    enabled: true  # Can be set to false to temporarily disable
    ports:
      - containerPort: 8080
        hostPort: 8080
    env:
      - name: API_URL
        value: "http://api-service:3000"
```

**Service Discovery**:
- Services can communicate using DNS: `http://<service-name>:<port>`
- Example: `http://api-service:3000` (from worker to API)
- Kubernetes automatically load-balances between replicas

## Complete Example

A full-featured Spring Boot application with all options.

```yaml
customComponents:
  - name: my-spring-app
    enabled: true
    namespace: default
    image: 992979781608.dkr.ecr.eu-west-1.amazonaws.com/my-spring-app:v1.2.3
    replicas: 1
    
    # Override container command (optional)
    command: ["java"]
    args:
      - "-jar"
      - "/app/application.jar"
      - "--spring.profiles.active=local"
    
    # Environment variables
    env:
      # Application config
      - name: SPRING_PROFILES_ACTIVE
        value: "local"
      - name: SERVER_PORT
        value: "8080"
      - name: MANAGEMENT_SERVER_PORT
        value: "8081"
      
      # MySQL connection
      - name: SPRING_DATASOURCE_URL
        value: "jdbc:mysql://mysql.mysql.svc.cluster.local:3306/shieldcoredb"
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
      
      # OpenSearch connection
      - name: SEARCH_ENGINE_OPENSEARCH_CLUSTER_URIS
        value: "https://opensearch.opensearch.svc.cluster.local:9200"
      - name: SEARCH_ENGINE_OPENSEARCH_CLUSTER_USERNAME
        valueFrom:
          secretKeyRef:
            name: opensearch-secret
            key: username
      - name: SEARCH_ENGINE_OPENSEARCH_CLUSTER_PASSWORD
        valueFrom:
          secretKeyRef:
            name: opensearch-secret
            key: password
    
    # Configuration files
    configFiles:
      - name: application.yaml
        path: /config/application.yaml
        contents: |
          server:
            port: 8080
          management:
            server:
              port: 8081
            endpoints:
              web:
                exposure:
                  include: "*"
          spring:
            application:
              name: my-spring-app
      
      - name: logback.xml
        path: /config/logback.xml
        contents: |
          <configuration>
            <appender name="STDOUT" class="ch.qos.logback.core.ConsoleAppender">
              <encoder>
                <pattern>%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
              </encoder>
            </appender>
            <root level="INFO">
              <appender-ref ref="STDOUT" />
            </root>
          </configuration>
    
    # Port mappings
    ports:
      - containerPort: 8080
        hostPort: 8080
        protocol: TCP
        nodePort: 30800
      - containerPort: 8081
        hostPort: 8081
        protocol: TCP
        nodePort: 30801
    
    # Resource limits
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
    
    # Custom labels and annotations
    labels:
      tier: backend
      version: v1.2.3
      environment: dev
    annotations:
      description: "Spring Boot microservice for core business logic"
      contact: "backend-team@example.com"
```

**Deploy and verify**:

```bash
# Start kindenv
devhelper-cli kindenv start

# Check status
devhelper-cli kindenv status

# View logs
kubectl logs -l app=my-spring-app -f

# Access application
curl http://localhost:8080/actuator/health
curl http://localhost:8081/actuator/metrics

# Test database connection
curl http://localhost:8080/api/users
```

## Troubleshooting

### Component Not Deploying

**Check configuration**:

```bash
# Validate kindenv.yaml syntax
devhelper-cli kindenv start --dry-run

# Check deployment status
kubectl get deployment my-spring-app -o yaml

# Check pod status
kubectl get pods -l app=my-spring-app
kubectl describe pod <pod-name>
```

**Common Issues**:

| Problem | Symptom | Solution |
|---------|---------|----------|
| Image pull failure | `ImagePullBackOff` | Check image exists, verify registry credentials |
| Port conflict | Configuration error | Ensure hostPort is unique, change to different port |
| Missing secret | `CreateContainerConfigError` | Verify secret exists: `kubectl get secret -n <namespace>` |
| Invalid resource format | Deployment fails | Check CPU/memory format (e.g., `"500m"`, `"1Gi"`) |
| CrashLoopBackOff | Pod keeps restarting | Check logs: `kubectl logs <pod-name>` |

### Viewing Logs

```bash
# Follow logs for a component
kubectl logs -l app=my-spring-app -f

# Get logs from all replicas
kubectl logs -l app=my-spring-app --all-containers=true

# Get logs from specific pod
kubectl logs <pod-name>

# Get previous logs (if pod crashed)
kubectl logs <pod-name> --previous
```

### Port Access Issues

```bash
# Check service exists
kubectl get svc <component-name>-service

# Check NodePort assignment
kubectl get svc <component-name>-service -o jsonpath='{.spec.ports}'

# Test port from within cluster
kubectl run -it --rm debug --image=busybox --restart=Never -- wget -O- http://<component-name>:8080

# Check Kind port mappings
docker ps -a | grep kind
```

### Resetting Component

```bash
# Delete and recreate deployment
kubectl delete deployment my-spring-app
devhelper-cli kindenv start

# Force restart all pods
kubectl rollout restart deployment my-spring-app

# Scale down and up
kubectl scale deployment my-spring-app --replicas=0
kubectl scale deployment my-spring-app --replicas=1
```

### Secret Issues

```bash
# Verify secret exists
kubectl get secret mysql-secret -n mysql

# Check secret has required keys
kubectl get secret mysql-secret -n mysql -o jsonpath='{.data}'

# Create custom secret if needed
kubectl create secret generic my-app-secret \
  --from-literal=api-key=your-api-key \
  --namespace=default
```

### Resource Constraints

```bash
# Check node resources
kubectl top nodes

# Check pod resource usage
kubectl top pods

# Describe node capacity
kubectl describe node <node-name>

# If out of resources, reduce limits or increase Kind cluster resources
```

### Config File Issues

```bash
# Verify ConfigMap was created
kubectl get configmap <component-name>-config

# Check ConfigMap contents
kubectl get configmap <component-name>-config -o yaml

# Verify file is mounted in pod
kubectl exec -it <pod-name> -- ls -la /config

# Read file contents
kubectl exec -it <pod-name> -- cat /config/application.yaml

# Check pod events for mount errors
kubectl describe pod <pod-name> | grep -A 10 Events
```

**Common Config Issues**:

| Problem | Symptom | Solution |
|---------|---------|----------|
| Config not updated | Old config in pod | Restart pod: `kubectl delete pod <pod-name>` |
| File not found | FileNotFound error | Check `path` is absolute, verify ConfigMap exists |
| Invalid YAML syntax | Parsing error in app | Validate YAML: `yamllint` or online validator |
| File too large | ConfigMap create fails | Split into multiple files, keep each <100KB |
| Duplicate mount path | Only one file mounted | Ensure unique `path` for each config file |
| Permission denied | App can't read file | Files are read-only 0644 by default (should work) |

## Next Steps

1. **Explore Examples**: Check `/examples/custom-component-app/` for a complete working example
2. **Read Architecture**: Review `/specs/002-custom-components/data-model.md` for detailed configuration options
3. **Advanced Features**: Learn about health checks, volumes, and init containers (future enhancements)
4. **Production Readiness**: Implement proper health checks, resource limits, and monitoring

## Configuration Reference

Quick reference for all configuration options:

### Required Fields (Minimum Configuration)

```yaml
customComponents:
  - name: string                    # REQUIRED: DNS-compatible name (lowercase, alphanumeric, hyphens)
    image: string                   # REQUIRED: Container image [registry/]repository[:tag]
```

### Optional Fields with Defaults

```yaml
customComponents:
  - name: my-app
    image: my-image:latest
    
    # These are optional - if not specified, defaults are applied
    enabled: boolean                # Default: true
    namespace: string               # Default: "default"
    replicas: int                   # Default: 1
    resources:                      # Default: 100m/128Mi requests, 500m/512Mi limits
      requests:
        cpu: string                 # Default: "100m"
        memory: string              # Default: "128Mi"
      limits:
        cpu: string                 # Default: "500m"
        memory: string              # Default: "512Mi"
```

### Optional Fields (No Defaults - Only If Needed)

```yaml
customComponents:
  - name: my-app
    image: my-image:latest
    
    # Container configuration (optional)
    command: []string               # Override ENTRYPOINT (e.g., ["/bin/sh"])
    args: []string                  # Container args (e.g., ["-c", "java -jar app.jar"])
    
    # Environment variables (optional)
    env:
      - name: string                # Required if env specified
        value: string               # Either value OR valueFrom required
        valueFrom:                  # Alternative to value
          secretKeyRef:
            name: string            # Secret name
            key: string             # Secret key
    
    # Port mappings (optional)
    ports:
      - containerPort: int          # Required if ports specified
        hostPort: int               # Optional: defaults to containerPort
        protocol: string            # Optional: TCP or UDP (default: TCP)
        nodePort: int               # Optional: auto-assigned in range 30000-32767
    
    # Configuration files (optional)
    configFiles:
      - name: string                # Required if configFiles specified (filename only)
        path: string                # Required if configFiles specified (absolute path)
        contents: string            # Required if configFiles specified (inline YAML/JSON/XML)
    
    # Metadata (optional)
    labels:                         # Optional: adds to auto-generated labels
      key: value
    annotations:                    # Optional
      key: value
```

### Field Requirements Summary

**REQUIRED** (must always provide):
- `name` - Component identifier
- `image` - Container image

**OPTIONAL WITH DEFAULTS** (can omit, sensible defaults applied):
- `enabled` → true
- `namespace` → "default"
- `replicas` → 1
- `resources` → 100m/128Mi requests, 500m/512Mi limits

**OPTIONAL** (only specify if needed):
- `command`, `args` - Custom entrypoint
- `env` - Environment variables
- `ports` - External port access
- `configFiles` - Configuration file mounting
- `labels`, `annotations` - Custom metadata

## FAQ

**Q: Can I use private container registries?**  
A: Yes, if using AWS ECR or Harbor (configured in kindenv.yaml), pull secrets are automatically created. For other registries, create pull secrets manually.

**Q: How do I expose a service on port 80 or 443?**  
A: Ports < 1024 require root privileges. Use higher ports (e.g., 8080) and access via localhost:8080.

**Q: Can I mount configuration files?**  
A: Yes! Use the `configFiles` array to mount custom config files. They're automatically converted to ConfigMaps and mounted at your specified paths.

**Q: Can I use external config files instead of inline contents?**  
A: Not in the current implementation. All config files must be specified inline in kindenv.yaml. This keeps everything self-contained in one file.

**Q: How do I update config files without rebuilding the container?**  
A: Edit the `contents` field in kindenv.yaml and restart the cluster with `kindenv stop && kindenv start`. The ConfigMap will be updated and pods restarted automatically.

**Q: Can I mount volumes or persistent storage?**  
A: Not in the MVP release beyond config files. Persistent volume support is planned for a future enhancement.

**Q: How do I update a running component?**  
A: Edit kindenv.yaml and run `devhelper-cli kindenv start` again. The deployment will be updated automatically.

**Q: Can components in different namespaces communicate?**  
A: Yes, use fully qualified DNS names: `http://<service>.<namespace>.svc.cluster.local:<port>`

**Q: What happens when I run `kindenv stop`?**  
A: All custom component deployments and services are deleted. Namespaces and secrets are preserved.

---

**Need Help?** 
- Check [Troubleshooting](#troubleshooting) section
- Review complete examples in `/examples/`
- Read detailed documentation in `/specs/002-custom-components/`
