# Data Model: MySQL 8 Support for KindEnv

**Date**: 2026-01-30  
**Purpose**: Define data structures and entities for MySQL integration

## Configuration Entities

### MySQL Component Configuration

**Entity**: `MySQLConfig`  
**Purpose**: Configuration settings for MySQL deployment within kindenv  
**Location**: `internal/kindenv/config.go`

```go
type MySQLConfig struct {
    Enabled      bool              `yaml:"enabled"`
    ChartVersion string            `yaml:"chartVersion"`
    Database     string            `yaml:"database"`
    NodePorts    MySQLNodePorts    `yaml:"nodePorts"`
    Resources    MySQLResources    `yaml:"resources"`
    Persistence  MySQLPersistence  `yaml:"persistence"`
}

type MySQLNodePorts struct {
    MySQL int `yaml:"mysql"`
}

type MySQLResources struct {
    CPU    string `yaml:"cpu"`
    Memory string `yaml:"memory"`
}

type MySQLPersistence struct {
    Enabled bool   `yaml:"enabled"`
    Size    string `yaml:"size"`
}
```

**Validation Rules**:
- `Enabled`: Boolean flag, defaults to `false`
- `ChartVersion`: Must be valid semantic version (e.g., "9.4.6")
- `Database`: Must be valid MySQL database name (alphanumeric, underscores, 1-64 chars)
- `NodePorts.MySQL`: Must be in range 30000-32767 (Kubernetes NodePort range)
- `Resources.CPU`: Must be valid Kubernetes CPU format (e.g., "500m", "1")
- `Resources.Memory`: Must be valid Kubernetes memory format (e.g., "1Gi", "512Mi")
- `Persistence.Size`: Must be valid Kubernetes storage format (e.g., "8Gi", "10Gi")

**Relationships**:
- Integrates with existing `KindEnvConfig.Components` structure
- References existing `KindEnvConfig.Secrets.MySQL` for credentials
- Inherits ECR configuration from `KindEnvConfig.Images`

**State Transitions**:
1. `Disabled` → `Enabled`: Triggers MySQL installation on next `kindenv start`
2. `Enabled` → `Configuration Changed`: Triggers MySQL reconfiguration on restart
3. `Enabled` → `Disabled`: Triggers MySQL cleanup on next `kindenv start`

### MySQL Secret Configuration (Existing)

**Entity**: `MySQLSecretConfig`  
**Purpose**: Credential management for MySQL access  
**Location**: `internal/kindenv/config.go` (existing)

```go
type MySQLSecretConfig struct {
    Enabled   bool   `yaml:"enabled"`
    Name      string `yaml:"name"`
    Namespace string `yaml:"namespace"`
    Username  string `yaml:"username"`
    Password  string `yaml:"password"`
}
```

**Validation Rules**:
- `Name`: Must be valid Kubernetes secret name (DNS-1123 subdomain)
- `Namespace`: Must be valid Kubernetes namespace name
- `Username`: Must be valid MySQL username (1-32 chars, alphanumeric + underscore)
- `Password`: Must be non-empty string (recommend 8+ characters)

**Relationships**:
- Referenced by `MySQLConfig` for credential management
- Creates Kubernetes Secret resource in specified namespace
- Used by Bitnami MySQL Helm chart via `auth.existingSecret`

## Runtime Entities

### MySQL Deployment State

**Entity**: `MySQLDeploymentState`  
**Purpose**: Runtime state tracking for MySQL deployment  
**Location**: In-memory state during kindenv operations

```go
type MySQLDeploymentState struct {
    Installed     bool
    Namespace     string
    PodName       string
    ServiceName   string
    SecretName    string
    Status        MySQLStatus
    ConnectionInfo MySQLConnectionInfo
}

type MySQLStatus string

const (
    MySQLStatusPending    MySQLStatus = "Pending"
    MySQLStatusInstalling MySQLStatus = "Installing"
    MySQLStatusRunning    MySQLStatus = "Running"
    MySQLStatusFailed     MySQLStatus = "Failed"
    MySQLStatusStopped    MySQLStatus = "Stopped"
)

type MySQLConnectionInfo struct {
    Host     string
    Port     int
    Database string
    Username string
}
```

**State Transitions**:
1. `Pending` → `Installing`: Helm chart installation started
2. `Installing` → `Running`: Pod ready and health checks passing
3. `Installing` → `Failed`: Installation timeout or error
4. `Running` → `Stopped`: kindenv stop or MySQL pod failure
5. `Failed` → `Installing`: Retry installation attempt

### MySQL Health Check

**Entity**: `MySQLHealthCheck`  
**Purpose**: Health monitoring data for MySQL instance  
**Location**: Runtime data for status reporting

```go
type MySQLHealthCheck struct {
    Timestamp    time.Time
    PodReady     bool
    ServiceReady bool
    DatabaseReady bool
    ErrorMessage string
    Uptime       time.Duration
}
```

**Validation Rules**:
- `Timestamp`: Must be valid time
- `PodReady`: Kubernetes pod readiness status
- `ServiceReady`: Kubernetes service availability
- `DatabaseReady`: MySQL connection test result
- `ErrorMessage`: Detailed error information when checks fail

## Kubernetes Resource Entities

### MySQL Kubernetes Resources

**Entity**: Kubernetes resources created by MySQL deployment  
**Purpose**: Track Kubernetes objects for cleanup and management

```yaml
# Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: mysql

# Secret (from existing secrets.mysql config)
apiVersion: v1
kind: Secret
metadata:
  name: mysql-credentials
  namespace: mysql
type: Opaque
stringData:
  username: "root"
  password: "password"

# Helm Release (managed by Helm)
# Creates: StatefulSet, Service, ConfigMap, etc.
```

**Resource Relationships**:
- `Namespace`: Contains all MySQL-related resources
- `Secret`: Referenced by StatefulSet for authentication
- `StatefulSet`: Created by Helm chart, manages MySQL pod
- `Service`: Exposes MySQL with NodePort configuration
- `PersistentVolumeClaim`: Created when persistence enabled

## Configuration Defaults

### Default Values

```go
func NewDefaultMySQLConfig() MySQLConfig {
    return MySQLConfig{
        Enabled:      false,
        ChartVersion: "9.4.6",
        Database:     "mysql",
        NodePorts: MySQLNodePorts{
            MySQL: 30306,
        },
        Resources: MySQLResources{
            CPU:    "500m",
            Memory: "1Gi",
        },
        Persistence: MySQLPersistence{
            Enabled: false,
            Size:    "8Gi",
        },
    }
}
```

### Integration with Existing Config

```go
// Extension to existing KindEnvConfig
type KindEnvConfig struct {
    // ... existing fields ...
    Components struct {
        Temporal struct { /* existing */ } `yaml:"temporal"`
        Redis    struct { /* existing */ } `yaml:"redis"`
        MySQL    MySQLConfig              `yaml:"mysql"`  // NEW
        // ... other components ...
    } `yaml:"components"`
    
    Secrets struct {
        MySQL MySQLSecretConfig `yaml:"mysql"`  // EXISTING - reused
    } `yaml:"secrets"`
}
```

## Data Flow

### Configuration Loading Flow
1. Load `kindenv.yaml` configuration file
2. Parse `components.mysql` section into `MySQLConfig`
3. Parse `secrets.mysql` section into `MySQLSecretConfig`
4. Validate configuration parameters
5. Apply defaults for missing values

### Installation Flow
1. Check if MySQL component is enabled
2. Create MySQL namespace if not exists
3. Create Kubernetes secret from `secrets.mysql` configuration
4. Install Bitnami MySQL Helm chart with configuration parameters
5. Wait for MySQL pod to be ready
6. Update deployment state to `Running`

### Status Reporting Flow
1. Query Kubernetes API for MySQL pod status
2. Test MySQL database connection
3. Aggregate health check results
4. Format status output for `kindenv status` command

### Cleanup Flow
1. Identify MySQL resources in cluster
2. Remove Helm release
3. Delete Kubernetes secret
4. Delete namespace (if empty)
5. Update deployment state to `Stopped`