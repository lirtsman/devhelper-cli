# Data Model: Custom Components for KindEnv

**Feature**: Custom Components for KindEnv  
**Date**: 2026-01-30  
**Phase**: 1 - Design & Contracts

## Overview

This document defines the data structures, relationships, and validation rules for custom component support in kindenv. All structures are designed to integrate seamlessly with the existing KindEnvConfig and follow Kubernetes resource patterns.

## Core Entities

### 1. CustomComponent

**Purpose**: Represents a user-defined service to be deployed in the Kind cluster.

**Go Struct Definition**:

```go
// CustomComponent defines a user-configured service deployment
type CustomComponent struct {
    // Identity
    Name        string `yaml:"name" json:"name"`
    Enabled     bool   `yaml:"enabled" json:"enabled"`
    Namespace   string `yaml:"namespace" json:"namespace"`
    
    // Container Specification
    Image       string   `yaml:"image" json:"image"`
    Command     []string `yaml:"command,omitempty" json:"command,omitempty"`
    Args        []string `yaml:"args,omitempty" json:"args,omitempty"`
    
    // Configuration
    Env         []EnvVar              `yaml:"env,omitempty" json:"env,omitempty"`
    Ports       []PortMapping         `yaml:"ports,omitempty" json:"ports,omitempty"`
    Resources   *ResourceRequirements `yaml:"resources,omitempty" json:"resources,omitempty"`
    ConfigFiles []ConfigFile          `yaml:"configFiles,omitempty" json:"configFiles,omitempty"`
    
    // Scaling and Metadata
    Replicas    *int                  `yaml:"replicas,omitempty" json:"replicas,omitempty"`
    Labels      map[string]string     `yaml:"labels,omitempty" json:"labels,omitempty"`
    Annotations map[string]string     `yaml:"annotations,omitempty" json:"annotations,omitempty"`
    
    // Health Checks (future enhancement)
    ReadinessProbe *Probe `yaml:"readinessProbe,omitempty" json:"readinessProbe,omitempty"`
    LivenessProbe  *Probe `yaml:"livenessProbe,omitempty" json:"livenessProbe,omitempty"`
}
```

**Field Descriptions**:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| Name | string | **Yes** | - | Unique identifier for the component (DNS-compatible) |
| Image | string | **Yes** | - | Container image (format: `[registry/]repository[:tag]`) |
| Enabled | bool | Optional | true | Whether to deploy this component |
| Namespace | string | Optional | "default" | Kubernetes namespace for deployment |
| Replicas | *int | Optional | 1 | Number of pod replicas |
| Resources | *ResourceRequirements | Optional | See below* | CPU and memory limits |
| Command | []string | Optional | nil (use image default) | Override container entrypoint |
| Args | []string | Optional | nil (use image default) | Arguments to container command |
| Env | []EnvVar | Optional | nil (no env vars) | Environment variables |
| Ports | []PortMapping | Optional | nil (no ports exposed) | Port mappings for external access |
| ConfigFiles | []ConfigFile | Optional | nil (no files mounted) | Configuration files to mount |
| Labels | map[string]string | Optional | auto-generated** | Kubernetes labels |
| Annotations | map[string]string | Optional | nil | Kubernetes annotations |

**Default Resource Requirements*:
- Requests: CPU "100m", Memory "128Mi"
- Limits: CPU "500m", Memory "512Mi"

**Auto-generated Labels*:
- `app: <component-name>`
- `managed-by: kindenv`
- `component-type: custom`

**Validation Rules**:

```go
func (c *CustomComponent) Validate() error {
    // Required fields
    if c.Name == "" {
        return fmt.Errorf("custom component name is required")
    }
    if c.Image == "" {
        return fmt.Errorf("custom component '%s': image is required", c.Name)
    }
    
    // Name format (DNS-1123 label)
    if !isValidDNSLabel(c.Name) {
        return fmt.Errorf("custom component name '%s' must be valid DNS label (lowercase, alphanumeric, hyphens)", c.Name)
    }
    
    // Image format validation
    if !isValidImageFormat(c.Image) {
        return fmt.Errorf("custom component '%s': invalid image format '%s' (expected: [registry/]repository[:tag])", c.Name, c.Image)
    }
    
    // Namespace validation
    if c.Namespace == "" {
        c.Namespace = "default" // Set default
    }
    
    // Replicas validation
    if c.Replicas != nil && *c.Replicas < 1 {
        return fmt.Errorf("custom component '%s': replicas must be >= 1", c.Name)
    }
    if c.Replicas == nil {
        replicas := 1
        c.Replicas = &replicas
    }
    
    // Environment variables validation
    for i, env := range c.Env {
        if err := env.Validate(); err != nil {
            return fmt.Errorf("custom component '%s': env[%d]: %w", c.Name, i, err)
        }
    }
    
    // Port validation
    for i, port := range c.Ports {
        if err := port.Validate(); err != nil {
            return fmt.Errorf("custom component '%s': ports[%d]: %w", c.Name, i, err)
        }
    }
    
    // Resource validation
    if c.Resources != nil {
        if err := c.Resources.Validate(); err != nil {
            return fmt.Errorf("custom component '%s': resources: %w", c.Name, err)
        }
    } else {
        c.Resources = defaultResourceRequirements()
    }
    
    // Config files validation
    if len(c.ConfigFiles) > 0 {
        pathsSeen := make(map[string]bool)
        namesSeen := make(map[string]bool)
        totalSize := 0
        
        for i, cf := range c.ConfigFiles {
            if err := cf.Validate(); err != nil {
                return fmt.Errorf("custom component '%s': configFiles[%d]: %w", c.Name, i, err)
            }
            
            // Check for duplicate mount paths
            if pathsSeen[cf.Path] {
                return fmt.Errorf("custom component '%s': duplicate mount path '%s' in config files", c.Name, cf.Path)
            }
            pathsSeen[cf.Path] = true
            
            // Check for duplicate filenames
            if namesSeen[cf.Name] {
                return fmt.Errorf("custom component '%s': duplicate config file name '%s'", c.Name, cf.Name)
            }
            namesSeen[cf.Name] = true
            
            // Track total size (ConfigMap limit is 1MB)
            totalSize += len(cf.Contents)
        }
        
        if totalSize > 1024*1024 {
            return fmt.Errorf("custom component '%s': total config files size exceeds 1MB limit (%d bytes)", c.Name, totalSize)
        }
    }
    
    return nil
}
```

**Default Values**:

```go
func (c *CustomComponent) SetDefaults() {
    if c.Namespace == "" {
        c.Namespace = "default"
    }
    
    if c.Replicas == nil {
        replicas := 1
        c.Replicas = &replicas
    }
    
    if c.Resources == nil {
        c.Resources = defaultResourceRequirements()
    }
    
    if c.Labels == nil {
        c.Labels = make(map[string]string)
    }
    
    // Auto-generate standard labels
    c.Labels["app"] = c.Name
    c.Labels["managed-by"] = "kindenv"
    c.Labels["component-type"] = "custom"
}
```

### 2. EnvVar

**Purpose**: Represents an environment variable with support for direct values or secret references.

**Go Struct Definition**:

```go
// EnvVar defines an environment variable for a container
type EnvVar struct {
    Name      string         `yaml:"name" json:"name"`
    Value     string         `yaml:"value,omitempty" json:"value,omitempty"`
    ValueFrom *EnvVarSource  `yaml:"valueFrom,omitempty" json:"valueFrom,omitempty"`
}

// EnvVarSource defines the source for an environment variable value
type EnvVarSource struct {
    SecretKeyRef *SecretKeySelector `yaml:"secretKeyRef,omitempty" json:"secretKeyRef,omitempty"`
    // Future: ConfigMapKeyRef, FieldRef, ResourceFieldRef
}

// SecretKeySelector selects a key from a Secret
type SecretKeySelector struct {
    Name string `yaml:"name" json:"name"`
    Key  string `yaml:"key" json:"key"`
}
```

**Field Descriptions**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Name | string | Yes | Environment variable name |
| Value | string | Conditional | Direct value (mutually exclusive with ValueFrom) |
| ValueFrom | *EnvVarSource | Conditional | Reference to secret/configmap |
| SecretKeyRef.Name | string | If using secrets | Secret name |
| SecretKeyRef.Key | string | If using secrets | Key within the secret |

**Validation Rules**:

```go
func (e *EnvVar) Validate() error {
    if e.Name == "" {
        return fmt.Errorf("environment variable name is required")
    }
    
    // Validate env var name format (must be valid shell variable name)
    if !isValidEnvVarName(e.Name) {
        return fmt.Errorf("invalid environment variable name '%s' (must match [A-Z_][A-Z0-9_]*)", e.Name)
    }
    
    // Either Value or ValueFrom must be set, but not both
    hasValue := e.Value != ""
    hasValueFrom := e.ValueFrom != nil
    
    if !hasValue && !hasValueFrom {
        return fmt.Errorf("environment variable '%s': either 'value' or 'valueFrom' must be specified", e.Name)
    }
    
    if hasValue && hasValueFrom {
        return fmt.Errorf("environment variable '%s': 'value' and 'valueFrom' are mutually exclusive", e.Name)
    }
    
    // Validate ValueFrom if present
    if hasValueFrom {
        if e.ValueFrom.SecretKeyRef != nil {
            if e.ValueFrom.SecretKeyRef.Name == "" {
                return fmt.Errorf("environment variable '%s': secretKeyRef.name is required", e.Name)
            }
            if e.ValueFrom.SecretKeyRef.Key == "" {
                return fmt.Errorf("environment variable '%s': secretKeyRef.key is required", e.Name)
            }
        }
    }
    
    return nil
}
```

**Usage Examples**:

```yaml
env:
  # Direct value
  - name: APP_ENV
    value: "development"
  
  # Secret reference
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: mysql-secret
        key: password
```

### 3. PortMapping

**Purpose**: Represents port exposure configuration for accessing services from the host machine.

**Go Struct Definition**:

```go
// PortMapping defines port exposure for a custom component
type PortMapping struct {
    ContainerPort int    `yaml:"containerPort" json:"containerPort"`
    HostPort      int    `yaml:"hostPort,omitempty" json:"hostPort,omitempty"`
    Protocol      string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
    NodePort      int    `yaml:"nodePort,omitempty" json:"nodePort,omitempty"`
    ServiceName   string `yaml:"serviceName,omitempty" json:"serviceName,omitempty"`
}
```

**Field Descriptions**:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| ContainerPort | int | Yes | - | Port exposed by the container |
| HostPort | int | No | auto-assigned | Port on the host machine (for Kind port mapping) |
| Protocol | string | No | "TCP" | Protocol (TCP or UDP) |
| NodePort | int | No | auto-assigned | NodePort in range 30000-32767 |
| ServiceName | string | No | component-name | Kubernetes Service name |

**Validation Rules**:

```go
func (p *PortMapping) Validate() error {
    // Container port validation
    if p.ContainerPort < 1 || p.ContainerPort > 65535 {
        return fmt.Errorf("containerPort must be between 1 and 65535, got %d", p.ContainerPort)
    }
    
    // Host port validation (if specified)
    if p.HostPort != 0 {
        if p.HostPort < 1024 || p.HostPort > 65535 {
            return fmt.Errorf("hostPort must be between 1024 and 65535, got %d", p.HostPort)
        }
    }
    
    // NodePort validation (if specified)
    if p.NodePort != 0 {
        if p.NodePort < 30000 || p.NodePort > 32767 {
            return fmt.Errorf("nodePort must be between 30000 and 32767, got %d", p.NodePort)
        }
    }
    
    // Protocol validation
    if p.Protocol == "" {
        p.Protocol = "TCP"
    }
    p.Protocol = strings.ToUpper(p.Protocol)
    if p.Protocol != "TCP" && p.Protocol != "UDP" {
        return fmt.Errorf("protocol must be TCP or UDP, got %s", p.Protocol)
    }
    
    return nil
}
```

**Port Assignment Logic**:

```go
// assignPorts assigns NodePort and HostPort if not explicitly set
func (c *CustomComponent) assignPorts(usedPorts map[int]bool) error {
    for i := range c.Ports {
        port := &c.Ports[i]
        
        // Auto-assign NodePort if not specified
        if port.NodePort == 0 {
            nodePort, err := findAvailableNodePort(usedPorts)
            if err != nil {
                return fmt.Errorf("failed to assign NodePort for component '%s': %w", c.Name, err)
            }
            port.NodePort = nodePort
            usedPorts[nodePort] = true
        } else {
            // Validate specified NodePort is not in use
            if usedPorts[port.NodePort] {
                return fmt.Errorf("NodePort %d is already in use (component: %s)", port.NodePort, c.Name)
            }
            usedPorts[port.NodePort] = true
        }
        
        // Default HostPort to ContainerPort if not specified
        if port.HostPort == 0 {
            port.HostPort = port.ContainerPort
        }
    }
    
    return nil
}
```

### 4. ResourceRequirements

**Purpose**: Represents CPU and memory resource requests and limits.

**Go Struct Definition**:

```go
// ResourceRequirements defines resource requests and limits
type ResourceRequirements struct {
    Requests *ResourceList `yaml:"requests,omitempty" json:"requests,omitempty"`
    Limits   *ResourceList `yaml:"limits,omitempty" json:"limits,omitempty"`
}

// ResourceList defines CPU and memory quantities
type ResourceList struct {
    CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
    Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
}
```

**Field Descriptions**:

| Field | Format | Example | Description |
|-------|--------|---------|-------------|
| Requests.CPU | Cores or millicores | "500m", "0.5", "1" | Guaranteed CPU allocation |
| Requests.Memory | Bytes with unit | "512Mi", "1Gi" | Guaranteed memory allocation |
| Limits.CPU | Cores or millicores | "1000m", "2" | Maximum CPU usage |
| Limits.Memory | Bytes with unit | "1Gi", "2Gi" | Maximum memory usage |

**Validation Rules**:

```go
func (r *ResourceRequirements) Validate() error {
    if r.Requests != nil {
        if err := r.Requests.Validate(); err != nil {
            return fmt.Errorf("requests: %w", err)
        }
    }
    
    if r.Limits != nil {
        if err := r.Limits.Validate(); err != nil {
            return fmt.Errorf("limits: %w", err)
        }
    }
    
    // Validate that limits >= requests (if both specified)
    if r.Requests != nil && r.Limits != nil {
        if err := validateResourceLimitsGreaterThanRequests(r.Requests, r.Limits); err != nil {
            return err
        }
    }
    
    return nil
}

func (r *ResourceList) Validate() error {
    if r.CPU != "" {
        if !isValidCPUQuantity(r.CPU) {
            return fmt.Errorf("invalid CPU quantity '%s' (examples: 100m, 0.5, 1)", r.CPU)
        }
    }
    
    if r.Memory != "" {
        if !isValidMemoryQuantity(r.Memory) {
            return fmt.Errorf("invalid memory quantity '%s' (examples: 128Mi, 1Gi, 512M)", r.Memory)
        }
    }
    
    return nil
}
```

**Default Resource Requirements**:

```go
func defaultResourceRequirements() *ResourceRequirements {
    return &ResourceRequirements{
        Requests: &ResourceList{
            CPU:    "100m",
            Memory: "128Mi",
        },
        Limits: &ResourceList{
            CPU:    "500m",
            Memory: "512Mi",
        },
    }
}
```

### 5. ConfigFile

**Purpose**: Represents a custom configuration file to be mounted into the container.

**Go Struct Definition**:

```go
// ConfigFile defines a configuration file to be mounted in the container
type ConfigFile struct {
    Name     string `yaml:"name" json:"name"`
    Path     string `yaml:"path" json:"path"`
    Contents string `yaml:"contents" json:"contents"`
}
```

**Field Descriptions**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Name | string | Yes | Filename (used as ConfigMap key) |
| Path | string | Yes | Absolute mount path in container (e.g., `/config/application.yaml`) |
| Contents | string | Yes | Inline file contents (YAML, JSON, XML, properties, etc.) |

**Validation Rules**:

```go
func (cf *ConfigFile) Validate() error {
    if cf.Name == "" {
        return fmt.Errorf("config file name is required")
    }
    
    if cf.Path == "" {
        return fmt.Errorf("config file '%s': path is required", cf.Name)
    }
    
    if cf.Contents == "" {
        return fmt.Errorf("config file '%s': contents cannot be empty", cf.Name)
    }
    
    // Validate path format (must be absolute)
    if !strings.HasPrefix(cf.Path, "/") {
        return fmt.Errorf("config file '%s': path must be absolute (start with /), got '%s'", cf.Name, cf.Path)
    }
    
    // Validate filename has no directory separators
    if strings.Contains(cf.Name, "/") || strings.Contains(cf.Name, "\\") {
        return fmt.Errorf("config file name '%s' cannot contain directory separators", cf.Name)
    }
    
    // Validate contents size (ConfigMap limit is 1MB)
    if len(cf.Contents) > 1024*1024 {
        return fmt.Errorf("config file '%s': contents exceed 1MB limit (%d bytes)", cf.Name, len(cf.Contents))
    }
    
    return nil
}
```

**ConfigMap Generation**:

```go
// GenerateConfigMap creates a Kubernetes ConfigMap for the component's config files
func (c *CustomComponent) GenerateConfigMap() (*ConfigMapSpec, error) {
    if len(c.ConfigFiles) == 0 {
        return nil, nil // No ConfigMap needed
    }
    
    data := make(map[string]string)
    for _, cf := range c.ConfigFiles {
        // Validate config file
        if err := cf.Validate(); err != nil {
            return nil, err
        }
        
        // Check for duplicate keys (filenames must be unique)
        if _, exists := data[cf.Name]; exists {
            return nil, fmt.Errorf("duplicate config file name '%s'", cf.Name)
        }
        
        data[cf.Name] = cf.Contents
    }
    
    return &ConfigMapSpec{
        Name:      fmt.Sprintf("%s-config", c.Name),
        Namespace: c.Namespace,
        Data:      data,
        Labels: map[string]string{
            "app":        c.Name,
            "managed-by": "kindenv",
        },
    }, nil
}
```

**Volume Mount Generation**:

```go
// GenerateVolumeMounts creates volume and volumeMount specs for config files
func (c *CustomComponent) GenerateVolumeMounts() ([]VolumeSpec, []VolumeMountSpec, error) {
    if len(c.ConfigFiles) == 0 {
        return nil, nil, nil
    }
    
    // Single ConfigMap volume for all config files
    volumes := []VolumeSpec{
        {
            Name: fmt.Sprintf("%s-config-volume", c.Name),
            ConfigMap: &ConfigMapVolumeSource{
                Name:        fmt.Sprintf("%s-config", c.Name),
                DefaultMode: 0644, // Read-only
            },
        },
    }
    
    // Individual volumeMount for each config file (using subPath)
    var mounts []VolumeMountSpec
    pathsSeen := make(map[string]bool)
    
    for _, cf := range c.ConfigFiles {
        // Check for duplicate mount paths
        if pathsSeen[cf.Path] {
            return nil, nil, fmt.Errorf("duplicate mount path '%s' for config files", cf.Path)
        }
        pathsSeen[cf.Path] = true
        
        mounts = append(mounts, VolumeMountSpec{
            Name:      fmt.Sprintf("%s-config-volume", c.Name),
            MountPath: cf.Path,
            SubPath:   cf.Name, // Mounts individual file from ConfigMap
            ReadOnly:  true,
        })
    }
    
    return volumes, mounts, nil
}
```

**Usage Example**:

```yaml
configFiles:
  - name: application.yaml
    path: /config/application.yaml
    contents: |
      server:
        port: 8080
      database:
        host: mysql.mysql.svc.cluster.local
        port: 3306
  
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
```

### 6. Probe (Health Checks)

**Purpose**: Defines readiness and liveness probes for health monitoring.

**Go Struct Definition**:

```go
// Probe defines a health check for a container
type Probe struct {
    HTTPGet             *HTTPGetAction `yaml:"httpGet,omitempty" json:"httpGet,omitempty"`
    TCPSocket           *TCPSocketAction `yaml:"tcpSocket,omitempty" json:"tcpSocket,omitempty"`
    Exec                *ExecAction `yaml:"exec,omitempty" json:"exec,omitempty"`
    InitialDelaySeconds int         `yaml:"initialDelaySeconds,omitempty" json:"initialDelaySeconds,omitempty"`
    PeriodSeconds       int         `yaml:"periodSeconds,omitempty" json:"periodSeconds,omitempty"`
    TimeoutSeconds      int         `yaml:"timeoutSeconds,omitempty" json:"timeoutSeconds,omitempty"`
    SuccessThreshold    int         `yaml:"successThreshold,omitempty" json:"successThreshold,omitempty"`
    FailureThreshold    int         `yaml:"failureThreshold,omitempty" json:"failureThreshold,omitempty"`
}

type HTTPGetAction struct {
    Path   string `yaml:"path" json:"path"`
    Port   int    `yaml:"port" json:"port"`
    Scheme string `yaml:"scheme,omitempty" json:"scheme,omitempty"`
}

type TCPSocketAction struct {
    Port int `yaml:"port" json:"port"`
}

type ExecAction struct {
    Command []string `yaml:"command" json:"command"`
}
```

**Note**: Health check support is a **future enhancement** (not in MVP). Included in data model for completeness.

## Entity Relationships

```
KindEnvConfig
    └── CustomComponents []CustomComponent
            ├── Env []EnvVar
            │       └── ValueFrom.SecretKeyRef → Kubernetes Secret
            ├── Ports []PortMapping
            │       └── Creates → Kubernetes Service (NodePort)
            ├── Resources *ResourceRequirements
            │       ├── Requests *ResourceList
            │       └── Limits *ResourceList
            ├── ConfigFiles []ConfigFile
            │       └── Generates → Kubernetes ConfigMap
            │               └── Mounted as Volume in Pod
            ├── Labels map[string]string
            ├── Annotations map[string]string
            └── Generates → Kubernetes Deployment
                        └── Pod Template
                                ├── Containers
                                ├── Volumes (ConfigMap volumes)
                                ├── VolumeMounts (subPath for each file)
                                └── ImagePullSecrets
```

## Integration with Existing Config

**Modified KindEnvConfig Struct**:

```go
type KindEnvConfig struct {
    Tools struct {
        // ... existing tool configuration
    } `yaml:"tools"`
    
    Cluster struct {
        // ... existing cluster configuration
    } `yaml:"cluster"`
    
    Components struct {
        // ... existing components (Temporal, Redis, Dapr, etc.)
    } `yaml:"components"`
    
    // NEW: Custom components section
    CustomComponents []CustomComponent `yaml:"customComponents,omitempty"`
    
    Images struct {
        // ... existing image configuration
    } `yaml:"images"`
    
    Secrets struct {
        // ... existing secrets configuration
    } `yaml:"secrets"`
}
```

## Validation Summary

### Configuration-Time Validation

1. **Required Fields**: name, image
2. **Format Validation**: DNS labels, image format, resource quantities
3. **Mutual Exclusivity**: env value vs. valueFrom
4. **Range Validation**: ports, replicas, resource quantities

### Pre-Deployment Validation

1. **Port Conflicts**: No duplicate NodePorts or HostPorts
2. **Namespace Existence**: Target namespace must exist or be created
3. **Secret Existence**: Referenced secrets must exist in target namespace
4. **Registry Access**: Image registry must be accessible (best-effort check)

### Runtime Validation

1. **Kubernetes API**: Deployment/Service creation succeeds
2. **Image Pull**: Container image can be pulled successfully
3. **Pod Startup**: Container starts without crash loops
4. **Health Checks**: Readiness/liveness probes pass (future)

## State Transitions

```
[Configured] → Validate → [Valid]
                ↓
            [Invalid] → Error Report → [User Fix Required]

[Valid] → Pre-Deploy Checks → [Ready to Deploy]
            ↓
        [Validation Failed] → Error Report → [User Fix Required]

[Ready to Deploy] → kubectl apply → [Deploying]
                                        ↓
                                    [Pending] → Image Pull → [Running]
                                        ↓                       ↓
                                [Image Pull Failed]     [Health Check]
                                        ↓                       ↓
                                    [Failed]                [Ready]
```

## Error Handling

### Error Types

```go
type ValidationError struct {
    Component string
    Field     string
    Value     interface{}
    Message   string
}

type DeploymentError struct {
    Component   string
    Phase       string // "image-pull", "pod-start", "readiness"
    K8sError    error
    Suggestions []string
}

type ConflictError struct {
    Component     string
    ConflictType  string // "port", "name"
    ConflictsWith string
}
```

### Error Reporting Format

```go
func (e *ValidationError) Error() string {
    return fmt.Sprintf(
        "❌ Configuration error in component '%s':\n   Field: %s\n   Value: %v\n   Problem: %s",
        e.Component, e.Field, e.Value, e.Message,
    )
}
```

## Performance Considerations

1. **Configuration Parsing**: O(n) for n components, expected <10ms
2. **Validation**: O(n*m) for n components with m ports/env vars, expected <50ms
3. **Deployment**: Parallel deployment of independent components
4. **Status Checks**: Cached results, refreshed every 5 seconds

## Backward Compatibility

- New `customComponents` section is optional
- Existing kindenv.yaml files work without modification
- Configuration version not required (schema is additive)

## Future Extensions

1. **Volumes**: Persistent storage support
2. **Health Checks**: Readiness/liveness probes
3. **ConfigMaps**: Environment variable source
4. **Init Containers**: Pre-startup tasks
5. **Security Context**: Run as non-root, capabilities
6. **Horizontal Pod Autoscaler**: Auto-scaling based on metrics
7. **Ingress**: HTTP routing instead of NodePort

---

**Next**: Generate contracts/ with schema definitions and deployment templates.
