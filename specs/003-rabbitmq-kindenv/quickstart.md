# Quickstart Guide: RabbitMQ Support for KindEnv

**Feature**: `003-rabbitmq-kindenv`  
**Date**: 2026-02-05  
**Phase**: 1 (Design & Contracts)

## Overview

This quickstart guide provides a step-by-step implementation roadmap for adding RabbitMQ support to the kindenv command. It follows Test-Driven Development (TDD) principles and leverages the existing MySQL integration pattern.

## Prerequisites

- Go 1.21+ installed
- Access to the devhelper-cli repository
- Familiarity with existing MySQL integration (spec 001-mysql8-kindenv)
- Understanding of Kubernetes, Helm, and RabbitMQ concepts

## Implementation Roadmap

### Phase 1: Configuration Schema (TDD)

#### Step 1.1: Write Configuration Tests

**File**: `internal/kindenv/config_test.go`

```go
// Add these test cases to existing test file
func TestRabbitMQConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      RabbitMQConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid configuration",
			config: RabbitMQConfig{
				Enabled:      true,
				Namespace:    "rabbitmq",
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
				Resources: RabbitMQResources{
					CPU:    "500m",
					Memory: "1Gi",
				},
				Persistence: RabbitMQPersistence{
					Enabled: false,
					Size:    "8Gi",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid virtual host",
			config: RabbitMQConfig{
				Enabled:      true,
				VirtualHost:  "invalid-vhost", // Must start with /
				ChartVersion: "11.0.0",
				// ... other fields
			},
			wantErr:     true,
			errContains: "virtual host",
		},
		{
			name: "invalid AMQP port range",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				NodePorts: RabbitMQNodePorts{
					AMQP:       29999, // Below valid range
					Management: 31672,
				},
				// ... other fields
			},
			wantErr:     true,
			errContains: "AMQP nodeport must be in range 30000-32767",
		},
		// Add more test cases for:
		// - Duplicate AMQP and Management ports
		// - Invalid CPU format
		// - Invalid memory format
		// - Invalid persistence size when enabled
		// - Missing chart version
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRabbitMQConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

#### Step 1.2: Implement Configuration Structures

**File**: `internal/kindenv/config.go`

Add RabbitMQ configuration to the `Components` struct:

```go
type Components struct {
	// ... existing components (Temporal, Redis, MySQL, etc.)
	
	RabbitMQ struct {
		Enabled      bool   `yaml:"enabled"`
		Namespace    string `yaml:"namespace"`
		ChartVersion string `yaml:"chartVersion"`
		VirtualHost  string `yaml:"virtualHost"`
		NodePorts    struct {
			AMQP       int `yaml:"amqp"`
			Management int `yaml:"management"`
		} `yaml:"nodePorts"`
		Resources struct {
			CPU    string `yaml:"cpu"`
			Memory string `yaml:"memory"`
		} `yaml:"resources"`
		Persistence struct {
			Enabled bool   `yaml:"enabled"`
			Size    string `yaml:"size"`
		} `yaml:"persistence"`
	} `yaml:"rabbitmq"`
}
```

Add RabbitMQ secret configuration to the `Secrets` struct:

```go
type Secrets struct {
	// ... existing secrets (MySQL)
	
	RabbitMQ struct {
		Enabled      bool   `yaml:"enabled"`
		Name         string `yaml:"name"`
		Namespace    string `yaml:"namespace"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		ErlangCookie string `yaml:"erlangCookie"`
	} `yaml:"rabbitmq"`
}
```

#### Step 1.3: Implement Configuration Defaults

**File**: `internal/kindenv/config.go` → `LoadConfig()` function

Add RabbitMQ defaults:

```go
// Set RabbitMQ component defaults
config.Components.RabbitMQ.Enabled = false
config.Components.RabbitMQ.Namespace = "rabbitmq"
config.Components.RabbitMQ.ChartVersion = "11.0.0"
config.Components.RabbitMQ.VirtualHost = "/"
config.Components.RabbitMQ.NodePorts.AMQP = 30672
config.Components.RabbitMQ.NodePorts.Management = 31672
config.Components.RabbitMQ.Resources.CPU = "500m"
config.Components.RabbitMQ.Resources.Memory = "1Gi"
config.Components.RabbitMQ.Persistence.Enabled = false
config.Components.RabbitMQ.Persistence.Size = "8Gi"

// Set RabbitMQ secret defaults
config.Secrets.RabbitMQ.Enabled = true
config.Secrets.RabbitMQ.Name = "rabbitmq-credentials"
config.Secrets.RabbitMQ.Namespace = "rabbitmq"
config.Secrets.RabbitMQ.Username = "user"
config.Secrets.RabbitMQ.Password = "password"
config.Secrets.RabbitMQ.ErlangCookie = "" // Auto-generated
```

#### Step 1.4: Implement Port Mapping Support

**File**: `internal/kindenv/config.go`

Add RabbitMQ to `generateDefaultPortMappings()`:

```go
// Add RabbitMQ port mappings
if config.Components.RabbitMQ.Enabled {
	// AMQP protocol
	mappings = append(mappings, struct {
		ContainerPort interface{} `yaml:"containerPort"`
		HostPort      int         `yaml:"hostPort"`
		Protocol      string      `yaml:"protocol"`
	}{
		ContainerPort: "${{ components.rabbitmq.nodePorts.amqp }}",
		HostPort:      5672,
		Protocol:      "TCP",
	})
	
	// Management UI
	mappings = append(mappings, struct {
		ContainerPort interface{} `yaml:"containerPort"`
		HostPort      int         `yaml:"hostPort"`
		Protocol      string      `yaml:"protocol"`
	}{
		ContainerPort: "${{ components.rabbitmq.nodePorts.management }}",
		HostPort:      15672,
		Protocol:      "TCP",
	})
}
```

Add RabbitMQ to `processVariableSubstitutions()`:

```go
case "rabbitmq":
	switch portName {
	case "amqp":
		value = config.Components.RabbitMQ.NodePorts.AMQP
	case "management":
		value = config.Components.RabbitMQ.NodePorts.Management
	default:
		return fmt.Errorf("unknown rabbitmq port: %s", portName)
	}
```

#### Step 1.5: Run Configuration Tests

```bash
go test ./internal/kindenv -run TestRabbitMQConfigValidation -v
```

**Expected**: Tests should fail (RED phase - no implementation yet)

### Phase 2: RabbitMQ Manager Implementation (TDD)

#### Step 2.1: Create RabbitMQ Manager Interface and Types

**File**: `internal/kindenv/rabbitmq.go` (NEW FILE)

Copy the interface definitions from `contracts/rabbitmq-api-interface.go` and implement validation functions:

```go
package kindenv

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

// Copy all interface definitions from contracts/rabbitmq-api-interface.go
// (RabbitMQManager, RabbitMQConfigValidator, RabbitMQStatusReporter)
// (All type definitions: RabbitMQConfig, RabbitMQStatus, etc.)
// (All error types: RabbitMQError, ValidationError)

// ValidateRabbitMQConfig validates RabbitMQ configuration parameters
func ValidateRabbitMQConfig(config RabbitMQConfig) error {
	if !config.Enabled {
		return nil // Skip validation if RabbitMQ is disabled
	}

	// Validate chart version
	if config.ChartVersion == "" {
		return &ValidationError{
			Field:  "chartVersion",
			Value:  "",
			Reason: "chart version must be specified when RabbitMQ is enabled",
		}
	}

	// Validate virtual host
	if err := ValidateVirtualHost(config.VirtualHost); err != nil {
		return err
	}

	// Validate NodePorts
	if err := ValidateNodePorts(config.NodePorts.AMQP, config.NodePorts.Management); err != nil {
		return err
	}

	// Validate resources
	if err := ValidateResources(config.Resources.CPU, config.Resources.Memory); err != nil {
		return err
	}

	// Validate persistence if enabled
	if config.Persistence.Enabled {
		if config.Persistence.Size == "" {
			return &ValidationError{
				Field:  "persistence.size",
				Value:  "",
				Reason: "persistence size must be specified when persistence is enabled",
			}
		}
		persistenceSizeRegex := regexp.MustCompile(`^[0-9]+[KMGT]i$`)
		if !persistenceSizeRegex.MatchString(config.Persistence.Size) {
			return &ValidationError{
				Field:  "persistence.size",
				Value:  config.Persistence.Size,
				Reason: "must be in valid format (e.g., 8Gi, 10Gi)",
			}
		}
	}

	return nil
}

// ValidateVirtualHost validates RabbitMQ virtual host name format
func ValidateVirtualHost(vhost string) error {
	if vhost == "" {
		vhost = "/" // Default virtual host
	}
	// Virtual host must start with "/" or be alphanumeric
	vhostRegex := regexp.MustCompile(`^/|^[a-zA-Z0-9_-]+$`)
	if !vhostRegex.MatchString(vhost) {
		return &ValidationError{
			Field:  "virtualHost",
			Value:  vhost,
			Reason: "must start with / or be alphanumeric (e.g., '/', '/dev', 'app1')",
		}
	}
	return nil
}

// ValidateNodePorts validates both AMQP and Management NodePort numbers
func ValidateNodePorts(amqpPort, managementPort int) error {
	// Validate AMQP port range
	if amqpPort < 30000 || amqpPort > 32767 {
		return &ValidationError{
			Field:  "nodePorts.amqp",
			Value:  fmt.Sprintf("%d", amqpPort),
			Reason: "must be in range 30000-32767",
		}
	}
	
	// Validate Management port range
	if managementPort < 30000 || managementPort > 32767 {
		return &ValidationError{
			Field:  "nodePorts.management",
			Value:  fmt.Sprintf("%d", managementPort),
			Reason: "must be in range 30000-32767",
		}
	}
	
	// Ensure ports are different
	if amqpPort == managementPort {
		return &ValidationError{
			Field:  "nodePorts",
			Value:  fmt.Sprintf("amqp=%d, management=%d", amqpPort, managementPort),
			Reason: "AMQP and Management ports must be different",
		}
	}
	
	return nil
}

// ValidateResources validates CPU and memory resource specifications
// (Copy from mysql.go and adapt)
func ValidateResources(cpu, memory string) error {
	// ... implementation
}

// ValidateChartVersion validates Helm chart version format
// (Copy from mysql.go and adapt)
func ValidateChartVersion(version string) error {
	// ... implementation
}
```

#### Step 2.2: Write RabbitMQ Manager Tests

**File**: `internal/kindenv/rabbitmq_test.go` (NEW FILE)

```go
package kindenv

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidateVirtualHost(t *testing.T) {
	tests := []struct {
		name    string
		vhost   string
		wantErr bool
	}{
		{"default root", "/", false},
		{"named vhost", "/dev", false},
		{"alphanumeric", "app1", false},
		{"with dash", "my-app", false},
		{"with underscore", "my_app", false},
		{"invalid special chars", "app@123", true},
		{"invalid space", "my app", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVirtualHost(tt.vhost)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNodePorts(t *testing.T) {
	tests := []struct {
		name           string
		amqpPort       int
		managementPort int
		wantErr        bool
		errContains    string
	}{
		{"valid ports", 30672, 31672, false, ""},
		{"AMQP below range", 29999, 31672, true, "AMQP"},
		{"AMQP above range", 32768, 31672, true, "AMQP"},
		{"Management below range", 30672, 29999, true, "Management"},
		{"Management above range", 30672, 32768, true, "Management"},
		{"duplicate ports", 30672, 30672, true, "must be different"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNodePorts(tt.amqpPort, tt.managementPort)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

#### Step 2.3: Run Tests (RED Phase)

```bash
go test ./internal/kindenv -run TestValidateVirtualHost -v
go test ./internal/kindenv -run TestValidateNodePorts -v
```

#### Step 2.4: Implement Validation Functions (GREEN Phase)

Implement all validation functions in `rabbitmq.go` until tests pass.

### Phase 3: Kubernetes Integration

#### Step 3.1: Update kindenv_start.go

**File**: `cmd/kindenv_start.go`

Add RabbitMQ installation logic similar to MySQL:

```go
// Add after MySQL installation section

// Install RabbitMQ if enabled
if config.Components.RabbitMQ.Enabled {
	fmt.Println(cyan("📦 Installing RabbitMQ..."))
	
	// Validate RabbitMQ configuration
	if err := kindenv.ValidateRabbitMQConfig(/* convert config */); err != nil {
		return fmt.Errorf("RabbitMQ configuration validation failed: %w", err)
	}
	
	// Create namespace if not exists
	if err := createNamespace(config.Components.RabbitMQ.Namespace); err != nil {
		return fmt.Errorf("failed to create RabbitMQ namespace: %w", err)
	}
	
	// Create Kubernetes secret for RabbitMQ credentials
	if err := createRabbitMQSecret(config); err != nil {
		return fmt.Errorf("failed to create RabbitMQ secret: %w", err)
	}
	
	// Install RabbitMQ Helm chart
	if err := installRabbitMQHelmChart(config); err != nil {
		return fmt.Errorf("failed to install RabbitMQ: %w", err)
	}
	
	fmt.Println(green("✅ RabbitMQ installed successfully"))
}
```

#### Step 3.2: Implement Helper Functions

**File**: `cmd/kindenv_start.go`

```go
func createRabbitMQSecret(config *kindenv.KindEnvConfig) error {
	// Generate erlang cookie if not provided
	erlangCookie := config.Secrets.RabbitMQ.ErlangCookie
	if erlangCookie == "" {
		erlangCookie = generateErlangCookie()
	}
	
	// Create secret using kubectl or Kubernetes client-go
	secretData := map[string]string{
		"rabbitmq-username":     config.Secrets.RabbitMQ.Username,
		"rabbitmq-password":     config.Secrets.RabbitMQ.Password,
		"rabbitmq-erlang-cookie": erlangCookie,
	}
	
	// Implementation using kubectl or client-go
	// ...
	
	return nil
}

func installRabbitMQHelmChart(config *kindenv.KindEnvConfig) error {
	// Build Helm values based on configuration
	helmValues := buildRabbitMQHelmValues(config)
	
	// Install using Helm Go SDK or helm command
	chartName := "bitnami/rabbitmq"
	releaseName := "rabbitmq"
	namespace := config.Components.RabbitMQ.Namespace
	
	// Implementation using Helm SDK
	// ...
	
	return nil
}

func buildRabbitMQHelmValues(config *kindenv.KindEnvConfig) map[string]interface{} {
	return map[string]interface{}{
		"global": map[string]interface{}{
			"imageRegistry": getImageRegistry(config),
		},
		"auth": map[string]interface{}{
			"username":     config.Secrets.RabbitMQ.Username,
			"password":     config.Secrets.RabbitMQ.Password,
			"erlangCookie": config.Secrets.RabbitMQ.ErlangCookie,
		},
		"replicaCount": 1,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    config.Components.RabbitMQ.Resources.CPU,
				"memory": config.Components.RabbitMQ.Resources.Memory,
			},
			"limits": map[string]interface{}{
				"cpu":    config.Components.RabbitMQ.Resources.CPU,
				"memory": config.Components.RabbitMQ.Resources.Memory,
			},
		},
		"persistence": map[string]interface{}{
			"enabled": config.Components.RabbitMQ.Persistence.Enabled,
			"size":    config.Components.RabbitMQ.Persistence.Size,
		},
		"service": map[string]interface{}{
			"type": "NodePort",
			"nodePorts": map[string]interface{}{
				"amqp":    config.Components.RabbitMQ.NodePorts.AMQP,
				"manager": config.Components.RabbitMQ.NodePorts.Management,
			},
		},
		"plugins": "rabbitmq_management rabbitmq_prometheus",
	}
}

func generateErlangCookie() string {
	// Generate cryptographically secure random string (20+ characters)
	// Use crypto/rand for secure random generation
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const cookieLength = 32
	
	// Implementation
	// ...
	
	return "generated-cookie-string"
}
```

### Phase 4: Status Reporting

#### Step 4.1: Update kindenv_status.go

**File**: `cmd/kindenv_status.go`

Add RabbitMQ status reporting:

```go
// Add after MySQL status section

// Display RabbitMQ status if enabled
if config.Components.RabbitMQ.Enabled {
	fmt.Println(cyan("\n📦 RabbitMQ:"))
	
	status := getRabbitMQStatus(config.Components.RabbitMQ.Namespace)
	
	if status.PodReady {
		fmt.Printf("  Status: %s\n", green("Running"))
		fmt.Printf("  AMQP URL: %s\n", status.ConnectionInfo.AMQPURL)
		fmt.Printf("  Management UI: %s\n", status.ConnectionInfo.ManagementURL)
		fmt.Printf("  Virtual Host: %s\n", status.ConnectionInfo.VirtualHost)
		fmt.Printf("  Username: %s\n", status.ConnectionInfo.Username)
	} else {
		fmt.Printf("  Status: %s\n", red(string(status.State)))
		if status.ErrorMessage != "" {
			fmt.Printf("  Error: %s\n", status.ErrorMessage)
		}
	}
}
```

#### Step 4.2: Implement Status Functions

```go
func getRabbitMQStatus(namespace string) *kindenv.RabbitMQStatus {
	status := &kindenv.RabbitMQStatus{
		State:       kindenv.RabbitMQStatePending,
		LastChecked: time.Now(),
	}
	
	// Check pod status
	podReady := checkPodReady(namespace, "rabbitmq-0")
	status.PodReady = podReady
	
	// Check service status
	serviceReady := checkServiceReady(namespace, "rabbitmq")
	status.ServiceReady = serviceReady
	
	// Test AMQP connectivity
	amqpReady := testAMQPConnection("localhost", 5672)
	status.AMQPReady = amqpReady
	
	// Test Management API
	managementReady := testManagementAPI("localhost", 15672)
	status.ManagementReady = managementReady
	
	// Determine overall state
	if podReady && serviceReady && amqpReady && managementReady {
		status.State = kindenv.RabbitMQStateRunning
		status.ConnectionInfo = &kindenv.RabbitMQConnectionInfo{
			AMQPHost:       "localhost",
			AMQPPort:       5672,
			ManagementHost: "localhost",
			ManagementPort: 15672,
			VirtualHost:    "/",
			Username:       "user",
			AMQPURL:        "amqp://user:***@localhost:5672/",
			ManagementURL:  "http://localhost:15672",
		}
	} else {
		status.State = kindenv.RabbitMQStateFailed
	}
	
	return status
}
```

### Phase 5: Testing Strategy

#### Integration Test Plan

**File**: `cmd/kindenv_start_test.go`

```go
func TestRabbitMQInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	// Setup: Create test Kind cluster
	// Execute: Install RabbitMQ via kindenv start
	// Verify: Pod running, services available, AMQP accessible
	// Cleanup: Uninstall RabbitMQ, delete cluster
}
```

#### Manual Test Checklist

1. **Configuration Loading**
   ```bash
   # Test with RabbitMQ enabled
   devhelper-cli kindenv start --config test-rabbitmq.yaml
   ```

2. **Installation Verification**
   ```bash
   kubectl get pods -n rabbitmq
   kubectl get svc -n rabbitmq
   kubectl logs -n rabbitmq rabbitmq-0
   ```

3. **Connectivity Testing**
   ```bash
   # AMQP connection
   telnet localhost 5672
   
   # Management UI
   curl http://localhost:15672/api/overview -u user:password
   open http://localhost:15672
   ```

4. **Status Command**
   ```bash
   devhelper-cli kindenv status
   ```

5. **Cleanup**
   ```bash
   devhelper-cli kindenv stop
   ```

### Phase 6: Documentation Updates

#### Update KINDENV.md

**File**: `docs/KINDENV.md`

Add RabbitMQ section with:
- Configuration examples
- Connection information
- Troubleshooting tips
- Management UI usage guide

#### Update kindenv.yaml Example

**File**: `examples/kindenv.yaml`

Add RabbitMQ configuration example.

## Common Patterns and Helpers

### Pattern 1: Copy from MySQL Implementation

For maximum consistency, follow these mappings:

| MySQL Component | RabbitMQ Equivalent |
|----------------|---------------------|
| `mysql.go` | `rabbitmq.go` |
| `MySQLConfig` | `RabbitMQConfig` |
| `MySQLStatus` | `RabbitMQStatus` |
| `MySQLManager` | `RabbitMQManager` |
| `ValidateMySQLConfig` | `ValidateRabbitMQConfig` |
| Port 3306 | Ports 5672 + 15672 |
| Single NodePort | Dual NodePorts |

### Pattern 2: Validation Function Template

```go
func Validate<Entity>(<params>) error {
	if <simple-check> {
		return &ValidationError{
			Field:  "<field-name>",
			Value:  "<value>",
			Reason: "<clear-error-message>",
		}
	}
	
	// Regex validation
	regex := regexp.MustCompile(`<pattern>`)
	if !regex.MatchString(<value>) {
		return &ValidationError{
			Field:  "<field-name>",
			Value:  <value>,
			Reason: "<format-example>",
		}
	}
	
	return nil
}
```

### Pattern 3: Table-Driven Tests

```go
tests := []struct {
	name        string
	input       <Type>
	wantErr     bool
	errContains string
}{
	{"valid case", <valid-input>, false, ""},
	{"invalid case", <invalid-input>, true, "<error-substring>"},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		err := <FunctionUnderTest>(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		} else {
			assert.NoError(t, err)
		}
	})
}
```

## Troubleshooting Guide

### Issue: Tests Failing

**Solution**: Run tests with verbose output:
```bash
go test ./internal/kindenv -v
go test ./cmd -v -run TestRabbitMQ
```

### Issue: Helm Chart Not Found

**Solution**: Ensure Bitnami repository is added:
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm search repo bitnami/rabbitmq
```

### Issue: Port Mapping Not Working

**Solution**: Recreate Kind cluster:
```bash
devhelper-cli kindenv stop
devhelper-cli kindenv start
```

### Issue: RabbitMQ Pod Not Starting

**Solution**: Check pod logs and events:
```bash
kubectl describe pod -n rabbitmq rabbitmq-0
kubectl logs -n rabbitmq rabbitmq-0
kubectl get events -n rabbitmq
```

## Success Criteria

- ✅ All unit tests passing (80%+ coverage)
- ✅ Configuration validation working correctly
- ✅ RabbitMQ installs successfully via `kindenv start`
- ✅ AMQP accessible on localhost:5672
- ✅ Management UI accessible on localhost:15672
- ✅ Status command shows accurate RabbitMQ status
- ✅ Cleanup works properly via `kindenv stop`
- ✅ Documentation updated with examples

## Next Steps

After completing this implementation:

1. **Run `/speckit.tasks`** to generate detailed task breakdown
2. **Create pull request** with implementation
3. **Add integration tests** to CI/CD pipeline
4. **Update user documentation** with RabbitMQ examples
5. **Consider future enhancements**: clustering, plugins, monitoring

## References

- [MySQL Implementation](../001-mysql8-kindenv/) - Reference pattern
- [RabbitMQ Helm Chart](https://github.com/bitnami/charts/tree/main/bitnami/rabbitmq)
- [Kind Port Mapping](https://kind.sigs.k8s.io/docs/user/configuration/#extra-port-mappings)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Assertions](https://pkg.go.dev/github.com/stretchr/testify/assert)
