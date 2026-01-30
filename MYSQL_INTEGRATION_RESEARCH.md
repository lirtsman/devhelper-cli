# Bitnami MySQL Helm Chart Integration Research

This document provides comprehensive research on integrating the Bitnami MySQL Helm chart into the kindenv command, based on existing Redis implementation patterns.

## 1. Chart Configuration Parameters

### 1.1 Database Credentials Configuration

**Key Parameters:**
- `auth.rootPassword` - Root user password (required for probes to work)
- `auth.database` - Custom database name to create (default: `my_database`)
- `auth.username` - Custom database username
- `auth.password` - Password for the custom user
- `auth.createDatabase` - Whether to create the database (default: `true`)
- `auth.existingSecret` - Use existing Kubernetes secret (contains keys: `mysql-root-password`, `mysql-password`)

**Example Helm Values:**
```yaml
auth:
  rootPassword: "password"
  database: "mysql"
  username: "root"
  password: "password"
  createDatabase: true
```

**Helm Command Equivalent:**
```bash
--set auth.rootPassword=password \
--set auth.database=mysql \
--set auth.username=root \
--set auth.password=password \
--set auth.createDatabase=true
```

### 1.2 Resource Limits Configuration

**Key Parameters:**
- `primary.resourcesPreset` - Preset values: `none`, `nano`, `micro`, `small`, `medium`, `large`, `xlarge`, `2xlarge` (default: `small`)
- `primary.resources.requests.cpu` - CPU request (e.g., `500m`)
- `primary.resources.requests.memory` - Memory request (e.g., `1Gi`)
- `primary.resources.limits.cpu` - CPU limit (e.g., `1000m`)
- `primary.resources.limits.memory` - Memory limit (e.g., `2Gi`)

**Example Helm Values:**
```yaml
primary:
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1000m
      memory: 2Gi
```

**Helm Command Equivalent:**
```bash
--set primary.resources.requests.cpu=500m \
--set primary.resources.requests.memory=1Gi \
--set primary.resources.limits.cpu=1000m \
--set primary.resources.limits.memory=2Gi
```

### 1.3 Persistence Configuration

**Key Parameters:**
- `primary.persistence.enabled` - Enable persistent volumes (default: `true`)
- `primary.persistence.size` - Storage size (default: `8Gi`)
- `primary.persistence.storageClass` - Storage class name (empty = default)
- `primary.persistence.accessModes` - Access modes (default: `["ReadWriteOnce"]`)

**Example Helm Values:**
```yaml
primary:
  persistence:
    enabled: false  # Disable for faster startup in dev
    size: 8Gi
    storageClass: ""
    accessModes:
      - ReadWriteOnce
```

**Helm Command Equivalent:**
```bash
--set primary.persistence.enabled=false \
--set primary.persistence.size=8Gi
```

### 1.4 NodePort Service Configuration

**Key Parameters:**
- `primary.service.type` - Service type: `ClusterIP`, `NodePort`, `LoadBalancer` (default: `ClusterIP`)
- `primary.service.ports.mysql` - Service port (default: `3306`)
- `primary.service.nodePorts.mysql` - NodePort number (empty = auto-assigned)

**Example Helm Values:**
```yaml
primary:
  service:
    type: NodePort
    ports:
      mysql: 3306
    nodePorts:
      mysql: 30306  # Specific NodePort
```

**Helm Command Equivalent:**
```bash
--set primary.service.type=NodePort \
--set primary.service.ports.mysql=3306 \
--set primary.service.nodePorts.mysql=30306
```

### 1.5 Image Repository Configuration (bitnamilegacy)

**Key Parameters:**
- `image.registry` - Image registry (default: `docker.io`)
- `image.repository` - Image repository (default: `bitnami/mysql`)
- `global.imageRegistry` - Global registry override
- `image.pullSecrets` - Array of pull secret names

**Example Helm Values:**
```yaml
image:
  registry: docker.io
  repository: bitnamilegacy/mysql
  tag: 8.0.35-debian-11-r0
  pullPolicy: IfNotPresent
```

**Helm Command Equivalent:**
```bash
--set image.repository=bitnamilegacy/mysql \
--set image.tag=8.0.35-debian-11-r0
```

**For ECR Integration:**
```bash
--set global.imageRegistry=992979781608.dkr.ecr.eu-west-1.amazonaws.com \
--set image.repository=bitnamilegacy/mysql \
--set image.pullSecrets[0].name=ecr-credentials
```

### 1.6 Health Check Configuration

**Key Parameters:**
- `primary.livenessProbe.enabled` - Enable liveness probe (default: `true`)
- `primary.livenessProbe.initialDelaySeconds` - Initial delay (default: `5`)
- `primary.readinessProbe.enabled` - Enable readiness probe (default: `true`)
- `primary.readinessProbe.initialDelaySeconds` - Initial delay (default: `5`)
- `primary.startupProbe.enabled` - Enable startup probe (default: `true`)
- `primary.startupProbe.initialDelaySeconds` - Initial delay (default: `15`)

**Example Helm Values:**
```yaml
primary:
  livenessProbe:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 1
    failureThreshold: 3
  readinessProbe:
    enabled: true
    initialDelaySeconds: 5
    periodSeconds: 10
    timeoutSeconds: 1
    failureThreshold: 3
  startupProbe:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 1
    failureThreshold: 10
```

## 2. Integration Patterns (Based on Redis Implementation)

### 2.1 Helm Chart Installation Pattern

**Redis Pattern (from `kindenv_start.go` lines 696-708):**
```go
_, err = executeCommand("helm", "upgrade", "--install",
    "redis", "bitnami/redis",
    "--namespace", "redis",
    "--version", config.Components.Redis.ChartVersion,
    "--set", "master.service.type=NodePort",
    "--set", fmt.Sprintf("master.service.nodePorts.redis=%d", config.Components.Redis.NodePorts.Redis),
    "--set", fmt.Sprintf("auth.enabled=%t", config.Components.Redis.Auth.Enabled),
    "--set", "replica.replicaCount=0",
    "--set", "image.repository=bitnamilegacy/redis")
```

**MySQL Equivalent Pattern:**
```go
helmArgs := []string{
    "upgrade", "--install",
    "mysql", "bitnami/mysql",
    "--namespace", "mysql",
    "--version", config.Components.MySQL.ChartVersion,
    "--set", "primary.service.type=NodePort",
    "--set", fmt.Sprintf("primary.service.nodePorts.mysql=%d", config.Components.MySQL.NodePorts.MySQL),
    "--set", fmt.Sprintf("auth.rootPassword=%s", config.Secrets.MySQL.Password),
    "--set", fmt.Sprintf("auth.database=%s", config.Components.MySQL.Database),
    "--set", fmt.Sprintf("auth.username=%s", config.Secrets.MySQL.Username),
    "--set", fmt.Sprintf("auth.password=%s", config.Secrets.MySQL.Password),
    "--set", fmt.Sprintf("primary.persistence.enabled=%t", config.Components.MySQL.Persistence.Enabled),
    "--set", fmt.Sprintf("primary.resources.requests.cpu=%s", config.Components.MySQL.Resources.Requests.CPU),
    "--set", fmt.Sprintf("primary.resources.requests.memory=%s", config.Components.MySQL.Resources.Requests.Memory),
    "--set", fmt.Sprintf("primary.resources.limits.cpu=%s", config.Components.MySQL.Resources.Limits.CPU),
    "--set", fmt.Sprintf("primary.resources.limits.memory=%s", config.Components.MySQL.Resources.Limits.Memory),
    "--set", "image.repository=bitnamilegacy/mysql",
}

// Add ECR image registry override if enabled
if config.Images.UseAwsEcr {
    helmArgs = append(helmArgs,
        "--set", fmt.Sprintf("global.imageRegistry=%s", ecrRegistry),
        "--set", "image.pullSecrets[0].name=ecr-credentials")
}

_, err = executeCommand("helm", helmArgs...)
```

### 2.2 Namespace Creation Pattern

**Redis Pattern (from `kindenv_start.go` lines 672-684):**
```go
// Create namespace
namespaceYaml, err := executeCommand("kubectl", "create", "namespace", "redis", "--dry-run=client", "-o", "yaml")
if err != nil {
    fmt.Printf("%s Error creating Redis namespace: %v\n", red("❌"), err)
    os.Exit(1)
}

cmd := exec.Command("kubectl", "apply", "-f", "-")
cmd.Stdin = strings.NewReader(namespaceYaml)
if err := cmd.Run(); err != nil {
    fmt.Printf("%s Error applying Redis namespace: %v\n", red("❌"), err)
    os.Exit(1)
}
```

**MySQL Equivalent Pattern:**
```go
// Create namespace
namespaceYaml, err := executeCommand("kubectl", "create", "namespace", "mysql", "--dry-run=client", "-o", "yaml")
if err != nil {
    fmt.Printf("%s Error creating MySQL namespace: %v\n", red("❌"), err)
    os.Exit(1)
}

cmd := exec.Command("kubectl", "apply", "-f", "-")
cmd.Stdin = strings.NewReader(namespaceYaml)
if err := cmd.Run(); err != nil {
    fmt.Printf("%s Error applying MySQL namespace: %v\n", red("❌"), err)
    os.Exit(1)
}
```

### 2.3 ECR Credentials Setup Pattern

**Redis Pattern (from `kindenv_start.go` lines 686-693):**
```go
// Set up ECR credentials if needed
if config.Images.UseAwsEcr {
    err = setupECRCreds("redis", ecrRegistry, ecrPassword)
    if err != nil {
        fmt.Printf("%s Error setting up ECR credentials for Redis: %v\n", red("❌"), err)
        os.Exit(1)
    }
}
```

**MySQL Equivalent Pattern:**
```go
// Set up ECR credentials if needed
if config.Images.UseAwsEcr {
    err = setupECRCreds("mysql", ecrRegistry, ecrPassword)
    if err != nil {
        fmt.Printf("%s Error setting up ECR credentials for MySQL: %v\n", red("❌"), err)
        os.Exit(1)
    }
}
```

### 2.4 Pod Readiness Wait Pattern

**Redis Pattern (from `kindenv_start.go` lines 710-751):**
```go
// Wait for Redis to be ready using label selectors
fmt.Println(yellow("Waiting for Redis to be ready..."))

// Give resources a moment to be created
fmt.Println(yellow("Pausing briefly to allow resources to be created..."))
time.Sleep(10 * time.Second)

// Wait specifically for the redis-master-0 pod
fmt.Println(yellow("Waiting for redis-master-0 pod to be created..."))
podCheckCmd := exec.Command("kubectl", "get", "pod", "redis-master-0", "-n", "redis", "--no-headers")

// Retry a few times for the pod to appear
var podExists bool
for i := 0; i < 6; i++ {
    podOutput, err := podCheckCmd.CombinedOutput()
    if err == nil && len(podOutput) > 0 {
        podExists = true
        break
    }
    if i < 5 {
        fmt.Printf("Waiting for redis-master-0 pod to appear (attempt %d/6)...\n", i+1)
        time.Sleep(5 * time.Second)
    }
}

if podExists {
    fmt.Println(yellow("Found redis-master-0 pod, waiting for it to be ready..."))
    _, err = executeCommand("kubectl", "wait", "--for=condition=Ready", "pod/redis-master-0", "-n", "redis", "--timeout=2m")
    if err != nil {
        fmt.Printf("%s Warning: Redis master pod is not ready: %v\n", yellow("⚠️"), err)
        fmt.Println(yellow("Continuing despite Redis not being fully ready..."))
    } else {
        fmt.Printf("%s Redis is ready\n", green("✅"))
    }
}
```

**MySQL Equivalent Pattern:**
```go
// Wait for MySQL to be ready
fmt.Println(yellow("Waiting for MySQL to be ready..."))

// Give resources a moment to be created
fmt.Println(yellow("Pausing briefly to allow resources to be created..."))
time.Sleep(10 * time.Second)

// Wait specifically for the mysql-primary-0 pod (StatefulSet naming)
fmt.Println(yellow("Waiting for mysql-primary-0 pod to be created..."))
podCheckCmd := exec.Command("kubectl", "get", "pod", "mysql-primary-0", "-n", "mysql", "--no-headers")

// Retry a few times for the pod to appear
var podExists bool
for i := 0; i < 10; i++ { // MySQL may take longer to start
    podOutput, err := podCheckCmd.CombinedOutput()
    if err == nil && len(podOutput) > 0 {
        podExists = true
        break
    }
    if i < 9 {
        fmt.Printf("Waiting for mysql-primary-0 pod to appear (attempt %d/10)...\n", i+1)
        time.Sleep(5 * time.Second)
    }
}

if podExists {
    fmt.Println(yellow("Found mysql-primary-0 pod, waiting for it to be ready..."))
    // MySQL startup probe has longer initial delay, so use longer timeout
    _, err = executeCommand("kubectl", "wait", "--for=condition=Ready", "pod/mysql-primary-0", "-n", "mysql", "--timeout=5m")
    if err != nil {
        fmt.Printf("%s Warning: MySQL primary pod is not ready: %v\n", yellow("⚠️"), err)
        fmt.Println(yellow("Continuing despite MySQL not being fully ready..."))
    } else {
        fmt.Printf("%s MySQL is ready\n", green("✅"))
    }
} else {
    fmt.Printf("%s MySQL primary pod (mysql-primary-0) not found\n", yellow("⚠️"))
    fmt.Println(yellow("Continuing despite MySQL pod not being detected..."))
}
```

### 2.5 Secret Creation Pattern

**MySQL Secret Pattern (from `kindenv_start.go` lines 641-666):**
```go
// Create MySQL secret if MySQL secrets are enabled
if config.Secrets.MySQL.Enabled {
    fmt.Println(yellow("Creating MySQL credentials secret"))

    mysqlSecretYaml := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  username: "%s"
  password: "%s"
`, config.Secrets.MySQL.Name, config.Secrets.MySQL.Namespace,
        config.Secrets.MySQL.Username, config.Secrets.MySQL.Password)

    cmd := exec.Command("kubectl", "apply", "-f", "-")
    cmd.Stdin = strings.NewReader(mysqlSecretYaml)
    err = cmd.Run()
    if err != nil {
        fmt.Printf("%s Error creating MySQL secret: %v\n", red("❌"), err)
        os.Exit(1)
    }

    fmt.Printf("%s MySQL credentials secret created\n", green("✅"))
}
```

**Note:** The Bitnami MySQL chart can use `auth.existingSecret` to reference this secret, but it expects keys:
- `mysql-root-password`
- `mysql-password`

So the secret creation should be:
```go
mysqlSecretYaml := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  mysql-root-password: "%s"
  mysql-password: "%s"
`, config.Secrets.MySQL.Name, config.Secrets.MySQL.Namespace,
    config.Secrets.MySQL.Password, config.Secrets.MySQL.Password)
```

### 2.6 Port Mapping Pattern

**Redis Pattern (from `kindenv_start.go` lines 346-353):**
```go
// Ensure Redis port is mapped if enabled
if config.Components.Redis.Enabled {
    redisPortKey := fmt.Sprintf("%d/TCP", config.Components.Redis.NodePorts.Redis)
    if !mappedPorts[redisPortKey] {
        fmt.Printf("%s Adding missing Redis port mapping\n", yellow("➕"))
        addPortMapping(config.Components.Redis.NodePorts.Redis, 6379, "TCP")
    }
}
```

**MySQL Equivalent Pattern:**
```go
// Ensure MySQL port is mapped if enabled
if config.Components.MySQL.Enabled {
    mysqlPortKey := fmt.Sprintf("%d/TCP", config.Components.MySQL.NodePorts.MySQL)
    if !mappedPorts[mysqlPortKey] {
        fmt.Printf("%s Adding missing MySQL port mapping\n", yellow("➕"))
        addPortMapping(config.Components.MySQL.NodePorts.MySQL, 3306, "TCP")
    }
}
```

## 3. Best Practices

### 3.1 MySQL Security in Development Environments

**Recommendations:**
1. **Use Strong Passwords:** Even in dev, use non-trivial passwords to catch password-related issues early
2. **Disable Network Policies:** For local development, set `networkPolicy.enabled=false` to avoid connectivity issues
3. **Use Existing Secrets:** Leverage Kubernetes secrets rather than passing passwords via Helm values
4. **Root Password Required:** Always set `auth.rootPassword` as it's required for health probes to work

**Example Configuration:**
```yaml
auth:
  rootPassword: "dev-password-123"  # Required for probes
  database: "dev_db"
  username: "dev_user"
  password: "dev-password-123"
networkPolicy:
  enabled: false  # Disable for local dev
```

### 3.2 Resource Sizing for Local Development

**Recommended Defaults:**
- **CPU Request:** `500m` (0.5 cores) - sufficient for most dev workloads
- **CPU Limit:** `1000m` (1 core) - prevents resource exhaustion
- **Memory Request:** `1Gi` - minimum for MySQL 8
- **Memory Limit:** `2Gi` - allows for query buffers and connections
- **Storage:** `8Gi` (if persistence enabled) - sufficient for dev databases

**Resource Preset Alternative:**
Instead of explicit resources, use `primary.resourcesPreset: "small"` which provides:
- CPU: ~500m request, ~1000m limit
- Memory: ~1Gi request, ~2Gi limit

**Example Configuration:**
```yaml
primary:
  resourcesPreset: "small"  # Simple option
  # OR explicit resources:
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1000m
      memory: 2Gi
```

### 3.3 Health Check Configuration

**Recommended Settings for Development:**
- **Startup Probe:** Critical for MySQL as it takes time to initialize
  - `initialDelaySeconds: 30` - MySQL needs time to start
  - `periodSeconds: 10`
  - `failureThreshold: 10` - Allow up to 100 seconds for startup
- **Readiness Probe:** Check if MySQL can accept connections
  - `initialDelaySeconds: 5`
  - `periodSeconds: 10`
- **Liveness Probe:** Check if MySQL is still running
  - `initialDelaySeconds: 30` - Wait for initial startup
  - `periodSeconds: 10`

**Example Configuration:**
```yaml
primary:
  startupProbe:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 1
    failureThreshold: 10
  readinessProbe:
    enabled: true
    initialDelaySeconds: 5
    periodSeconds: 10
  livenessProbe:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 10
```

### 3.4 Persistence Configuration

**Development Environment Recommendations:**
- **Disable Persistence by Default:** Faster startup, cleaner state
  - `primary.persistence.enabled: false`
- **Enable When Needed:** For testing data persistence scenarios
  - `primary.persistence.enabled: true`
  - `primary.persistence.size: 8Gi`

**Example Configuration:**
```yaml
primary:
  persistence:
    enabled: false  # Faster startup for dev
    # When enabled:
    size: 8Gi
    storageClass: ""  # Use default storage class
    accessModes:
      - ReadWriteOnce
```

### 3.5 Image Repository Best Practices

**For bitnamilegacy Images:**
- Use specific tags (not `latest`) for reproducibility
- Example: `bitnamilegacy/mysql:8.0.35-debian-11-r0`
- Set `image.pullPolicy: IfNotPresent` to avoid unnecessary pulls

**For ECR Integration:**
- Set `global.imageRegistry` to ECR registry URL
- Use `image.pullSecrets` to reference ECR credentials secret
- Example:
  ```yaml
  global:
    imageRegistry: "992979781608.dkr.ecr.eu-west-1.amazonaws.com"
  image:
    repository: bitnamilegacy/mysql
    pullSecrets:
      - name: ecr-credentials
  ```

### 3.6 Architecture Configuration

**For Development:**
- Use `architecture: standalone` (default)
- Avoid replication in local dev (adds complexity and resource usage)
- Only enable replication if specifically testing replication features

**Example Configuration:**
```yaml
architecture: standalone  # Simple single-instance for dev
```

## 4. Complete Example Implementation

### 4.1 Config Structure (from `internal/kindenv/config.go`)

```go
Components struct {
    // ... existing components ...
    MySQL struct {
        Enabled      bool   `yaml:"enabled"`
        Namespace    string `yaml:"namespace"`
        ChartVersion string `yaml:"chartVersion"`
        Database     string `yaml:"database"`
        NodePorts    struct {
            MySQL int `yaml:"mysql"`
        } `yaml:"nodePorts"`
        Persistence struct {
            Enabled bool   `yaml:"enabled"`
            Size    string `yaml:"size"`
        } `yaml:"persistence"`
        Resources struct {
            Requests struct {
                CPU    string `yaml:"cpu"`
                Memory string `yaml:"memory"`
            } `yaml:"requests"`
            Limits struct {
                CPU    string `yaml:"cpu"`
                Memory string `yaml:"memory"`
            } `yaml:"limits"`
        } `yaml:"resources"`
    } `yaml:"mysql"`
}
```

### 4.2 Default Configuration Values

```go
config.Components.MySQL.Enabled = false  // Disabled by default
config.Components.MySQL.Namespace = "mysql"
config.Components.MySQL.ChartVersion = "9.4.5"  // Latest stable
config.Components.MySQL.Database = "mysql"
config.Components.MySQL.NodePorts.MySQL = 30306
config.Components.MySQL.Persistence.Enabled = false
config.Components.MySQL.Persistence.Size = "8Gi"
config.Components.MySQL.Resources.Requests.CPU = "500m"
config.Components.MySQL.Resources.Requests.Memory = "1Gi"
config.Components.MySQL.Resources.Limits.CPU = "1000m"
config.Components.MySQL.Resources.Limits.Memory = "2Gi"
```

### 4.3 Complete Installation Function

```go
// Install MySQL
if config.Components.MySQL.Enabled {
    fmt.Println(yellow("Installing MySQL"))

    // Create namespace
    namespaceYaml, err := executeCommand("kubectl", "create", "namespace", config.Components.MySQL.Namespace, "--dry-run=client", "-o", "yaml")
    if err != nil {
        fmt.Printf("%s Error creating MySQL namespace: %v\n", red("❌"), err)
        os.Exit(1)
    }

    cmd := exec.Command("kubectl", "apply", "-f", "-")
    cmd.Stdin = strings.NewReader(namespaceYaml)
    if err := cmd.Run(); err != nil {
        fmt.Printf("%s Error applying MySQL namespace: %v\n", red("❌"), err)
        os.Exit(1)
    }

    // Set up ECR credentials if needed
    if config.Images.UseAwsEcr {
        err = setupECRCreds(config.Components.MySQL.Namespace, ecrRegistry, ecrPassword)
        if err != nil {
            fmt.Printf("%s Error setting up ECR credentials for MySQL: %v\n", red("❌"), err)
            os.Exit(1)
        }
    }

    // Build Helm arguments
    helmArgs := []string{
        "upgrade", "--install",
        "mysql", "bitnami/mysql",
        "--namespace", config.Components.MySQL.Namespace,
        "--version", config.Components.MySQL.ChartVersion,
        "--set", "architecture=standalone",
        "--set", "primary.service.type=NodePort",
        "--set", fmt.Sprintf("primary.service.nodePorts.mysql=%d", config.Components.MySQL.NodePorts.MySQL),
        "--set", fmt.Sprintf("auth.rootPassword=%s", config.Secrets.MySQL.Password),
        "--set", fmt.Sprintf("auth.database=%s", config.Components.MySQL.Database),
        "--set", fmt.Sprintf("auth.username=%s", config.Secrets.MySQL.Username),
        "--set", fmt.Sprintf("auth.password=%s", config.Secrets.MySQL.Password),
        "--set", fmt.Sprintf("primary.persistence.enabled=%t", config.Components.MySQL.Persistence.Enabled),
        "--set", fmt.Sprintf("primary.resources.requests.cpu=%s", config.Components.MySQL.Resources.Requests.CPU),
        "--set", fmt.Sprintf("primary.resources.requests.memory=%s", config.Components.MySQL.Resources.Requests.Memory),
        "--set", fmt.Sprintf("primary.resources.limits.cpu=%s", config.Components.MySQL.Resources.Limits.CPU),
        "--set", fmt.Sprintf("primary.resources.limits.memory=%s", config.Components.MySQL.Resources.Limits.Memory),
        "--set", "image.repository=bitnamilegacy/mysql",
        "--set", "networkPolicy.enabled=false",  // Disable for local dev
    }

    // Add ECR image registry override if enabled
    if config.Images.UseAwsEcr {
        helmArgs = append(helmArgs,
            "--set", fmt.Sprintf("global.imageRegistry=%s", ecrRegistry),
            "--set", "image.pullSecrets[0].name=ecr-credentials")
    }

    // Execute Helm command
    if verbose {
        fmt.Printf("Command: helm %s\n", strings.Join(helmArgs, " "))
    }

    _, err = executeCommand("helm", helmArgs...)
    if err != nil {
        fmt.Printf("%s Error installing MySQL: %v\n", red("❌"), err)
        os.Exit(1)
    }

    // Wait for MySQL to be ready
    fmt.Println(yellow("Waiting for MySQL to be ready..."))
    time.Sleep(10 * time.Second)

    podCheckCmd := exec.Command("kubectl", "get", "pod", "mysql-primary-0", "-n", config.Components.MySQL.Namespace, "--no-headers")
    var podExists bool
    for i := 0; i < 10; i++ {
        podOutput, err := podCheckCmd.CombinedOutput()
        if err == nil && len(podOutput) > 0 {
            podExists = true
            break
        }
        if i < 9 {
            fmt.Printf("Waiting for mysql-primary-0 pod to appear (attempt %d/10)...\n", i+1)
            time.Sleep(5 * time.Second)
        }
    }

    if podExists {
        fmt.Println(yellow("Found mysql-primary-0 pod, waiting for it to be ready..."))
        _, err = executeCommand("kubectl", "wait", "--for=condition=Ready", "pod/mysql-primary-0", "-n", config.Components.MySQL.Namespace, "--timeout=5m")
        if err != nil {
            fmt.Printf("%s Warning: MySQL primary pod is not ready: %v\n", yellow("⚠️"), err)
            fmt.Println(yellow("Continuing despite MySQL not being fully ready..."))
        } else {
            fmt.Printf("%s MySQL is ready\n", green("✅"))
        }
    } else {
        fmt.Printf("%s MySQL primary pod (mysql-primary-0) not found\n", yellow("⚠️"))
        fmt.Println(yellow("Continuing despite MySQL pod not being detected..."))
    }
}
```

## 5. Summary

### Key Helm Parameters for MySQL Integration:
1. **Credentials:** `auth.rootPassword`, `auth.database`, `auth.username`, `auth.password`
2. **Service:** `primary.service.type=NodePort`, `primary.service.nodePorts.mysql`
3. **Persistence:** `primary.persistence.enabled`, `primary.persistence.size`
4. **Resources:** `primary.resources.requests/limits.cpu/memory`
5. **Image:** `image.repository=bitnamilegacy/mysql`, `global.imageRegistry` (for ECR)
6. **Health:** `primary.startupProbe/readinessProbe/livenessProbe` settings

### Integration Patterns:
- Follow Redis implementation patterns for namespace creation, ECR setup, and pod readiness waiting
- Use StatefulSet pod naming: `mysql-primary-0` (not `mysql-master-0`)
- Allow longer startup timeouts (5 minutes) due to MySQL initialization
- Use existing secrets pattern for credentials management

### Best Practices:
- Disable persistence by default for faster dev startup
- Use resource presets (`small`) or explicit limits (500m CPU, 1Gi memory minimum)
- Configure proper health probes with appropriate delays
- Use bitnamilegacy images with specific tags
- Disable network policies for local development
