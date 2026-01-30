package kindenv

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

// MySQLManager defines the interface for MySQL lifecycle management
type MySQLManager interface {
	// Install deploys MySQL using Helm chart with provided configuration
	Install(ctx context.Context, config MySQLConfig) error

	// Uninstall removes MySQL deployment and cleans up resources
	Uninstall(ctx context.Context, namespace string) error

	// GetStatus returns current MySQL deployment status and health
	GetStatus(ctx context.Context, namespace string) (*MySQLStatus, error)

	// ValidateConfig validates MySQL configuration parameters
	ValidateConfig(config MySQLConfig) error

	// WaitForReady waits for MySQL to be ready with timeout
	WaitForReady(ctx context.Context, namespace string, timeout time.Duration) error
}

// MySQLConfigValidator defines interface for configuration validation
type MySQLConfigValidator interface {
	// ValidateDatabase validates database name format
	ValidateDatabase(database string) error

	// ValidateResources validates CPU and memory resource specifications
	ValidateResources(cpu, memory string) error

	// ValidateNodePort validates NodePort is in valid range and available
	ValidateNodePort(port int) error

	// ValidateChartVersion validates Helm chart version format
	ValidateChartVersion(version string) error
}

// MySQLStatusReporter defines interface for status reporting
type MySQLStatusReporter interface {
	// GetPodStatus returns Kubernetes pod status information
	GetPodStatus(ctx context.Context, namespace, podName string) (*PodStatus, error)

	// GetServiceStatus returns Kubernetes service status information
	GetServiceStatus(ctx context.Context, namespace, serviceName string) (*ServiceStatus, error)

	// TestConnection tests MySQL database connectivity
	TestConnection(ctx context.Context, connectionInfo MySQLConnectionInfo) error

	// GetHealthCheck performs comprehensive health check
	GetHealthCheck(ctx context.Context, namespace string) (*MySQLHealthCheck, error)
}

// MySQLConfig represents MySQL component configuration
type MySQLConfig struct {
	Enabled      bool             `yaml:"enabled"`
	ChartVersion string           `yaml:"chartVersion"`
	Database     string           `yaml:"database"`
	NodePorts    MySQLNodePorts   `yaml:"nodePorts"`
	Resources    MySQLResources   `yaml:"resources"`
	Persistence  MySQLPersistence `yaml:"persistence"`
}

// MySQLNodePorts represents NodePort configuration
type MySQLNodePorts struct {
	MySQL int `yaml:"mysql"`
}

// MySQLResources represents resource limits
type MySQLResources struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// MySQLPersistence represents persistence configuration
type MySQLPersistence struct {
	Enabled bool   `yaml:"enabled"`
	Size    string `yaml:"size"`
}

// MySQLStatus represents MySQL deployment status
type MySQLStatus struct {
	State          MySQLState           `json:"state"`
	PodReady       bool                 `json:"podReady"`
	ServiceReady   bool                 `json:"serviceReady"`
	DatabaseReady  bool                 `json:"databaseReady"`
	ConnectionInfo *MySQLConnectionInfo `json:"connectionInfo,omitempty"`
	ErrorMessage   string               `json:"errorMessage,omitempty"`
	LastChecked    time.Time            `json:"lastChecked"`
}

// MySQLState represents MySQL deployment state
type MySQLState string

const (
	MySQLStatePending    MySQLState = "Pending"
	MySQLStateInstalling MySQLState = "Installing"
	MySQLStateRunning    MySQLState = "Running"
	MySQLStateFailed     MySQLState = "Failed"
	MySQLStateStopped    MySQLState = "Stopped"
)

// MySQLConnectionInfo represents connection details
type MySQLConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
}

// MySQLHealthCheck represents health check results
type MySQLHealthCheck struct {
	Timestamp     time.Time `json:"timestamp"`
	PodReady      bool      `json:"podReady"`
	ServiceReady  bool      `json:"serviceReady"`
	DatabaseReady bool      `json:"databaseReady"`
	ErrorMessage  string    `json:"errorMessage,omitempty"`
	Uptime        string    `json:"uptime,omitempty"`
}

// PodStatus represents Kubernetes pod status
type PodStatus struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Phase     string    `json:"phase"`
	Ready     bool      `json:"ready"`
	StartTime time.Time `json:"startTime,omitempty"`
}

// ServiceStatus represents Kubernetes service status
type ServiceStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	NodePort  int    `json:"nodePort,omitempty"`
	Ready     bool   `json:"ready"`
}

// Error types for MySQL operations
type MySQLError struct {
	Operation string
	Reason    string
	Err       error
}

func (e *MySQLError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("MySQL %s failed: %s: %v", e.Operation, e.Reason, e.Err)
	}
	return fmt.Sprintf("MySQL %s failed: %s", e.Operation, e.Reason)
}

func (e *MySQLError) Unwrap() error {
	return e.Err
}

// ValidationError represents configuration validation errors
type ValidationError struct {
	Field  string
	Value  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%s': %s", e.Field, e.Value, e.Reason)
}

// ValidateMySQLConfig validates MySQL configuration parameters
func ValidateMySQLConfig(config MySQLConfig) error {
	if !config.Enabled {
		return nil // Skip validation if MySQL is disabled
	}

	// Validate chart version
	if config.ChartVersion == "" {
		return &ValidationError{
			Field:  "chartVersion",
			Value:  "",
			Reason: "chart version must be specified when MySQL is enabled",
		}
	}

	// Validate database name
	if err := ValidateDatabase(config.Database); err != nil {
		return err
	}

	// Validate NodePort
	if err := ValidateNodePort(config.NodePorts.MySQL); err != nil {
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

// ValidateDatabase validates database name format
func ValidateDatabase(database string) error {
	if database == "" {
		return &ValidationError{
			Field:  "database",
			Value:  "",
			Reason: "database name must be specified when MySQL is enabled",
		}
	}
	// MySQL database name: alphanumeric, underscore, 1-64 chars, must start with letter or underscore
	dbRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)
	if !dbRegex.MatchString(database) {
		return &ValidationError{
			Field:  "database",
			Value:  database,
			Reason: "must be valid MySQL database name (alphanumeric, underscore, 1-64 chars, start with letter or underscore)",
		}
	}
	return nil
}

// ValidateResources validates CPU and memory resource specifications
func ValidateResources(cpu, memory string) error {
	if cpu != "" {
		cpuRegex := regexp.MustCompile(`^[0-9]+m?$`)
		if !cpuRegex.MatchString(cpu) {
			return &ValidationError{
				Field:  "resources.cpu",
				Value:  cpu,
				Reason: "must be in valid format (e.g., 500m, 1)",
			}
		}
	}
	if memory != "" {
		memoryRegex := regexp.MustCompile(`^[0-9]+[KMGT]i$`)
		if !memoryRegex.MatchString(memory) {
			return &ValidationError{
				Field:  "resources.memory",
				Value:  memory,
				Reason: "must be in valid format (e.g., 1Gi, 512Mi)",
			}
		}
	}
	return nil
}

// ValidateNodePort validates NodePort is in valid range
func ValidateNodePort(port int) error {
	if port < 30000 || port > 32767 {
		return &ValidationError{
			Field:  "nodePorts.mysql",
			Value:  fmt.Sprintf("%d", port),
			Reason: "must be in range 30000-32767",
		}
	}
	return nil
}

// ValidateChartVersion validates Helm chart version format
func ValidateChartVersion(version string) error {
	if version == "" {
		return &ValidationError{
			Field:  "chartVersion",
			Value:  "",
			Reason: "chart version must be specified",
		}
	}
	// Semantic version pattern: X.Y.Z
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+`)
	if !versionRegex.MatchString(version) {
		return &ValidationError{
			Field:  "chartVersion",
			Value:  version,
			Reason: "must be valid semantic version (e.g., 9.4.6)",
		}
	}
	return nil
}

// NewMySQLError creates a new MySQL error
func NewMySQLError(operation, reason string, err error) *MySQLError {
	return &MySQLError{
		Operation: operation,
		Reason:    reason,
		Err:       err,
	}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsMySQLError checks if an error is a MySQLError
func IsMySQLError(err error) bool {
	_, ok := err.(*MySQLError)
	return ok
}
