package kindenv

// This file defines the interfaces and contracts for custom component management in kindenv.
// These interfaces provide clear boundaries for testing, mocking, and future extensibility.

import (
	"context"
	"time"
)

// CustomComponentManager defines the interface for managing custom components in the Kind cluster
type CustomComponentManager interface {
	// Deploy deploys all enabled custom components to the cluster
	Deploy(ctx context.Context, components []CustomComponent) error

	// DeployComponent deploys a single custom component
	DeployComponent(ctx context.Context, component *CustomComponent) error

	// Remove removes all custom components from the cluster
	Remove(ctx context.Context, components []CustomComponent) error

	// RemoveComponent removes a single custom component
	RemoveComponent(ctx context.Context, component *CustomComponent) error

	// GetStatus returns the status of all custom components
	GetStatus(ctx context.Context, components []CustomComponent) ([]ComponentStatus, error)

	// GetComponentStatus returns the status of a single custom component
	GetComponentStatus(ctx context.Context, component *CustomComponent) (*ComponentStatus, error)

	// Validate validates custom component configurations before deployment
	Validate(ctx context.Context, components []CustomComponent) error

	// AssignPorts assigns NodePort and HostPort values to components that don't have them specified
	AssignPorts(components []CustomComponent, existingPorts map[int]bool) error
}

// ComponentStatus represents the runtime status of a deployed custom component
type ComponentStatus struct {
	// Identity
	Name      string
	Namespace string

	// Deployment Status
	Phase             ComponentPhase // Pending, Running, Failed, Unknown
	Ready             bool           // Whether all replicas are ready
	Replicas          int            // Desired replicas
	ReadyReplicas     int            // Number of ready replicas
	AvailableReplicas int            // Number of available replicas

	// Pod Status
	Pods []PodStatus

	// Service Status
	ServiceReady bool
	ServicePorts []ServicePort

	// Resource Status
	ResourcesAllocated bool
	CPUUsage           string // e.g., "250m"
	MemoryUsage        string // e.g., "512Mi"

	// Error Information
	Error   error
	Reason  string // Human-readable reason for current state
	Message string // Detailed message about current state

	// Timing
	CreatedAt  time.Time
	LastUpdate time.Time
}

// ComponentPhase represents the high-level state of a component
type ComponentPhase string

const (
	ComponentPhasePending  ComponentPhase = "Pending"
	ComponentPhaseRunning  ComponentPhase = "Running"
	ComponentPhaseFailed   ComponentPhase = "Failed"
	ComponentPhaseUnknown  ComponentPhase = "Unknown"
	ComponentPhaseDisabled ComponentPhase = "Disabled"
)

// PodStatus represents the status of a single pod
type PodStatus struct {
	Name   string
	Ready  bool
	Phase  string // Pending, Running, Succeeded, Failed, Unknown
	Reason string
	// Restarts int
	IP string
}

// ServicePort represents an exposed port
type ServicePort struct {
	Name          string
	ContainerPort int
	NodePort      int
	HostPort      int
	Protocol      string
}

// KubernetesClient defines the interface for Kubernetes operations
// This abstraction allows for testing without a real cluster
type KubernetesClient interface {
	// Deployment Operations
	CreateDeployment(ctx context.Context, deployment *DeploymentSpec) error
	UpdateDeployment(ctx context.Context, deployment *DeploymentSpec) error
	DeleteDeployment(ctx context.Context, namespace, name string) error
	GetDeployment(ctx context.Context, namespace, name string) (*DeploymentStatus, error)

	// Service Operations
	CreateService(ctx context.Context, service *ServiceSpec) error
	DeleteService(ctx context.Context, namespace, name string) error
	GetService(ctx context.Context, namespace, name string) (*ServiceStatus, error)

	// Namespace Operations
	CreateNamespace(ctx context.Context, name string) error
	NamespaceExists(ctx context.Context, name string) (bool, error)

	// Secret Operations
	SecretExists(ctx context.Context, namespace, name string) (bool, error)
	GetSecret(ctx context.Context, namespace, name string) (map[string][]byte, error)

	// Pod Operations
	ListPods(ctx context.Context, namespace string, labelSelector map[string]string) ([]PodInfo, error)
	GetPodLogs(ctx context.Context, namespace, podName string) (string, error)
}

// DeploymentSpec defines the specification for creating a Kubernetes Deployment
type DeploymentSpec struct {
	Name        string
	Namespace   string
	Replicas    int
	Labels      map[string]string
	Annotations map[string]string

	// Container Spec
	ContainerName   string
	Image           string
	Command         []string
	Args            []string
	Env             []EnvVarSpec
	Ports           []ContainerPort
	Resources       ResourceRequirementsSpec
	VolumeMounts    []VolumeMountSpec
	ImagePullPolicy string

	// Pod Spec
	Volumes          []VolumeSpec
	ImagePullSecrets []string
	RestartPolicy    string
}

// EnvVarSpec defines an environment variable for a container
type EnvVarSpec struct {
	Name      string
	Value     string
	ValueFrom *EnvVarSourceSpec
}

// EnvVarSourceSpec defines the source of an environment variable value
type EnvVarSourceSpec struct {
	SecretKeyRef *SecretKeySelectorSpec
}

// SecretKeySelectorSpec defines a reference to a secret key
type SecretKeySelectorSpec struct {
	Name string
	Key  string
}

// ContainerPort defines a port to expose from a container
type ContainerPort struct {
	Name          string
	ContainerPort int
	Protocol      string
}

// ResourceRequirementsSpec defines resource requirements for a container
type ResourceRequirementsSpec struct {
	Requests ResourceListSpec
	Limits   ResourceListSpec
}

// ResourceListSpec defines CPU and memory quantities
type ResourceListSpec struct {
	CPU    string
	Memory string
}

// DeploymentStatus represents the current state of a deployment
type DeploymentStatus struct {
	Name              string
	Namespace         string
	Replicas          int
	ReadyReplicas     int
	AvailableReplicas int
	Conditions        []DeploymentCondition
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// DeploymentCondition represents a deployment condition
type DeploymentCondition struct {
	Type    string // Available, Progressing, ReplicaFailure
	Status  string // True, False, Unknown
	Reason  string
	Message string
}

// ServiceSpec defines the specification for creating a Kubernetes Service
type ServiceSpec struct {
	Name        string
	Namespace   string
	Type        string // ClusterIP, NodePort, LoadBalancer
	Selector    map[string]string
	Ports       []ServicePortSpec
	Labels      map[string]string
	Annotations map[string]string
}

// ServicePortSpec defines a port exposed by a service
type ServicePortSpec struct {
	Name       string
	Protocol   string
	Port       int
	TargetPort int
	NodePort   int
}

// ServiceStatus represents the current state of a service
type ServiceStatus struct {
	Name      string
	Namespace string
	Type      string
	ClusterIP string
	Ports     []ServicePortSpec
	CreatedAt time.Time
}

// ConfigMapSpec defines the specification for creating a Kubernetes ConfigMap
type ConfigMapSpec struct {
	Name        string
	Namespace   string
	Data        map[string]string // key: filename, value: file contents
	Labels      map[string]string
	Annotations map[string]string
}

// VolumeSpec defines a volume for mounting in a pod
type VolumeSpec struct {
	Name      string
	ConfigMap *ConfigMapVolumeSource
	Secret    *SecretVolumeSource
}

// ConfigMapVolumeSource defines a ConfigMap volume source
type ConfigMapVolumeSource struct {
	Name        string
	DefaultMode int // File permissions (e.g., 0644)
}

// SecretVolumeSource defines a Secret volume source
type SecretVolumeSource struct {
	Name        string
	DefaultMode int
}

// VolumeMountSpec defines how a volume is mounted in a container
type VolumeMountSpec struct {
	Name      string
	MountPath string
	SubPath   string // For mounting individual files from ConfigMap
	ReadOnly  bool
}

// PodInfo represents information about a pod
type PodInfo struct {
	Name      string
	Namespace string
	Phase     string // Pending, Running, Succeeded, Failed, Unknown
	Ready     bool
	Restarts  int
	IP        string
	Node      string
	StartTime time.Time
	Reason    string
	Message   string
}

// Validator defines the interface for validating custom components
type Validator interface {
	// ValidateConfiguration validates the component configuration structure
	ValidateConfiguration(component *CustomComponent) error

	// ValidatePreDeployment validates the component before deployment
	// (checks cluster state, secrets, ports, etc.)
	ValidatePreDeployment(ctx context.Context, component *CustomComponent) error

	// ValidateImage validates the image format and accessibility
	ValidateImage(image string) error

	// ValidatePort validates port configuration and checks for conflicts
	ValidatePort(port PortMapping, usedPorts map[int]bool) error

	// ValidateResources validates resource specifications
	ValidateResources(resources *ResourceRequirements) error

	// ValidateSecretReference validates that a secret exists
	ValidateSecretReference(ctx context.Context, namespace, secretName, key string) error
}

// DeploymentOrchestrator defines the interface for orchestrating multi-component deployments
type DeploymentOrchestrator interface {
	// DeployAll deploys all components in the correct order with dependencies
	DeployAll(ctx context.Context, components []CustomComponent) error

	// WaitForReady waits for all components to become ready
	WaitForReady(ctx context.Context, components []CustomComponent, timeout time.Duration) error

	// Rollback rolls back a failed deployment
	Rollback(ctx context.Context, components []CustomComponent) error
}

// PortManager defines the interface for managing port assignments
type PortManager interface {
	// AllocateNodePort allocates an available NodePort in the range 30000-32767
	AllocateNodePort(usedPorts map[int]bool) (int, error)

	// ValidatePortConflicts checks for port conflicts with existing components
	ValidatePortConflicts(components []CustomComponent, existingPorts map[int]bool) error

	// GetUsedPorts returns a map of all currently used ports in the cluster
	GetUsedPorts(ctx context.Context) (map[int]bool, error)

	// UpdateKindClusterPorts updates Kind cluster port mappings
	UpdateKindClusterPorts(ctx context.Context, ports []PortMapping) error
}

// SecretManager defines the interface for managing Kubernetes secrets
type SecretManager interface {
	// CreateSecret creates a new secret
	CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error

	// GetSecret retrieves a secret
	GetSecret(ctx context.Context, namespace, name string) (map[string][]byte, error)

	// SecretExists checks if a secret exists
	SecretExists(ctx context.Context, namespace, name string) (bool, error)

	// DeleteSecret deletes a secret
	DeleteSecret(ctx context.Context, namespace, name string) error
}

// ConfigMapManager defines the interface for managing Kubernetes ConfigMaps
type ConfigMapManager interface {
	// CreateConfigMap creates a new ConfigMap
	CreateConfigMap(ctx context.Context, spec *ConfigMapSpec) error

	// UpdateConfigMap updates an existing ConfigMap
	UpdateConfigMap(ctx context.Context, spec *ConfigMapSpec) error

	// DeleteConfigMap deletes a ConfigMap
	DeleteConfigMap(ctx context.Context, namespace, name string) error

	// ConfigMapExists checks if a ConfigMap exists
	ConfigMapExists(ctx context.Context, namespace, name string) (bool, error)

	// GetConfigMap retrieves a ConfigMap
	GetConfigMap(ctx context.Context, namespace, name string) (*ConfigMapSpec, error)
}

// NamespaceManager defines the interface for managing Kubernetes namespaces
type NamespaceManager interface {
	// CreateNamespace creates a new namespace
	CreateNamespace(ctx context.Context, name string, labels map[string]string) error

	// DeleteNamespace deletes a namespace
	DeleteNamespace(ctx context.Context, name string) error

	// NamespaceExists checks if a namespace exists
	NamespaceExists(ctx context.Context, name string) (bool, error)

	// EnsureNamespace creates namespace if it doesn't exist
	EnsureNamespace(ctx context.Context, name string, labels map[string]string) error
}

// StatusReporter defines the interface for reporting component status
type StatusReporter interface {
	// FormatStatus formats component status for display
	FormatStatus(status ComponentStatus) string

	// FormatMultipleStatus formats multiple component statuses for display
	FormatMultipleStatus(statuses []ComponentStatus) string

	// PrintStatus prints component status to stdout
	PrintStatus(status ComponentStatus)

	// PrintMultipleStatus prints multiple component statuses to stdout
	PrintMultipleStatus(statuses []ComponentStatus)
}

// ConfigurationLoader defines the interface for loading custom component configuration
type ConfigurationLoader interface {
	// Load loads custom component configuration from kindenv.yaml
	Load(configPath string) ([]CustomComponent, error)

	// Validate validates the loaded configuration
	Validate(components []CustomComponent) error

	// ApplyDefaults applies default values to components
	ApplyDefaults(components []CustomComponent) error
}

// TemplateGenerator defines the interface for generating Kubernetes manifests
type TemplateGenerator interface {
	// GenerateDeployment generates a Deployment YAML from a CustomComponent
	GenerateDeployment(component *CustomComponent) (string, error)

	// GenerateService generates a Service YAML from a CustomComponent
	GenerateService(component *CustomComponent) (string, error)

	// GenerateAll generates all Kubernetes resources for a component
	GenerateAll(component *CustomComponent) ([]string, error)
}

// Error types for better error handling
type ValidationError struct {
	Component string
	Field     string
	Value     interface{}
	Message   string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type DeploymentError struct {
	Component   string
	Phase       string
	K8sError    error
	Suggestions []string
}

func (e *DeploymentError) Error() string {
	return e.K8sError.Error()
}

type ConflictError struct {
	Component     string
	ConflictType  string
	ConflictsWith string
}

func (e *ConflictError) Error() string {
	return "conflict detected"
}
