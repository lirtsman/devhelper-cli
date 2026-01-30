# Research: MySQL 8 Integration for KindEnv

**Date**: 2026-01-30  
**Purpose**: Technical research for implementing MySQL 8 support in kindenv

## Bitnami MySQL Helm Chart Integration

### Decision: Use Bitnami MySQL Chart with Established Patterns
**Rationale**: Consistent with existing Redis implementation, well-maintained chart with comprehensive configuration options
**Alternatives considered**: Official MySQL Helm chart (less feature-rich), direct Kubernetes manifests (more complex to maintain)

### Key Configuration Parameters

#### Database Credentials
```yaml
auth:
  rootPassword: "password"           # Required for health probes
  database: "mysql"                  # Database name
  username: "mysql"                  # Custom user (optional)
  password: "password"               # Custom user password
  existingSecret: "mysql-credentials" # Use Kubernetes secrets (recommended)
```

#### Resource Configuration
```yaml
primary:
  resourcesPreset: "small"           # Quick setup option
  resources:
    requests:
      cpu: "500m"
      memory: "1Gi"
    limits:
      cpu: "1000m"
      memory: "2Gi"
  persistence:
    enabled: false                   # Disabled by default for dev
    size: "8Gi"                     # When enabled
```

#### Service Configuration
```yaml
primary:
  service:
    type: "NodePort"
    nodePorts:
      mysql: 30306                  # Configurable NodePort
```

#### Image Repository (ECR Integration)
```yaml
global:
  imageRegistry: "992979781608.dkr.ecr.eu-west-1.amazonaws.com"
image:
  repository: "bitnamilegacy/mysql"
  pullSecrets:
    - name: "ecr-credentials"
```

## Integration Patterns (Based on Existing Redis Implementation)

### Decision: Follow Established kindenv Component Pattern
**Rationale**: Maintains consistency with Redis and Temporal components, reuses proven patterns
**Alternatives considered**: Separate MySQL command (breaks consistency), direct kubectl usage (less maintainable)

### Go Implementation Pattern
```go
// Namespace creation (matches Redis pattern)
namespaceYaml, err := executeCommand("kubectl", "create", "namespace", "mysql", "--dry-run=client", "-o", "yaml")
if err != nil {
    return fmt.Errorf("failed to create MySQL namespace: %w", err)
}

// ECR credentials setup (when enabled)
if config.Images.UseAwsEcr {
    err = setupECRCreds("mysql", ecrRegistry, ecrPassword)
    if err != nil {
        return fmt.Errorf("failed to setup ECR credentials for MySQL: %w", err)
    }
}

// Helm installation with conditional ECR settings
helmArgs := []string{
    "upgrade", "--install",
    "mysql", "bitnami/mysql",
    "--namespace", "mysql",
    "--version", config.Components.MySQL.ChartVersion,
    "--set", "primary.service.type=NodePort",
    "--set", fmt.Sprintf("primary.service.nodePorts.mysql=%d", config.Components.MySQL.NodePorts.MySQL),
    "--set", fmt.Sprintf("auth.existingSecret=%s", config.Secrets.MySQL.Name),
    "--set", fmt.Sprintf("auth.database=%s", config.Components.MySQL.Database),
    "--set", fmt.Sprintf("primary.persistence.enabled=%t", config.Components.MySQL.Persistence.Enabled),
}

// ECR-specific image configuration
if config.Images.UseAwsEcr {
    helmArgs = append(helmArgs,
        "--set", fmt.Sprintf("global.imageRegistry=%s", ecrRegistry),
        "--set", "image.repository=bitnamilegacy/mysql")
}
```

### Pod Readiness Pattern
```go
// Wait for MySQL StatefulSet pod (mysql-primary-0)
podCheckCmd := exec.Command("kubectl", "get", "pod", "mysql-primary-0", "-n", "mysql", "--no-headers")

var podExists bool
for i := 0; i < 10; i++ { // 5 minutes timeout (10 * 30s)
    podOutput, err := podCheckCmd.CombinedOutput()
    if err == nil && len(podOutput) > 0 && strings.Contains(string(podOutput), "Running") {
        podExists = true
        break
    }
    if i < 9 {
        fmt.Printf("Waiting for MySQL pod to be ready (attempt %d/10)...\n", i+1)
        time.Sleep(30 * time.Second) // MySQL needs longer startup time
    }
}
```

## Configuration Structure Integration

### Decision: Extend Existing Config Structure
**Rationale**: Reuses existing secrets.mysql for credentials, adds new Components.MySQL section for deployment settings
**Alternatives considered**: Separate MySQL config section (breaks existing patterns), merge with secrets (mixes concerns)

### Configuration Schema Extension
```go
type KindEnvConfig struct {
    // ... existing fields ...
    Components struct {
        // ... existing components ...
        MySQL struct {
            Enabled      bool   `yaml:"enabled"`
            ChartVersion string `yaml:"chartVersion"`
            Database     string `yaml:"database"`
            NodePorts    struct {
                MySQL int `yaml:"mysql"`
            } `yaml:"nodePorts"`
            Resources struct {
                CPU    string `yaml:"cpu"`
                Memory string `yaml:"memory"`
            } `yaml:"resources"`
            Persistence struct {
                Enabled bool   `yaml:"enabled"`
                Size    string `yaml:"size"`
            } `yaml:"persistence"`
        } `yaml:"mysql"`
    } `yaml:"components"`
    // Reuse existing Secrets.MySQL structure
}
```

## Best Practices and Recommendations

### Security in Development Environments
- **Decision**: Use Kubernetes secrets for credentials, disable network policies
- **Rationale**: Balances security with development ease
- **Implementation**: Leverage existing secrets.mysql configuration

### Resource Sizing for Local Development
- **Decision**: Default to 500m CPU, 1Gi memory with 1000m/2Gi limits
- **Rationale**: Sufficient for development workloads without overwhelming local machines
- **Fallback**: Use Bitnami's "small" resource preset for simplicity

### Health Check Configuration
- **Decision**: Use Bitnami chart defaults with extended startup timeout
- **Rationale**: MySQL requires longer initialization time than Redis
- **Implementation**: 
  - Startup probe: 30s initial delay, 10 failure threshold
  - Readiness probe: 5s initial delay
  - Liveness probe: 30s initial delay

### Persistence Strategy
- **Decision**: Disabled by default, easily configurable
- **Rationale**: Faster startup for development, optional data retention
- **Implementation**: `primary.persistence.enabled=false` by default

## Integration Points

### Kind Cluster Port Mapping
- **Requirement**: Expose MySQL on host port 3306
- **Implementation**: Configure Kind cluster extraPortMappings in cluster configuration
- **NodePort**: Use configurable NodePort (default 30306) mapped to host 3306

### Helm Repository Management
- **Requirement**: Ensure Bitnami repository is available
- **Implementation**: Extend existing `kindenv init` command to add/update Bitnami repo
- **Pattern**: Follow existing pattern for temporalio, dapr repositories

### Status Monitoring Integration
- **Requirement**: Include MySQL status in `kindenv status` output
- **Implementation**: Extend existing status command to check MySQL pod health
- **Pattern**: Follow existing Redis status checking pattern

## Risk Mitigation

### Image Availability
- **Risk**: bitnamilegacy images may not be available in ECR
- **Mitigation**: Provide clear error messages, fallback to Docker Hub when ECR fails
- **Implementation**: Conditional image registry configuration

### Resource Constraints
- **Risk**: MySQL may fail to start on resource-constrained machines
- **Mitigation**: Configurable resource limits, clear error messages
- **Implementation**: Resource validation before installation

### Port Conflicts
- **Risk**: Port 3306 may be in use on host machine
- **Mitigation**: Configurable NodePort, clear error messages for conflicts
- **Implementation**: Port availability checking before Kind cluster creation