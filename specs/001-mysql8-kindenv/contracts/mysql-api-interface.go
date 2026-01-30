// MySQL API Interface Contract for KindEnv
// This file defines the interface contracts for MySQL integration

package kindenv

import (
	"context"
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

// HelmManager defines interface for Helm operations
type HelmManager interface {
	// AddRepository adds Bitnami Helm repository if not exists
	AddRepository(ctx context.Context, name, url string) error
	
	// UpdateRepository updates Helm repository cache
	UpdateRepository(ctx context.Context) error
	
	// InstallChart installs Helm chart with specified parameters
	InstallChart(ctx context.Context, params HelmInstallParams) error
	
	// UninstallChart removes Helm release
	UninstallChart(ctx context.Context, releaseName, namespace string) error
	
	// GetReleaseStatus returns Helm release status
	GetReleaseStatus(ctx context.Context, releaseName, namespace string) (*ReleaseStatus, error)
}

// KubernetesManager defines interface for Kubernetes operations
type KubernetesManager interface {
	// CreateNamespace creates Kubernetes namespace if not exists
	CreateNamespace(ctx context.Context, namespace string) error
	
	// DeleteNamespace deletes Kubernetes namespace
	DeleteNamespace(ctx context.Context, namespace string) error
	
	// CreateSecret creates Kubernetes secret with MySQL credentials
	CreateSecret(ctx context.Context, secretConfig MySQLSecretConfig) error
	
	// DeleteSecret deletes Kubernetes secret
	DeleteSecret(ctx context.Context, name, namespace string) error
	
	// GetPod returns pod information
	GetPod(ctx context.Context, name, namespace string) (*PodInfo, error)
	
	// GetService returns service information
	GetService(ctx context.Context, name, namespace string) (*ServiceInfo, error)
}

// Data Transfer Objects

// MySQLConfig represents MySQL component configuration
type MySQLConfig struct {
	Enabled      bool              `yaml:"enabled"`
	ChartVersion string            `yaml:"chartVersion"`
	Database     string            `yaml:"database"`
	NodePorts    MySQLNodePorts    `yaml:"nodePorts"`
	Resources    MySQLResources    `yaml:"resources"`
	Persistence  MySQLPersistence  `yaml:"persistence"`
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

// MySQLSecretConfig represents secret configuration
type MySQLSecretConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
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

// HelmInstallParams represents Helm installation parameters
type HelmInstallParams struct {
	ReleaseName  string            `json:"releaseName"`
	ChartName    string            `json:"chartName"`
	ChartVersion string            `json:"chartVersion"`
	Namespace    string            `json:"namespace"`
	Values       map[string]string `json:"values"`
	CreateNamespace bool           `json:"createNamespace"`
}

// ReleaseStatus represents Helm release status
type ReleaseStatus struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Status    string    `json:"status"`
	Version   int       `json:"version"`
	Updated   time.Time `json:"updated"`
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

// PodInfo represents detailed pod information
type PodInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Phase     string            `json:"phase"`
	Ready     bool              `json:"ready"`
	Labels    map[string]string `json:"labels"`
	StartTime time.Time         `json:"startTime,omitempty"`
}

// ServiceInfo represents detailed service information
type ServiceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	NodePort  int               `json:"nodePort,omitempty"`
	Labels    map[string]string `json:"labels"`
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

// Validation errors
type ValidationError struct {
	Field   string
	Value   string
	Reason  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%s': %s", e.Field, e.Value, e.Reason)
}