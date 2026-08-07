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
	ServiceYAML    string // Service YAML if ports are configured
	ConfigMapYAML  string // ConfigMap YAML if config files are configured
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

		// Validate component configuration
		if err := component.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed for component '%s': %w", component.Name, err)
		}

		// Validate secret references exist (pre-deployment check)
		if err := validateSecretReferences(ctx, component); err != nil {
			return nil, fmt.Errorf("pre-deployment validation failed for component '%s': %w", component.Name, err)
		}

		// Assign ports and validate conflicts
		usedPorts := make(map[int]bool)
		if len(component.Ports) > 0 {
			// Collect used ports from other components
			for j := 0; j < i; j++ {
				for _, port := range enabledComponents[j].Ports {
					if port.NodePort != 0 {
						usedPorts[port.NodePort] = true
					}
				}
			}

			// Validate port conflicts
			if err := validatePortConflicts(component, usedPorts); err != nil {
				return nil, fmt.Errorf("port conflict detected for component '%s': %w", component.Name, err)
			}

			// Assign ports (auto-assign NodePorts if needed)
			if err := assignPorts(component, usedPorts); err != nil {
				return nil, fmt.Errorf("failed to assign ports for component '%s': %w", component.Name, err)
			}
		}

		// Generate deployment YAML
		deploymentYAML, err := generateDeploymentYAML(component, config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate deployment YAML for component '%s': %w", component.Name, err)
		}

		// Generate service YAML if ports are configured
		var serviceYAML string
		if len(component.Ports) > 0 {
			serviceYAML, err = generateServiceYAML(component)
			if err != nil {
				return nil, fmt.Errorf("failed to generate service YAML for component '%s': %w", component.Name, err)
			}
		}

		// Generate ConfigMap YAML if config files are configured
		var configMapYAML string
		if len(component.ConfigFiles) > 0 {
			configMapYAML, err = generateConfigMapYAML(component)
			if err != nil {
				return nil, fmt.Errorf("failed to generate ConfigMap YAML for component '%s': %w", component.Name, err)
			}
		}

		deploymentInfos = append(deploymentInfos, DeploymentInfo{
			Component:      *component,
			DeploymentYAML: deploymentYAML,
			ServiceYAML:    serviceYAML,
			ConfigMapYAML:  configMapYAML,
			Namespace:      component.Namespace,
			Name:           component.Name,
		})
	}

	return deploymentInfos, nil
}

// DeploymentYAML represents the Kubernetes Deployment structure for YAML generation
type DeploymentYAML struct {
	APIVersion string             `yaml:"apiVersion"`
	Kind       string             `yaml:"kind"`
	Metadata   DeploymentMetadata `yaml:"metadata"`
	Spec       DeploymentSpec     `yaml:"spec"`
}

// DeploymentMetadata represents deployment metadata
type DeploymentMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// DeploymentSpec represents deployment specification
type DeploymentSpec struct {
	Replicas *int        `yaml:"replicas"`
	Selector Selector    `yaml:"selector"`
	Template PodTemplate `yaml:"template"`
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
	Containers       []Container       `yaml:"containers"`
	Volumes          []Volume          `yaml:"volumes,omitempty"`
	ImagePullSecrets []ImagePullSecret `yaml:"imagePullSecrets,omitempty"`
}

// ImagePullSecret represents an image pull secret reference
type ImagePullSecret struct {
	Name string `yaml:"name"`
}

// Container represents a container specification
type Container struct {
	Name         string              `yaml:"name"`
	Image        string              `yaml:"image"`
	Command      []string            `yaml:"command,omitempty"`
	Args         []string            `yaml:"args,omitempty"`
	Env          []EnvVarYAML        `yaml:"env,omitempty"`
	Ports        []ContainerPort     `yaml:"ports,omitempty"`
	Resources    *ContainerResources `yaml:"resources,omitempty"`
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
	Name      string           `yaml:"name"`
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
func generateDeploymentYAML(component *CustomComponent, config *KindEnvConfig) (string, error) {
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

	// Build volumes and volume mounts for config files
	var volumes []Volume
	var volumeMounts []VolumeMount
	if len(component.ConfigFiles) > 0 {
		// Generate volume specs
		volumeSpecs, err := generateVolumes(component)
		if err != nil {
			return "", fmt.Errorf("failed to generate volumes: %w", err)
		}

		// Convert VolumeSpec to Volume for YAML generation
		for _, vs := range volumeSpecs {
			if vs.ConfigMap != nil {
				volumes = append(volumes, Volume{
					Name: vs.Name,
					ConfigMap: &ConfigMapVolume{
						Name:        vs.ConfigMap.Name,
						DefaultMode: vs.ConfigMap.DefaultMode,
					},
				})
			}
		}

		// Generate volumeMount specs
		volumeMountSpecs, err := generateVolumeMounts(component)
		if err != nil {
			return "", fmt.Errorf("failed to generate volumeMounts: %w", err)
		}

		// Convert VolumeMountSpec to VolumeMount for YAML generation
		for _, vms := range volumeMountSpecs {
			volumeMounts = append(volumeMounts, VolumeMount{
				Name:      vms.Name,
				MountPath: vms.MountPath,
				SubPath:   vms.SubPath,
				ReadOnly:  vms.ReadOnly,
			})
		}

		// Add volumeMounts to container
		container.VolumeMounts = volumeMounts
	}

	// Build pod spec
	podSpec := PodSpec{
		Containers: []Container{container},
		Volumes:    volumes,
	}

	// Add imagePullSecrets if ECR is enabled
	if config != nil && config.Images.UseAwsEcr {
		podSpec.ImagePullSecrets = []ImagePullSecret{
			{Name: "ecr-credentials"},
		}
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

// assignPorts assigns NodePort and HostPort values to ports that don't have them specified
func assignPorts(component *CustomComponent, usedPorts map[int]bool) error {
	for i := range component.Ports {
		port := &component.Ports[i]

		// Auto-assign NodePort if not specified
		if port.NodePort == 0 {
			nodePort, err := findAvailableNodePort(usedPorts)
			if err != nil {
				return fmt.Errorf("failed to assign NodePort for component '%s': %w", component.Name, err)
			}
			port.NodePort = nodePort
			usedPorts[nodePort] = true
		} else {
			// Validate specified NodePort is not in use
			if usedPorts[port.NodePort] {
				return fmt.Errorf("NodePort %d is already in use (component: %s)", port.NodePort, component.Name)
			}
			usedPorts[port.NodePort] = true
		}

		// Default HostPort to ContainerPort if not specified
		if port.HostPort == 0 {
			port.HostPort = port.ContainerPort
		}
	}

	return nil
}

// findAvailableNodePort finds an available NodePort in the range 30000-32767
func findAvailableNodePort(usedPorts map[int]bool) (int, error) {
	// Start from 30000 and find first available port
	for port := 30000; port <= 32767; port++ {
		if !usedPorts[port] {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available NodePort in range 30000-32767")
}

// ServiceYAML represents the Kubernetes Service structure for YAML generation
type ServiceYAML struct {
	APIVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   ServiceMetadata `yaml:"metadata"`
	Spec       ServiceSpec     `yaml:"spec"`
}

// ServiceMetadata represents service metadata
type ServiceMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// ServiceSpec represents service specification
type ServiceSpec struct {
	Type     string            `yaml:"type"`
	Selector map[string]string `yaml:"selector"`
	Ports    []ServicePortYAML `yaml:"ports"`
}

// ServicePortYAML represents a service port
type ServicePortYAML struct {
	Name       string `yaml:"name,omitempty"`
	Protocol   string `yaml:"protocol"`
	Port       int    `yaml:"port"`
	TargetPort int    `yaml:"targetPort"`
	NodePort   int    `yaml:"nodePort,omitempty"`
}

// generateServiceYAML generates Kubernetes Service YAML for a custom component
func generateServiceYAML(component *CustomComponent) (string, error) {
	if len(component.Ports) == 0 {
		return "", nil // No service needed if no ports
	}

	// Build labels (same as deployment)
	labels := make(map[string]string)
	labels["app"] = component.Name
	labels["managed-by"] = "kindenv"
	labels["component-type"] = "custom"
	for k, v := range component.Labels {
		labels[k] = v
	}

	// Build selector (same as deployment)
	selector := map[string]string{
		"app":        component.Name,
		"managed-by": "kindenv",
	}

	// Build service ports
	var servicePorts []ServicePortYAML
	for _, port := range component.Ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = "TCP"
		}

		servicePort := ServicePortYAML{
			Protocol:   strings.ToUpper(protocol),
			Port:       port.ContainerPort,
			TargetPort: port.ContainerPort,
		}

		// Add name if multiple ports
		if len(component.Ports) > 1 {
			servicePort.Name = fmt.Sprintf("port-%d", port.ContainerPort)
		}

		// Add NodePort if specified
		if port.NodePort != 0 {
			servicePort.NodePort = port.NodePort
		}

		servicePorts = append(servicePorts, servicePort)
	}

	// Build service
	service := ServiceYAML{
		APIVersion: "v1",
		Kind:       "Service",
		Metadata: ServiceMetadata{
			Name:      component.Name,
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: ServiceSpec{
			Type:     "NodePort",
			Selector: selector,
			Ports:    servicePorts,
		},
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(service)
	if err != nil {
		return "", fmt.Errorf("failed to marshal service to YAML: %w", err)
	}

	return string(yamlBytes), nil
}
