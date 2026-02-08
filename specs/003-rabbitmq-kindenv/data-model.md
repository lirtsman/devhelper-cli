# Data Model: RabbitMQ Support for KindEnv

**Feature**: `003-rabbitmq-kindenv`  
**Date**: 2026-02-05  
**Phase**: 1 (Design & Contracts)

## Overview

This document defines the data structures, entities, and their relationships for the RabbitMQ integration in kindenv. It follows the same architectural pattern established by MySQL (spec 001-mysql8-kindenv).

## Entity Relationship Diagram

```
┌──────────────────────────────────────────────────────┐
│             KindEnvConfig                             │
│  (Root configuration structure)                       │
└───────────┬──────────────────────────────────────────┘
            │
            │ contains
            ▼
┌──────────────────────────────────────────────────────┐
│             Components                                 │
│  (All kindenv components)                             │
│  ┌────────────────────────────────────────────────┐  │
│  │  RabbitMQ                                      │  │
│  │  - Enabled: bool                               │  │
│  │  - Namespace: string                           │  │
│  │  - ChartVersion: string                        │  │
│  │  - VirtualHost: string                         │  │
│  │  - NodePorts: RabbitMQNodePorts               │  │
│  │  - Resources: RabbitMQResources               │  │
│  │  - Persistence: RabbitMQPersistence           │  │
│  └────────────────────────────────────────────────┘  │
└───────────┬──────────────────────────────────────────┘
            │
            │ references
            ▼
┌──────────────────────────────────────────────────────┐
│             Secrets                                    │
│  (Credential management)                              │
│  ┌────────────────────────────────────────────────┐  │
│  │  RabbitMQ                                      │  │
│  │  - Enabled: bool                               │  │
│  │  - Name: string                                │  │
│  │  - Namespace: string                           │  │
│  │  - Username: string                            │  │
│  │  - Password: string                            │  │
│  │  - ErlangCookie: string                        │  │
│  └────────────────────────────────────────────────┘  │
└───────────┬──────────────────────────────────────────┘
            │
            │ creates
            ▼
┌──────────────────────────────────────────────────────┐
│         Runtime Resources                             │
│  ┌────────────────────────────────────────────────┐  │
│  │  Kubernetes Resources                          │  │
│  │  - StatefulSet (rabbitmq-0)                    │  │
│  │  - Services (AMQP + Management)                │  │
│  │  - ConfigMap (configuration)                   │  │
│  │  - Secret (credentials)                        │  │
│  │  - PersistentVolume (optional)                 │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
            │
            │ monitored by
            ▼
┌──────────────────────────────────────────────────────┐
│         RabbitMQStatus                                │
│  (Runtime status information)                         │
│  - State: RabbitMQState                               │
│  - PodReady: bool                                     │
│  - ServiceReady: bool                                 │
│  - AMQPReady: bool                                    │
│  - ManagementReady: bool                              │
│  - ConnectionInfo: RabbitMQConnectionInfo            │
│  - ErrorMessage: string                               │
│  - LastChecked: time.Time                            │
└──────────────────────────────────────────────────────┘
```

## Core Entities

### 1. RabbitMQ Configuration Entity

**Purpose**: Defines user-configurable RabbitMQ component settings in kindenv.yaml

**Location**: `internal/kindenv/config.go` → `KindEnvConfig.Components.RabbitMQ`

**Structure**:
```go
type RabbitMQConfig struct {
    // Enabled indicates whether RabbitMQ should be deployed
    Enabled      bool                    `yaml:"enabled"`
    
    // Namespace is the Kubernetes namespace for RabbitMQ deployment
    Namespace    string                  `yaml:"namespace"`
    
    // ChartVersion specifies the Bitnami RabbitMQ Helm chart version
    ChartVersion string                  `yaml:"chartVersion"`
    
    // VirtualHost is the default RabbitMQ virtual host
    VirtualHost  string                  `yaml:"virtualHost"`
    
    // NodePorts defines port mappings for AMQP and Management UI
    NodePorts    RabbitMQNodePorts      `yaml:"nodePorts"`
    
    // Resources defines CPU, memory, and storage limits
    Resources    RabbitMQResources      `yaml:"resources"`
    
    // Persistence controls data persistence configuration
    Persistence  RabbitMQPersistence    `yaml:"persistence"`
}
```

**Validation Rules**:
- `Enabled`: boolean, no validation required
- `Namespace`: non-empty string, must be valid Kubernetes namespace name
- `ChartVersion`: non-empty string, semantic version format (e.g., "11.0.0")
- `VirtualHost`: optional, defaults to "/", must start with "/" or be alphanumeric
- `NodePorts`: nested object, see RabbitMQNodePorts validation
- `Resources`: nested object, see RabbitMQResources validation
- `Persistence`: nested object, see RabbitMQPersistence validation

**Default Values**:
```yaml
components:
  rabbitmq:
    enabled: false
    namespace: "rabbitmq"
    chartVersion: "11.0.0"
    virtualHost: "/"
    nodePorts:
      amqp: 30672
      management: 31672
    resources:
      cpu: "500m"
      memory: "1Gi"
    persistence:
      enabled: false
      size: "8Gi"
```

### 2. RabbitMQNodePorts Entity

**Purpose**: Defines Kubernetes NodePort mappings for RabbitMQ services

**Structure**:
```go
type RabbitMQNodePorts struct {
    // AMQP is the NodePort for AMQP protocol (5672)
    AMQP       int `yaml:"amqp"`
    
    // Management is the NodePort for Management UI (15672)
    Management int `yaml:"management"`
}
```

**Validation Rules**:
- `AMQP`: must be in range 30000-32767 (Kubernetes NodePort range)
- `Management`: must be in range 30000-32767 (Kubernetes NodePort range)
- `AMQP` and `Management` must be different values

**Kind Cluster Port Mapping**:
- NodePort `AMQP` (30672) → Host Port 5672
- NodePort `Management` (31672) → Host Port 15672

### 3. RabbitMQResources Entity

**Purpose**: Defines resource requests and limits for RabbitMQ pods

**Structure**:
```go
type RabbitMQResources struct {
    // CPU resource specification (e.g., "500m", "1")
    CPU    string `yaml:"cpu"`
    
    // Memory resource specification (e.g., "1Gi", "512Mi")
    Memory string `yaml:"memory"`
}
```

**Validation Rules**:
- `CPU`: must match pattern `^[0-9]+m?$` (e.g., "500m", "1", "2000m")
- `Memory`: must match pattern `^[0-9]+[KMGT]i$` (e.g., "1Gi", "512Mi", "2Gi")

**Resource Interpretation**:
- Both CPU and Memory are used as both requests and limits (guaranteed resources)
- Storage size defined in Persistence entity

### 4. RabbitMQPersistence Entity

**Purpose**: Controls persistent storage for RabbitMQ data (queues, messages, configuration)

**Structure**:
```go
type RabbitMQPersistence struct {
    // Enabled indicates whether persistence is enabled
    Enabled bool   `yaml:"enabled"`
    
    // Size is the PersistentVolume size (e.g., "8Gi", "10Gi")
    Size    string `yaml:"size"`
}
```

**Validation Rules**:
- `Enabled`: boolean, no validation required
- `Size`: required when `Enabled` is true, must match pattern `^[0-9]+[KMGT]i$`

**Persistence Behavior**:
- **Enabled**: Creates PersistentVolumeClaim, data survives pod restarts
- **Disabled**: Uses emptyDir volume, data lost on pod termination

### 5. RabbitMQ Secret Entity

**Purpose**: Stores sensitive RabbitMQ credentials in Kubernetes secrets

**Location**: `internal/kindenv/config.go` → `KindEnvConfig.Secrets.RabbitMQ`

**Structure**:
```go
type RabbitMQSecret struct {
    // Enabled indicates whether secret should be created
    Enabled      bool   `yaml:"enabled"`
    
    // Name is the Kubernetes secret name
    Name         string `yaml:"name"`
    
    // Namespace is the Kubernetes namespace for the secret
    Namespace    string `yaml:"namespace"`
    
    // Username is the RabbitMQ admin username
    Username     string `yaml:"username"`
    
    // Password is the RabbitMQ admin password
    Password     string `yaml:"password"`
    
    // ErlangCookie is the Erlang cluster authentication cookie
    ErlangCookie string `yaml:"erlangCookie"`
}
```

**Validation Rules**:
- `Enabled`: boolean
- `Name`: non-empty string when enabled
- `Namespace`: non-empty string when enabled, must match RabbitMQ namespace
- `Username`: non-empty string when enabled
- `Password`: non-empty string when enabled
- `ErlangCookie`: optional, auto-generated if empty, must be 20+ characters if provided

**Default Values**:
```yaml
secrets:
  rabbitmq:
    enabled: true
    name: "rabbitmq-credentials"
    namespace: "rabbitmq"
    username: "user"
    password: "password"
    erlangCookie: ""  # Auto-generated
```

**Security Considerations**:
- Password should be changed from default in production
- ErlangCookie automatically generated with crypto-random string if not provided
- Secret stored in Kubernetes with base64 encoding

### 6. RabbitMQStatus Entity

**Purpose**: Runtime status information for deployed RabbitMQ instance

**Structure**:
```go
type RabbitMQStatus struct {
    // State represents the current deployment state
    State RabbitMQState `json:"state"`
    
    // PodReady indicates if RabbitMQ pod is ready
    PodReady bool `json:"podReady"`
    
    // ServiceReady indicates if Kubernetes services are ready
    ServiceReady bool `json:"serviceReady"`
    
    // AMQPReady indicates if AMQP port is accessible
    AMQPReady bool `json:"amqpReady"`
    
    // ManagementReady indicates if Management UI is accessible
    ManagementReady bool `json:"managementReady"`
    
    // ConnectionInfo contains connection details for clients
    ConnectionInfo *RabbitMQConnectionInfo `json:"connectionInfo,omitempty"`
    
    // ErrorMessage contains error details if state is Failed
    ErrorMessage string `json:"errorMessage,omitempty"`
    
    // LastChecked is the timestamp of last status check
    LastChecked time.Time `json:"lastChecked"`
}
```

**State Enumeration**:
```go
type RabbitMQState string

const (
    RabbitMQStatePending    RabbitMQState = "Pending"     // Initial state
    RabbitMQStateInstalling RabbitMQState = "Installing"  // Helm install in progress
    RabbitMQStateRunning    RabbitMQState = "Running"     // Fully operational
    RabbitMQStateFailed     RabbitMQState = "Failed"      // Installation or runtime failure
    RabbitMQStateStopped    RabbitMQState = "Stopped"     // Intentionally stopped
)
```

**State Transitions**:
```
Pending → Installing → Running
                ↓         ↓
              Failed    Stopped
```

**Readiness Conditions**:
- **PodReady**: At least one replica pod in Running phase with ready=true
- **ServiceReady**: Both AMQP and Management services exist with endpoints
- **AMQPReady**: TCP connection successful to AMQP port (5672)
- **ManagementReady**: HTTP request successful to Management UI (15672/api/overview)

### 7. RabbitMQConnectionInfo Entity

**Purpose**: Provides connection details for clients to connect to RabbitMQ

**Structure**:
```go
type RabbitMQConnectionInfo struct {
    // AMQPHost is the host for AMQP connections
    AMQPHost string `json:"amqpHost"`
    
    // AMQPPort is the port for AMQP connections
    AMQPPort int `json:"amqpPort"`
    
    // ManagementHost is the host for Management UI
    ManagementHost string `json:"managementHost"`
    
    // ManagementPort is the port for Management UI
    ManagementPort int `json:"managementPort"`
    
    // VirtualHost is the default virtual host
    VirtualHost string `json:"virtualHost"`
    
    // Username is the admin username
    Username string `json:"username"`
    
    // AMQPURL is the full AMQP connection URL
    AMQPURL string `json:"amqpUrl"`
    
    // ManagementURL is the full Management UI URL
    ManagementURL string `json:"managementUrl"`
}
```

**URL Format**:
- **AMQP URL**: `amqp://user:password@localhost:5672/vhost`
- **Management URL**: `http://localhost:15672`

**Example Values**:
```json
{
    "amqpHost": "localhost",
    "amqpPort": 5672,
    "managementHost": "localhost",
    "managementPort": 15672,
    "virtualHost": "/",
    "username": "user",
    "amqpUrl": "amqp://user:***@localhost:5672/",
    "managementUrl": "http://localhost:15672"
}
```

**Display Rules**:
- Password masked in `amqpUrl` (shown as `***`)
- Full credentials available in Kubernetes secret

### 8. RabbitMQHealthCheck Entity

**Purpose**: Comprehensive health check results for monitoring

**Structure**:
```go
type RabbitMQHealthCheck struct {
    // Timestamp of health check execution
    Timestamp time.Time `json:"timestamp"`
    
    // PodReady indicates pod health
    PodReady bool `json:"podReady"`
    
    // ServiceReady indicates services availability
    ServiceReady bool `json:"serviceReady"`
    
    // AMQPReady indicates AMQP protocol availability
    AMQPReady bool `json:"amqpReady"`
    
    // ManagementReady indicates Management API availability
    ManagementReady bool `json:"managementReady"`
    
    // ErrorMessage contains failure details
    ErrorMessage string `json:"errorMessage,omitempty"`
    
    // Uptime is the RabbitMQ server uptime string
    Uptime string `json:"uptime,omitempty"`
    
    // NodeInfo contains RabbitMQ node information
    NodeInfo *RabbitMQNodeInfo `json:"nodeInfo,omitempty"`
}
```

### 9. RabbitMQNodeInfo Entity

**Purpose**: Detailed RabbitMQ node information from Management API

**Structure**:
```go
type RabbitMQNodeInfo struct {
    // Name is the RabbitMQ node name
    Name string `json:"name"`
    
    // Running indicates if node is running
    Running bool `json:"running"`
    
    // MemoryUsed is current memory usage in bytes
    MemoryUsed int64 `json:"memoryUsed"`
    
    // MemoryLimit is memory limit in bytes
    MemoryLimit int64 `json:"memoryLimit"`
    
    // DiskFree is free disk space in bytes
    DiskFree int64 `json:"diskFree"`
    
    // FDUsed is number of file descriptors used
    FDUsed int `json:"fdUsed"`
    
    // FDTotal is total file descriptors available
    FDTotal int `json:"fdTotal"`
    
    // SocketsUsed is number of sockets used
    SocketsUsed int `json:"socketsUsed"`
    
    // SocketsTotal is total sockets available
    SocketsTotal int `json:"socketsTotal"`
}
```

## Entity Relationships

### Configuration Flow
```
kindenv.yaml (User Input)
    ↓ parsed by
KindEnvConfig.LoadConfig()
    ↓ validates
RabbitMQConfig.Validate()
    ↓ creates
Kubernetes Secret (rabbitmq-credentials)
    ↓ installs
Helm Chart (bitnami/rabbitmq)
    ↓ creates
Kubernetes Resources (StatefulSet, Services, ConfigMap)
```

### Status Reporting Flow
```
User Command (kindenv status)
    ↓ queries
RabbitMQStatusReporter.GetStatus()
    ↓ checks
Pod Status (Kubernetes API)
Service Status (Kubernetes API)
AMQP Connectivity (TCP socket)
Management API (HTTP request)
    ↓ returns
RabbitMQStatus (aggregated status)
    ↓ displays
CLI Output (formatted table)
```

## Data Persistence

### Persistent Data (when enabled)
- **Location**: PersistentVolumeClaim → PersistentVolume → Host path (Kind)
- **Contents**:
  - RabbitMQ database files (Mnesia)
  - Message queue data
  - Durable queue definitions
  - Exchange definitions
  - Binding definitions
  - User and permission data
  - Virtual host configurations

### Ephemeral Data (always)
- **Location**: emptyDir volume
- **Contents**:
  - Log files
  - Temporary files
  - PID files
  - Socket files

### Secrets Data
- **Location**: Kubernetes etcd via Secret resource
- **Contents**:
  - Username (base64 encoded)
  - Password (base64 encoded)
  - Erlang Cookie (base64 encoded)

## Validation Rules Summary

| Entity | Field | Validation Rule | Error Message |
|--------|-------|----------------|---------------|
| RabbitMQConfig | ChartVersion | Non-empty, semantic version | "chart version must be valid semantic version" |
| RabbitMQConfig | VirtualHost | Starts with "/" or alphanumeric | "virtual host must start with / or be alphanumeric" |
| RabbitMQNodePorts | AMQP | Range 30000-32767 | "AMQP nodeport must be in range 30000-32767" |
| RabbitMQNodePorts | Management | Range 30000-32767 | "Management nodeport must be in range 30000-32767" |
| RabbitMQResources | CPU | Pattern `^[0-9]+m?$` | "CPU must be in valid format (e.g., 500m, 1)" |
| RabbitMQResources | Memory | Pattern `^[0-9]+[KMGT]i$` | "Memory must be in valid format (e.g., 1Gi, 512Mi)" |
| RabbitMQPersistence | Size | Pattern `^[0-9]+[KMGT]i$` when enabled | "Persistence size must be in valid format" |
| RabbitMQSecret | Username | Non-empty when enabled | "Username must be specified" |
| RabbitMQSecret | Password | Non-empty when enabled | "Password must be specified" |
| RabbitMQSecret | ErlangCookie | 20+ chars if provided | "Erlang cookie must be at least 20 characters" |

## Error Handling

### Custom Error Types

```go
// RabbitMQError represents RabbitMQ operation errors
type RabbitMQError struct {
    Operation string
    Reason    string
    Err       error
}

// ValidationError represents configuration validation errors
type ValidationError struct {
    Field  string
    Value  string
    Reason string
}
```

### Error Scenarios

1. **Configuration Validation**:
   - Invalid port numbers → `ValidationError`
   - Invalid resource format → `ValidationError`
   - Missing required fields → `ValidationError`

2. **Installation**:
   - Helm chart not found → `RabbitMQError{Operation: "install", Reason: "chart not found"}`
   - Insufficient cluster resources → `RabbitMQError{Operation: "install", Reason: "insufficient resources"}`
   - Port conflict → `RabbitMQError{Operation: "install", Reason: "port already in use"}`

3. **Runtime**:
   - Pod crash loop → `RabbitMQError{Operation: "health_check", Reason: "pod not ready"}`
   - Connection timeout → `RabbitMQError{Operation: "connect", Reason: "connection timeout"}`
   - Authentication failure → `RabbitMQError{Operation: "auth", Reason: "invalid credentials"}`

## State Machine

### RabbitMQ Deployment State Machine

```
┌─────────┐
│ Pending │ Initial state when config loaded
└────┬────┘
     │
     │ Install() called
     ▼
┌──────────────┐
│ Installing   │ Helm install in progress
└────┬─────────┘
     │
     ├─ Success: WaitForReady() completes
     │   ▼
     │  ┌─────────┐
     │  │ Running │ All health checks pass
     │  └─────────┘
     │       │
     │       │ Uninstall() called
     │       ▼
     │  ┌─────────┐
     │  │ Stopped │ Helm release uninstalled
     │  └─────────┘
     │
     └─ Failure: Timeout or error
         ▼
     ┌────────┐
     │ Failed │ Installation or runtime failure
     └────────┘
```

## Data Model Summary

This data model provides:

1. **Clear Configuration Structure**: Well-defined YAML schema for user configuration
2. **Type Safety**: Strong typing in Go with validation
3. **Status Monitoring**: Comprehensive status entities for health checks
4. **Error Handling**: Custom error types for different failure scenarios
5. **Flexibility**: Support for both persistent and ephemeral deployments
6. **Security**: Proper secret management with Kubernetes secrets
7. **Consistency**: Follows established MySQL integration pattern

All entities are designed to integrate seamlessly with existing kindenv components while providing RabbitMQ-specific functionality.
