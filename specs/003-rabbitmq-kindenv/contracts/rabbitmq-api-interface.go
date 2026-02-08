package kindenv

import (
	"context"
	"fmt"
	"time"
)

// RabbitMQManager defines the interface for RabbitMQ lifecycle management
// This interface provides methods for installing, managing, and monitoring RabbitMQ
// deployments in a Kind Kubernetes environment.
type RabbitMQManager interface {
	// Install deploys RabbitMQ using Bitnami Helm chart with provided configuration.
	// It creates necessary secrets, configures Helm values, and installs the chart.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - config: RabbitMQ configuration including chart version, resources, and persistence
	//
	// Returns:
	//   - error: Installation error or nil on success
	//
	// Errors:
	//   - RabbitMQError: When Helm installation fails
	//   - ValidationError: When configuration is invalid
	Install(ctx context.Context, config RabbitMQConfig) error

	// Uninstall removes RabbitMQ deployment and cleans up all associated resources.
	// This includes Helm release, secrets, configmaps, and optionally persistent volumes.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - namespace: Kubernetes namespace where RabbitMQ is deployed
	//
	// Returns:
	//   - error: Uninstallation error or nil on success
	//
	// Errors:
	//   - RabbitMQError: When Helm uninstall fails
	Uninstall(ctx context.Context, namespace string) error

	// GetStatus returns current RabbitMQ deployment status including pod health,
	// service availability, and connection information.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - namespace: Kubernetes namespace where RabbitMQ is deployed
	//
	// Returns:
	//   - *RabbitMQStatus: Current status information
	//   - error: Query error or nil on success
	//
	// Errors:
	//   - RabbitMQError: When status query fails
	GetStatus(ctx context.Context, namespace string) (*RabbitMQStatus, error)

	// ValidateConfig validates RabbitMQ configuration parameters before installation.
	// Checks virtual host format, resource specifications, port ranges, and chart version.
	//
	// Parameters:
	//   - config: RabbitMQ configuration to validate
	//
	// Returns:
	//   - error: Validation error or nil if config is valid
	//
	// Errors:
	//   - ValidationError: When validation fails with field and reason details
	ValidateConfig(config RabbitMQConfig) error

	// WaitForReady waits for RabbitMQ to be ready with specified timeout.
	// Polls pod status, service availability, and AMQP/Management connectivity.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - namespace: Kubernetes namespace where RabbitMQ is deployed
	//   - timeout: Maximum time to wait for ready state
	//
	// Returns:
	//   - error: Timeout or readiness check error, nil if ready
	//
	// Errors:
	//   - RabbitMQError: When readiness checks fail or timeout occurs
	WaitForReady(ctx context.Context, namespace string, timeout time.Duration) error
}

// RabbitMQConfigValidator defines interface for configuration validation
// This interface provides granular validation methods for different configuration aspects.
type RabbitMQConfigValidator interface {
	// ValidateVirtualHost validates RabbitMQ virtual host name format.
	// Virtual hosts must start with "/" or be alphanumeric strings.
	//
	// Parameters:
	//   - vhost: Virtual host name to validate
	//
	// Returns:
	//   - error: Validation error or nil if valid
	//
	// Errors:
	//   - ValidationError: When virtual host format is invalid
	ValidateVirtualHost(vhost string) error

	// ValidateResources validates CPU and memory resource specifications.
	// CPU must match pattern: ^[0-9]+m?$ (e.g., "500m", "1")
	// Memory must match pattern: ^[0-9]+[KMGT]i$ (e.g., "1Gi", "512Mi")
	//
	// Parameters:
	//   - cpu: CPU resource specification
	//   - memory: Memory resource specification
	//
	// Returns:
	//   - error: Validation error or nil if valid
	//
	// Errors:
	//   - ValidationError: When resource format is invalid
	ValidateResources(cpu, memory string) error

	// ValidateNodePorts validates both AMQP and Management NodePort numbers.
	// NodePorts must be in range 30000-32767 and must be unique.
	//
	// Parameters:
	//   - amqpPort: AMQP protocol NodePort
	//   - managementPort: Management UI NodePort
	//
	// Returns:
	//   - error: Validation error or nil if valid
	//
	// Errors:
	//   - ValidationError: When ports are out of range or duplicate
	ValidateNodePorts(amqpPort, managementPort int) error

	// ValidateChartVersion validates Helm chart version format.
	// Chart version must be non-empty and follow semantic versioning.
	//
	// Parameters:
	//   - version: Helm chart version string
	//
	// Returns:
	//   - error: Validation error or nil if valid
	//
	// Errors:
	//   - ValidationError: When version format is invalid
	ValidateChartVersion(version string) error
}

// RabbitMQStatusReporter defines interface for status reporting and health checks
// This interface provides methods for querying RabbitMQ deployment health and connectivity.
type RabbitMQStatusReporter interface {
	// GetPodStatus returns Kubernetes pod status information for RabbitMQ.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - namespace: Kubernetes namespace where pod is deployed
	//   - podName: Name of the RabbitMQ pod
	//
	// Returns:
	//   - *PodStatus: Pod status information
	//   - error: Query error or nil on success
	//
	// Errors:
	//   - RabbitMQError: When pod status query fails
	GetPodStatus(ctx context.Context, namespace, podName string) (*PodStatus, error)

	// GetServiceStatus returns Kubernetes service status information for RabbitMQ.
	// Checks both AMQP and Management UI services.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - namespace: Kubernetes namespace where service is deployed
	//   - serviceName: Name of the RabbitMQ service
	//
	// Returns:
	//   - *ServiceStatus: Service status information
	//   - error: Query error or nil on success
	//
	// Errors:
	//   - RabbitMQError: When service status query fails
	GetServiceStatus(ctx context.Context, namespace, serviceName string) (*ServiceStatus, error)

	// TestConnection tests RabbitMQ AMQP and Management API connectivity.
	// Verifies both protocol endpoints are accessible and responding.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - connectionInfo: Connection details (hosts, ports, credentials)
	//
	// Returns:
	//   - error: Connection test error or nil if successful
	//
	// Errors:
	//   - RabbitMQError: When connection test fails
	TestConnection(ctx context.Context, connectionInfo RabbitMQConnectionInfo) error

	// GetHealthCheck performs comprehensive health check including pod, service,
	// AMQP, and Management API availability, plus resource utilization metrics.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - namespace: Kubernetes namespace where RabbitMQ is deployed
	//
	// Returns:
	//   - *RabbitMQHealthCheck: Comprehensive health check results
	//   - error: Health check error or nil on success
	//
	// Errors:
	//   - RabbitMQError: When health check fails
	GetHealthCheck(ctx context.Context, namespace string) (*RabbitMQHealthCheck, error)
}

// RabbitMQConfig represents RabbitMQ component configuration
// This structure maps to the kindenv.yaml configuration file.
type RabbitMQConfig struct {
	// Enabled indicates whether RabbitMQ should be deployed
	Enabled bool `yaml:"enabled"`

	// Namespace is the Kubernetes namespace for RabbitMQ deployment
	Namespace string `yaml:"namespace"`

	// ChartVersion specifies the Bitnami RabbitMQ Helm chart version
	ChartVersion string `yaml:"chartVersion"`

	// VirtualHost is the default RabbitMQ virtual host (default: "/")
	VirtualHost string `yaml:"virtualHost"`

	// NodePorts defines port mappings for AMQP and Management UI
	NodePorts RabbitMQNodePorts `yaml:"nodePorts"`

	// Resources defines CPU, memory, and storage limits
	Resources RabbitMQResources `yaml:"resources"`

	// Persistence controls data persistence configuration
	Persistence RabbitMQPersistence `yaml:"persistence"`
}

// RabbitMQNodePorts represents NodePort configuration for RabbitMQ services
type RabbitMQNodePorts struct {
	// AMQP is the NodePort for AMQP protocol (default: 30672, maps to host port 5672)
	AMQP int `yaml:"amqp"`

	// Management is the NodePort for Management UI (default: 31672, maps to host port 15672)
	Management int `yaml:"management"`
}

// RabbitMQResources represents resource limits for RabbitMQ pods
type RabbitMQResources struct {
	// CPU resource specification (e.g., "500m", "1")
	CPU string `yaml:"cpu"`

	// Memory resource specification (e.g., "1Gi", "512Mi")
	Memory string `yaml:"memory"`
}

// RabbitMQPersistence represents persistence configuration
type RabbitMQPersistence struct {
	// Enabled indicates whether data persistence is enabled
	Enabled bool `yaml:"enabled"`

	// Size is the PersistentVolume size (e.g., "8Gi", "10Gi")
	Size string `yaml:"size"`
}

// RabbitMQStatus represents RabbitMQ deployment runtime status
type RabbitMQStatus struct {
	// State represents the current deployment state
	State RabbitMQState `json:"state"`

	// PodReady indicates if RabbitMQ pod is ready
	PodReady bool `json:"podReady"`

	// ServiceReady indicates if Kubernetes services are ready
	ServiceReady bool `json:"serviceReady"`

	// AMQPReady indicates if AMQP port (5672) is accessible
	AMQPReady bool `json:"amqpReady"`

	// ManagementReady indicates if Management UI (15672) is accessible
	ManagementReady bool `json:"managementReady"`

	// ConnectionInfo contains connection details for clients
	ConnectionInfo *RabbitMQConnectionInfo `json:"connectionInfo,omitempty"`

	// ErrorMessage contains error details if state is Failed
	ErrorMessage string `json:"errorMessage,omitempty"`

	// LastChecked is the timestamp of last status check
	LastChecked time.Time `json:"lastChecked"`
}

// RabbitMQState represents RabbitMQ deployment state
type RabbitMQState string

const (
	// RabbitMQStatePending indicates initial state when configuration is loaded
	RabbitMQStatePending RabbitMQState = "Pending"

	// RabbitMQStateInstalling indicates Helm installation is in progress
	RabbitMQStateInstalling RabbitMQState = "Installing"

	// RabbitMQStateRunning indicates RabbitMQ is fully operational
	RabbitMQStateRunning RabbitMQState = "Running"

	// RabbitMQStateFailed indicates installation or runtime failure
	RabbitMQStateFailed RabbitMQState = "Failed"

	// RabbitMQStateStopped indicates RabbitMQ was intentionally stopped
	RabbitMQStateStopped RabbitMQState = "Stopped"
)

// RabbitMQConnectionInfo represents connection details for RabbitMQ clients
type RabbitMQConnectionInfo struct {
	// AMQPHost is the host for AMQP connections (typically "localhost")
	AMQPHost string `json:"amqpHost"`

	// AMQPPort is the port for AMQP connections (typically 5672)
	AMQPPort int `json:"amqpPort"`

	// ManagementHost is the host for Management UI (typically "localhost")
	ManagementHost string `json:"managementHost"`

	// ManagementPort is the port for Management UI (typically 15672)
	ManagementPort int `json:"managementPort"`

	// VirtualHost is the default virtual host (typically "/")
	VirtualHost string `json:"virtualHost"`

	// Username is the admin username
	Username string `json:"username"`

	// AMQPURL is the full AMQP connection URL (password masked)
	AMQPURL string `json:"amqpUrl"`

	// ManagementURL is the full Management UI URL
	ManagementURL string `json:"managementUrl"`
}

// RabbitMQHealthCheck represents comprehensive health check results
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

// RabbitMQNodeInfo represents detailed RabbitMQ node information from Management API
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

// PodStatus represents Kubernetes pod status
type PodStatus struct {
	// Name is the pod name
	Name string `json:"name"`

	// Namespace is the pod namespace
	Namespace string `json:"namespace"`

	// Phase is the pod phase (Pending, Running, Succeeded, Failed, Unknown)
	Phase string `json:"phase"`

	// Ready indicates if pod is ready to serve requests
	Ready bool `json:"ready"`

	// StartTime is when the pod started
	StartTime time.Time `json:"startTime,omitempty"`
}

// ServiceStatus represents Kubernetes service status
type ServiceStatus struct {
	// Name is the service name
	Name string `json:"name"`

	// Namespace is the service namespace
	Namespace string `json:"namespace"`

	// Type is the service type (ClusterIP, NodePort, LoadBalancer)
	Type string `json:"type"`

	// NodePort is the NodePort number (for NodePort services)
	NodePort int `json:"nodePort,omitempty"`

	// Ready indicates if service has endpoints
	Ready bool `json:"ready"`
}

// Error types for RabbitMQ operations

// RabbitMQError represents a RabbitMQ operation error
type RabbitMQError struct {
	// Operation is the operation that failed (e.g., "install", "status", "connect")
	Operation string

	// Reason is a human-readable reason for the failure
	Reason string

	// Err is the underlying error (if any)
	Err error
}

func (e *RabbitMQError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("RabbitMQ %s failed: %s: %v", e.Operation, e.Reason, e.Err)
	}
	return fmt.Sprintf("RabbitMQ %s failed: %s", e.Operation, e.Reason)
}

func (e *RabbitMQError) Unwrap() error {
	return e.Err
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	// Field is the configuration field that failed validation
	Field string

	// Value is the invalid value
	Value string

	// Reason is a human-readable reason for the validation failure
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%s': %s", e.Field, e.Value, e.Reason)
}

// NewRabbitMQError creates a new RabbitMQ error
func NewRabbitMQError(operation, reason string, err error) *RabbitMQError {
	return &RabbitMQError{
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

// IsRabbitMQError checks if an error is a RabbitMQError
func IsRabbitMQError(err error) bool {
	_, ok := err.(*RabbitMQError)
	return ok
}
