package tw

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TemporalWorkerConfig defines the structure for the temporal worker configuration
type TemporalWorkerConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Enabled    bool   `yaml:"enabled"`
		WorkerType string `yaml:"workerType"`
		Image      struct {
			Registry string `yaml:"registry"`
			Tag      string `yaml:"tag"`
		} `yaml:"image"`
		Temporal struct {
			Namespace               string `yaml:"namespace"`
			QueueType               string `yaml:"queueType"`
			MaxConcurrentWorkflows  int    `yaml:"maxConcurrentWorkflows"`
			MaxConcurrentActivities int    `yaml:"maxConcurrentActivities"`
		} `yaml:"temporal"`
		Autoscaler struct {
			Enabled                bool   `yaml:"enabled"`
			IdleReplicaCount       int    `yaml:"idleReplicaCount"`
			MinReplicaCount        int    `yaml:"minReplicaCount"`
			MaxReplicaCount        int    `yaml:"maxReplicaCount"`
			QueueType              string `yaml:"queueType"`
			TargetQueueSize        int    `yaml:"targetQueueSize"`
			PollingIntervalSeconds int    `yaml:"pollingIntervalSeconds"`
			CooldownPeriodSeconds  int    `yaml:"cooldownPeriodSeconds"`
		} `yaml:"autoscaler"`
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
		WorkerSettings string `yaml:"workerSettings"`
	} `yaml:"spec"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(configPath string) (*TemporalWorkerConfig, error) {
	// Set defaults
	config := CreateDefaultConfig("")

	// If configPath is empty, check if tw.yaml exists in the current directory
	if configPath == "" {
		if _, err := os.Stat("tw.yaml"); err == nil {
			configPath = "tw.yaml"
		} else {
			// No config file provided or found, return defaults
			return config, nil
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// CreateDefaultConfig creates a default temporal worker configuration
// with sensible defaults for common fields
func CreateDefaultConfig(name string) *TemporalWorkerConfig {
	config := &TemporalWorkerConfig{}
	config.APIVersion = "orchestration.shieldfc.com/v1alpha1"
	config.Kind = "TemporalWorker"

	// Set project name from current directory if not provided
	if name == "" {
		// Get current working directory
		currentDir, err := os.Getwd()
		if err == nil {
			// Extract the base directory name
			name = filepath.Base(currentDir)
		} else {
			name = "temporal-worker" // Fallback
		}
	}

	config.Metadata.Name = name

	// Set default spec values
	config.Spec.Enabled = true
	config.Spec.WorkerType = name

	// Image defaults
	config.Spec.Image.Registry = "992979781608.dkr.ecr.eu-west-1.amazonaws.com"
	config.Spec.Image.Tag = "latest"

	// Temporal defaults
	config.Spec.Temporal.Namespace = "default"
	config.Spec.Temporal.QueueType = "workflow"
	config.Spec.Temporal.MaxConcurrentWorkflows = 3
	config.Spec.Temporal.MaxConcurrentActivities = 5

	// Autoscaler defaults
	config.Spec.Autoscaler.Enabled = true
	config.Spec.Autoscaler.IdleReplicaCount = 1
	config.Spec.Autoscaler.MinReplicaCount = 1
	config.Spec.Autoscaler.MaxReplicaCount = 10
	config.Spec.Autoscaler.QueueType = "workflow"
	config.Spec.Autoscaler.TargetQueueSize = 2
	config.Spec.Autoscaler.PollingIntervalSeconds = 15
	config.Spec.Autoscaler.CooldownPeriodSeconds = 30

	// Resource defaults
	config.Spec.Resources.Requests.CPU = "100m"
	config.Spec.Resources.Requests.Memory = "128Mi"
	config.Spec.Resources.Limits.CPU = "500m"
	config.Spec.Resources.Limits.Memory = "256Mi"

	// Default worker settings
	config.Spec.WorkerSettings = `logLevel: "info"

# Platform configuration
features:
  enableLogging: true
  enableMetrics: true
  enableRetries: true

# Processing configuration
maxConcurrentTasks: 5
batchSize: 100
retryCount: 3`

	return config
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *TemporalWorkerConfig, configPath string) error {
	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *TemporalWorkerConfig) Validate() error {
	// Perform validation on critical fields
	if c.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	if c.Spec.WorkerType == "" {
		return fmt.Errorf("spec.workerType is required")
	}

	return nil
}
