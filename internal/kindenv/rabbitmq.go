package kindenv

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

// RabbitMQConfig represents RabbitMQ component configuration
type RabbitMQConfig struct {
	Enabled      bool                 `yaml:"enabled"`
	ChartVersion string               `yaml:"chartVersion"`
	VirtualHost  string               `yaml:"virtualHost"`
	NodePorts    RabbitMQNodePorts    `yaml:"nodePorts"`
	Resources    RabbitMQResources    `yaml:"resources"`
	Persistence  RabbitMQPersistence  `yaml:"persistence"`
}

// RabbitMQNodePorts represents NodePort configuration
type RabbitMQNodePorts struct {
	AMQP       int `yaml:"amqp"`
	Management int `yaml:"management"`
}

// RabbitMQResources represents resource limits
type RabbitMQResources struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// RabbitMQPersistence represents persistence configuration
type RabbitMQPersistence struct {
	Enabled bool   `yaml:"enabled"`
	Size    string `yaml:"size"`
}

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
	if err := ValidateChartVersion(config.ChartVersion); err != nil {
		return err
	}

	// Validate virtual host format (must start with "/" or be alphanumeric)
	if config.VirtualHost != "" {
		if err := ValidateVirtualHost(config.VirtualHost); err != nil {
			return err
		}
	}

	// Validate NodePorts
	if err := ValidateRabbitMQNodePorts(config.NodePorts.AMQP, config.NodePorts.Management); err != nil {
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
// Virtual hosts must start with "/" or be alphanumeric strings starting with letter/underscore
func ValidateVirtualHost(vhost string) error {
	if vhost == "" {
		return nil // Empty virtual host defaults to "/" which is valid
	}
	// Virtual host pattern: starts with "/" followed by alphanumeric/underscore, or starts with letter/underscore followed by alphanumeric/underscore
	vhostRegex := regexp.MustCompile(`^(/[a-zA-Z0-9_]*|[a-zA-Z_][a-zA-Z0-9_]*)$`)
	if !vhostRegex.MatchString(vhost) {
		return &ValidationError{
			Field:  "virtualHost",
			Value:  vhost,
			Reason: "must start with / or be alphanumeric",
		}
	}
	return nil
}

// ValidateRabbitMQNodePorts validates both AMQP and Management NodePort numbers
// NodePorts must be in range 30000-32767 and must be unique
func ValidateRabbitMQNodePorts(amqpPort, managementPort int) error {
	if amqpPort < 30000 || amqpPort > 32767 {
		return &ValidationError{
			Field:  "nodePorts.amqp",
			Value:  fmt.Sprintf("%d", amqpPort),
			Reason: "must be in range 30000-32767",
		}
	}
	if managementPort < 30000 || managementPort > 32767 {
		return &ValidationError{
			Field:  "nodePorts.management",
			Value:  fmt.Sprintf("%d", managementPort),
			Reason: "must be in range 30000-32767",
		}
	}
	if amqpPort == managementPort {
		return &ValidationError{
			Field:  "nodePorts",
			Value:  fmt.Sprintf("amqp=%d, management=%d", amqpPort, managementPort),
			Reason: "amqp and management nodeports must be different",
		}
	}
	return nil
}

// RabbitMQManager defines the interface for RabbitMQ lifecycle management
type RabbitMQManager interface {
	Install(ctx context.Context, config RabbitMQConfig) error
	Uninstall(ctx context.Context, namespace string) error
	GetStatus(ctx context.Context, namespace string) (*RabbitMQStatus, error)
	ValidateConfig(config RabbitMQConfig) error
	WaitForReady(ctx context.Context, namespace string, timeout time.Duration) error
}

// RabbitMQConfigValidator defines interface for configuration validation
type RabbitMQConfigValidator interface {
	ValidateVirtualHost(vhost string) error
	ValidateResources(cpu, memory string) error
	ValidateNodePorts(amqpPort, managementPort int) error
	ValidateChartVersion(version string) error
}

// RabbitMQStatusReporter defines interface for status reporting and health checks
type RabbitMQStatusReporter interface {
	GetPodStatus(ctx context.Context, namespace, podName string) (*PodStatus, error)
	GetServiceStatus(ctx context.Context, namespace, serviceName string) (*ServiceStatus, error)
	TestConnection(ctx context.Context, connectionInfo RabbitMQConnectionInfo) error
	GetHealthCheck(ctx context.Context, namespace string) (*RabbitMQHealthCheck, error)
}

// RabbitMQStatus represents RabbitMQ deployment runtime status
type RabbitMQStatus struct {
	State          RabbitMQState           `json:"state"`
	PodReady       bool                    `json:"podReady"`
	ServiceReady   bool                    `json:"serviceReady"`
	AMQPReady      bool                    `json:"amqpReady"`
	ManagementReady bool                   `json:"managementReady"`
	ConnectionInfo *RabbitMQConnectionInfo `json:"connectionInfo,omitempty"`
	ErrorMessage   string                 `json:"errorMessage,omitempty"`
	LastChecked    time.Time               `json:"lastChecked"`
}

// RabbitMQState represents RabbitMQ deployment state
type RabbitMQState string

const (
	RabbitMQStatePending    RabbitMQState = "Pending"
	RabbitMQStateInstalling RabbitMQState = "Installing"
	RabbitMQStateRunning    RabbitMQState = "Running"
	RabbitMQStateFailed     RabbitMQState = "Failed"
	RabbitMQStateStopped    RabbitMQState = "Stopped"
)

// RabbitMQConnectionInfo represents connection details for RabbitMQ clients
type RabbitMQConnectionInfo struct {
	AMQPHost       string `json:"amqpHost"`
	AMQPPort       int    `json:"amqpPort"`
	ManagementHost string `json:"managementHost"`
	ManagementPort int    `json:"managementPort"`
	VirtualHost    string `json:"virtualHost"`
	Username       string `json:"username"`
	AMQPURL        string `json:"amqpUrl"`
	ManagementURL  string `json:"managementUrl"`
}

// RabbitMQHealthCheck represents comprehensive health check results
type RabbitMQHealthCheck struct {
	Timestamp       time.Time        `json:"timestamp"`
	PodReady        bool             `json:"podReady"`
	ServiceReady    bool             `json:"serviceReady"`
	AMQPReady       bool             `json:"amqpReady"`
	ManagementReady bool             `json:"managementReady"`
	ErrorMessage    string           `json:"errorMessage,omitempty"`
	Uptime          string           `json:"uptime,omitempty"`
	NodeInfo        *RabbitMQNodeInfo `json:"nodeInfo,omitempty"`
}

// RabbitMQNodeInfo represents detailed RabbitMQ node information from Management API
type RabbitMQNodeInfo struct {
	Name         string `json:"name"`
	Running      bool   `json:"running"`
	MemoryUsed   int64  `json:"memoryUsed"`
	MemoryLimit  int64  `json:"memoryLimit"`
	DiskFree     int64  `json:"diskFree"`
	FDUsed       int    `json:"fdUsed"`
	FDTotal      int    `json:"fdTotal"`
	SocketsUsed  int    `json:"socketsUsed"`
	SocketsTotal int    `json:"socketsTotal"`
}

// RabbitMQError represents a RabbitMQ operation error
type RabbitMQError struct {
	Operation string
	Reason    string
	Err       error
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

// NewRabbitMQError creates a new RabbitMQ error
func NewRabbitMQError(operation, reason string, err error) *RabbitMQError {
	return &RabbitMQError{
		Operation: operation,
		Reason:    reason,
		Err:       err,
	}
}

// IsRabbitMQError checks if an error is a RabbitMQError
func IsRabbitMQError(err error) bool {
	_, ok := err.(*RabbitMQError)
	return ok
}
