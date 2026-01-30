# Quick Start: MySQL 8 Support for KindEnv

**Target Audience**: Developers implementing MySQL 8 integration  
**Prerequisites**: Familiarity with Go, Kubernetes, Helm, and existing kindenv codebase  
**Estimated Time**: 2-3 hours for basic implementation

## Implementation Overview

This guide walks through implementing MySQL 8 support in kindenv following the established patterns used by Redis and Temporal components.

## Step 1: Extend Configuration Structure

### 1.1 Update KindEnvConfig in `internal/kindenv/config.go`

Add MySQL configuration to the Components struct:

```go
// Add to existing Components struct
type Components struct {
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
}
```

### 1.2 Add Default Configuration in `CreateDefaultConfig()`

```go
// Add to CreateDefaultConfig() function
config.Components.MySQL.Enabled = false
config.Components.MySQL.ChartVersion = "9.4.6"
config.Components.MySQL.Database = "mysql"
config.Components.MySQL.NodePorts.MySQL = 30306
config.Components.MySQL.Resources.CPU = "500m"
config.Components.MySQL.Resources.Memory = "1Gi"
config.Components.MySQL.Persistence.Enabled = false
config.Components.MySQL.Persistence.Size = "8Gi"
```

### 1.3 Add Validation Logic

```go
// Add to Validate() method
if c.Components.MySQL.Enabled {
    if c.Components.MySQL.ChartVersion == "" {
        return errors.New("mysql chart version must be specified when enabled")
    }
    if c.Components.MySQL.Database == "" {
        return errors.New("mysql database name must be specified when enabled")
    }
    if c.Components.MySQL.NodePorts.MySQL < 30000 || c.Components.MySQL.NodePorts.MySQL > 32767 {
        return errors.New("mysql nodeport must be in range 30000-32767")
    }
    // Add more validation as needed
}
```

## Step 2: Extend Helm Repository Setup

### 2.1 Update `cmd/kindenv_init.go`

The Bitnami repository should already be added for Redis. Verify it's present in the Helm repository setup section.

## Step 3: Implement MySQL Installation Logic

### 3.1 Create MySQL Installation Function in `cmd/kindenv_start.go`

Add MySQL installation after Redis installation:

```go
// Install MySQL (add after Redis installation)
if config.Components.MySQL.Enabled {
    fmt.Println(yellow("Installing MySQL"))

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

    // Set up ECR credentials if needed
    if config.Images.UseAwsEcr {
        err = setupECRCreds("mysql", ecrRegistry, ecrPassword)
        if err != nil {
            fmt.Printf("%s Error setting up ECR credentials for MySQL: %v\n", red("❌"), err)
            os.Exit(1)
        }
    }

    // Install MySQL with Helm
    helmArgs := []string{
        "upgrade", "--install",
        "mysql", "bitnami/mysql",
        "--namespace", "mysql",
        "--version", config.Components.MySQL.ChartVersion,
        "--set", "primary.service.type=NodePort",
        "--set", fmt.Sprintf("primary.service.nodePorts.mysql=%d", config.Components.MySQL.NodePorts.MySQL),
        "--set", fmt.Sprintf("auth.database=%s", config.Components.MySQL.Database),
        "--set", fmt.Sprintf("primary.persistence.enabled=%t", config.Components.MySQL.Persistence.Enabled),
        "--set", fmt.Sprintf("primary.resources.requests.cpu=%s", config.Components.MySQL.Resources.CPU),
        "--set", fmt.Sprintf("primary.resources.requests.memory=%s", config.Components.MySQL.Resources.Memory),
        "--set", "secondary.replicaCount=0",
    }

    // Add secret configuration if MySQL secrets are enabled
    if config.Secrets.MySQL.Enabled {
        helmArgs = append(helmArgs, "--set", fmt.Sprintf("auth.existingSecret=%s", config.Secrets.MySQL.Name))
    } else {
        // Use default credentials
        helmArgs = append(helmArgs,
            "--set", "auth.rootPassword=password",
            "--set", fmt.Sprintf("auth.username=%s", "mysql"),
            "--set", "auth.password=password")
    }

    // ECR-specific image configuration
    if config.Images.UseAwsEcr {
        helmArgs = append(helmArgs,
            "--set", fmt.Sprintf("global.imageRegistry=%s", ecrRegistry),
            "--set", "image.repository=bitnamilegacy/mysql")
    }

    _, err = executeCommand("helm", helmArgs...)
    if err != nil {
        fmt.Printf("%s Error installing MySQL: %v\n", red("❌"), err)
        os.Exit(1)
    }

    // Wait for MySQL to be ready
    fmt.Println(yellow("Waiting for MySQL to be ready..."))
    time.Sleep(10 * time.Second)

    // Wait for MySQL pod
    fmt.Println(yellow("Waiting for mysql-primary-0 pod to be created..."))
    podCheckCmd := exec.Command("kubectl", "get", "pod", "mysql-primary-0", "-n", "mysql", "--no-headers")

    var podExists bool
    for i := 0; i < 10; i++ { // Try for up to 5 minutes (10 * 30s)
        podOutput, err := podCheckCmd.CombinedOutput()
        if err == nil && len(podOutput) > 0 {
            podExists = true
            if verbose {
                fmt.Println(string(podOutput))
            }
            break
        }
        if i < 9 {
            fmt.Printf("Waiting for mysql-primary-0 pod to appear (attempt %d/10)...\n", i+1)
            time.Sleep(30 * time.Second)
        }
    }

    if podExists {
        fmt.Printf("%s MySQL installed successfully\n", green("✅"))
    } else {
        fmt.Printf("%s MySQL installation may have issues - check with 'kubectl get pods -n mysql'\n", yellow("⚠️"))
    }
}
```

## Step 4: Extend Status Monitoring

### 4.1 Update `cmd/kindenv_status.go`

Add MySQL status checking after existing component checks:

```go
// Add MySQL status check (after Redis status)
if config.Components.MySQL.Enabled {
    fmt.Printf("  MySQL:\n")
    
    // Check MySQL pod status
    podCmd := exec.Command("kubectl", "get", "pod", "mysql-primary-0", "-n", "mysql", "--no-headers")
    podOutput, err := podCmd.CombinedOutput()
    
    if err != nil {
        fmt.Printf("    Status: %s (Pod not found)\n", red("❌ Not Running"))
    } else {
        podStatus := strings.Fields(string(podOutput))
        if len(podStatus) >= 3 {
            status := podStatus[2] // READY column
            if strings.Contains(status, "1/1") {
                fmt.Printf("    Status: %s\n", green("✅ Running"))
                
                // Show connection info
                fmt.Printf("    Connection: localhost:%d\n", config.Components.MySQL.NodePorts.MySQL)
                fmt.Printf("    Database: %s\n", config.Components.MySQL.Database)
                if config.Secrets.MySQL.Enabled {
                    fmt.Printf("    Username: %s\n", config.Secrets.MySQL.Username)
                } else {
                    fmt.Printf("    Username: root\n")
                }
            } else {
                fmt.Printf("    Status: %s (Pod: %s)\n", yellow("⚠️ Starting"), status)
            }
        }
    }
}
```

## Step 5: Extend Cleanup Logic

### 5.1 Update `cmd/kindenv_stop.go`

Add MySQL cleanup logic:

```go
// Add MySQL cleanup (after existing component cleanup)
if config.Components.MySQL.Enabled {
    fmt.Println(yellow("Cleaning up MySQL"))
    
    // Uninstall Helm release
    _, err := executeCommand("helm", "uninstall", "mysql", "--namespace", "mysql")
    if err != nil {
        fmt.Printf("%s Warning: Failed to uninstall MySQL: %v\n", yellow("⚠️"), err)
    }
    
    // Delete namespace
    _, err = executeCommand("kubectl", "delete", "namespace", "mysql", "--ignore-not-found")
    if err != nil {
        fmt.Printf("%s Warning: Failed to delete MySQL namespace: %v\n", yellow("⚠️"), err)
    }
    
    fmt.Printf("%s MySQL cleanup completed\n", green("✅"))
}
```

## Step 6: Update Documentation

### 6.1 Update `KINDENV.md`

Add MySQL configuration example:

```yaml
components:
  mysql:
    enabled: true
    chartVersion: "9.4.6"
    database: "myapp"
    nodePorts:
      mysql: 30306
    resources:
      cpu: "500m"
      memory: "1Gi"
    persistence:
      enabled: false
      size: "8Gi"

secrets:
  mysql:
    enabled: true
    name: "mysql-credentials"
    namespace: "mysql"
    username: "root"
    password: "securepassword"
```

## Step 7: Add Tests

### 7.1 Create `cmd/kindenv_start_mysql_test.go`

```go
package cmd

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMySQLConfigValidation(t *testing.T) {
    tests := []struct {
        name        string
        config      MySQLConfig
        expectError bool
    }{
        {
            name: "valid config",
            config: MySQLConfig{
                Enabled:      true,
                ChartVersion: "9.4.6",
                Database:     "mysql",
                NodePorts:    MySQLNodePorts{MySQL: 30306},
                Resources:    MySQLResources{CPU: "500m", Memory: "1Gi"},
                Persistence:  MySQLPersistence{Enabled: false, Size: "8Gi"},
            },
            expectError: false,
        },
        {
            name: "invalid nodeport",
            config: MySQLConfig{
                Enabled:   true,
                NodePorts: MySQLNodePorts{MySQL: 25000}, // Invalid port
            },
            expectError: true,
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateMySQLConfig(tt.config)
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Step 8: Testing the Implementation

### 8.1 Manual Testing Steps

1. **Initialize kindenv with MySQL**:
   ```bash
   devhelper-cli kindenv init
   # Edit kindenv.yaml to enable MySQL
   devhelper-cli kindenv start
   ```

2. **Verify MySQL is running**:
   ```bash
   devhelper-cli kindenv status
   kubectl get pods -n mysql
   ```

3. **Test MySQL connection**:
   ```bash
   mysql -h localhost -P 3306 -u root -p
   ```

4. **Test cleanup**:
   ```bash
   devhelper-cli kindenv stop
   ```

### 8.2 Integration Testing

Create integration tests that:
- Start kindenv with MySQL enabled
- Verify MySQL pod is running
- Test database connection
- Verify cleanup works properly

## Step 9: Performance Considerations

### 9.1 Optimization Tips

- **Resource Limits**: Start with conservative defaults (500m CPU, 1Gi memory)
- **Startup Time**: MySQL takes 30-60 seconds to start, plan accordingly
- **Persistence**: Keep disabled by default for faster development cycles
- **Health Checks**: Use appropriate timeouts for MySQL startup probes

### 9.2 Troubleshooting Common Issues

- **Image Pull Errors**: Verify ECR credentials and bitnamilegacy image availability
- **Resource Constraints**: Check Kind cluster has sufficient resources
- **Port Conflicts**: Ensure NodePort 30306 is available
- **Startup Failures**: Check MySQL pod logs with `kubectl logs mysql-primary-0 -n mysql`

## Next Steps

1. Implement the basic functionality following this guide
2. Add comprehensive error handling and validation
3. Write unit and integration tests
4. Update documentation and examples
5. Test with different configurations and edge cases

## Reference Implementation

See the contracts and data model files for detailed interface definitions and configuration schemas that should guide the implementation.