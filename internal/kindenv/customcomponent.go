package kindenv

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// customcomponent.go provides the core functionality for deploying and managing
// custom components in the Kind cluster. This includes deployment generation,
// service creation, and orchestration logic.

// DeploymentInfo contains information needed to deploy a custom component
type DeploymentInfo struct {
	Component      CustomComponent
	DeploymentYAML string
	Namespace      string
	Name           string
}

// DeployCustomComponents prepares deployment information for all enabled custom components
// Returns a list of DeploymentInfo that can be applied to the cluster
func DeployCustomComponents(ctx context.Context, config *KindEnvConfig) ([]DeploymentInfo, error) {
	if len(config.CustomComponents) == 0 {
		return nil, nil
	}

	// Filter enabled components
	var enabledComponents []CustomComponent
	for _, component := range config.CustomComponents {
		if component.Enabled == nil || *component.Enabled {
			enabledComponents = append(enabledComponents, component)
		}
	}

	if len(enabledComponents) == 0 {
		return nil, nil
	}

	// Validate and generate deployment YAML for each component
	var deploymentInfos []DeploymentInfo
	for i := range enabledComponents {
		component := &enabledComponents[i]

		// Set defaults
		component.SetDefaults()

		// Validate component
		if err := component.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed for component '%s': %w", component.Name, err)
		}

		// Generate deployment YAML
		deploymentYAML, err := generateDeploymentYAML(component)
		if err != nil {
			return nil, fmt.Errorf("failed to generate deployment YAML for component '%s': %w", component.Name, err)
		}

		deploymentInfos = append(deploymentInfos, DeploymentInfo{
			Component:      *component,
			DeploymentYAML: deploymentYAML,
			Namespace:      component.Namespace,
			Name:           component.Name,
		})
	}

	return deploymentInfos, nil
}

// DeploymentYAML represents the Kubernetes Deployment structure for YAML generation
type DeploymentYAML struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   DeploymentMetadata     `yaml:"metadata"`
	Spec       DeploymentSpec         `yaml:"spec"`
}

// DeploymentMetadata represents deployment metadata
type DeploymentMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// DeploymentSpec represents deployment specification
type DeploymentSpec struct {
	Replicas *int              `yaml:"replicas"`
	Selector Selector          `yaml:"selector"`
	Template PodTemplate       `yaml:"template"`
}

// Selector represents label selector
type Selector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

// PodTemplate represents pod template
type PodTemplate struct {
	Metadata PodMetadata `yaml:"metadata"`
	Spec     PodSpec     `yaml:"spec"`
}

// PodMetadata represents pod metadata
type PodMetadata struct {
	Labels map[string]string `yaml:"labels"`
}

// PodSpec represents pod specification
type PodSpec struct {
	Containers []Container `yaml:"containers"`
	Volumes    []Volume    `yaml:"volumes,omitempty"`
}

// Container represents a container specification
type Container struct {
	Name      string                 `yaml:"name"`
	Image     string                 `yaml:"image"`
	Command   []string               `yaml:"command,omitempty"`
	Args      []string               `yaml:"args,omitempty"`
	Env       []EnvVarYAML           `yaml:"env,omitempty"`
	Ports     []ContainerPort        `yaml:"ports,omitempty"`
	Resources  *ContainerResources    `yaml:"resources,omitempty"`
	VolumeMounts []VolumeMount       `yaml:"volumeMounts,omitempty"`
}

// EnvVarYAML represents an environment variable in YAML
type EnvVarYAML struct {
	Name      string        `yaml:"name"`
	Value     string        `yaml:"value,omitempty"`
	ValueFrom *EnvVarSource `yaml:"valueFrom,omitempty"`
}

// ContainerPort represents a container port
type ContainerPort struct {
	Name          string `yaml:"name,omitempty"`
	ContainerPort int    `yaml:"containerPort"`
	Protocol      string `yaml:"protocol,omitempty"`
}

// ContainerResources represents container resource requirements
type ContainerResources struct {
	Requests *ResourceList `yaml:"requests,omitempty"`
	Limits   *ResourceList `yaml:"limits,omitempty"`
}

// Volume represents a volume specification
type Volume struct {
	Name     string            `yaml:"name"`
	ConfigMap *ConfigMapVolume `yaml:"configMap,omitempty"`
}

// ConfigMapVolume represents a ConfigMap volume source
type ConfigMapVolume struct {
	Name        string `yaml:"name"`
	DefaultMode int    `yaml:"defaultMode"`
}

// VolumeMount represents a volume mount
type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
	SubPath   string `yaml:"subPath,omitempty"`
	ReadOnly  bool   `yaml:"readOnly,omitempty"`
}

// generateDeploymentYAML generates Kubernetes Deployment YAML for a custom component
func generateDeploymentYAML(component *CustomComponent) (string, error) {
	// Ensure defaults are set
	component.SetDefaults()

	// Validate component
	if err := component.Validate(); err != nil {
		return "", fmt.Errorf("invalid component configuration: %w", err)
	}

	// Set replicas (default to 1 if nil)
	replicas := 1
	if component.Replicas != nil {
		replicas = *component.Replicas
	}

	// Build labels (merge auto-generated with custom)
	labels := make(map[string]string)
	// Auto-generated labels
	labels["app"] = component.Name
	labels["managed-by"] = "kindenv"
	labels["component-type"] = "custom"
	// Add custom labels
	for k, v := range component.Labels {
		labels[k] = v
	}

	// Build selector labels (only app and managed-by for selector)
	selectorLabels := map[string]string{
		"app":        component.Name,
		"managed-by": "kindenv",
	}

	// Build environment variables
	var envVars []EnvVarYAML
	for _, env := range component.Env {
		envVar := EnvVarYAML{
			Name: env.Name,
		}
		if env.Value != "" {
			envVar.Value = env.Value
		} else if env.ValueFrom != nil {
			envVar.ValueFrom = env.ValueFrom
		}
		envVars = append(envVars, envVar)
	}

	// Build container ports
	var ports []ContainerPort
	for _, port := range component.Ports {
		portName := fmt.Sprintf("port-%d", port.ContainerPort)
		protocol := port.Protocol
		if protocol == "" {
			protocol = "TCP"
		}
		ports = append(ports, ContainerPort{
			Name:          portName,
			ContainerPort: port.ContainerPort,
			Protocol:      strings.ToUpper(protocol),
		})
	}

	// Build resources
	var resources *ContainerResources
	if component.Resources != nil {
		resources = &ContainerResources{
			Requests: component.Resources.Requests,
			Limits:   component.Resources.Limits,
		}
	} else {
		// Apply defaults
		defaultRes := defaultResourceRequirements()
		resources = &ContainerResources{
			Requests: defaultRes.Requests,
			Limits:   defaultRes.Limits,
		}
	}

	// Build container
	container := Container{
		Name:      component.Name,
		Image:     component.Image,
		Env:       envVars,
		Ports:     ports,
		Resources: resources,
	}

	// Add command if specified
	if len(component.Command) > 0 {
		container.Command = component.Command
	}

	// Add args if specified
	if len(component.Args) > 0 {
		container.Args = component.Args
	}

	// Build volumes and volume mounts for config files (will be implemented in US6)
	// For now, we'll leave this empty for User Story 1

	// Build pod spec
	podSpec := PodSpec{
		Containers: []Container{container},
	}

	// Build deployment
	deployment := DeploymentYAML{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: DeploymentMetadata{
			Name:      component.Name,
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: DeploymentSpec{
			Replicas: &replicas,
			Selector: Selector{
				MatchLabels: selectorLabels,
			},
			Template: PodTemplate{
				Metadata: PodMetadata{
					Labels: labels,
				},
				Spec: podSpec,
			},
		},
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(deployment)
	if err != nil {
		return "", fmt.Errorf("failed to marshal deployment to YAML: %w", err)
	}

	return string(yamlBytes), nil
}
