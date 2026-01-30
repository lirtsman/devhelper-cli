package kindenv

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// configmap.go provides functionality for creating and managing Kubernetes ConfigMaps
// for custom component configuration files.

// ConfigMapYAML represents the Kubernetes ConfigMap structure for YAML generation
type ConfigMapYAML struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   ConfigMapMetadata `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
}

// ConfigMapMetadata represents ConfigMap metadata
type ConfigMapMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// generateConfigMapYAML generates Kubernetes ConfigMap YAML for a custom component's config files
func generateConfigMapYAML(component *CustomComponent) (string, error) {
	if len(component.ConfigFiles) == 0 {
		return "", nil
	}

	// Build labels (same as deployment)
	labels := make(map[string]string)
	labels["app"] = component.Name
	labels["managed-by"] = "kindenv"
	labels["component-type"] = "custom"
	for k, v := range component.Labels {
		labels[k] = v
	}

	// Build data map from config files
	data := make(map[string]string)
	for _, configFile := range component.ConfigFiles {
		data[configFile.Name] = configFile.Contents
	}

	// Build ConfigMap
	configMapName := component.Name + "-config"
	configMap := ConfigMapYAML{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		Metadata: ConfigMapMetadata{
			Name:      configMapName,
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Data: data,
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(configMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ConfigMap to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// createConfigMap creates a ConfigMap in the cluster for a custom component
func createConfigMap(ctx context.Context, component *CustomComponent) error {
	if len(component.ConfigFiles) == 0 {
		return nil
	}

	// Generate ConfigMap YAML
	configMapYAML, err := generateConfigMapYAML(component)
	if err != nil {
		return fmt.Errorf("failed to generate ConfigMap YAML: %w", err)
	}

	// Apply ConfigMap using kubectl
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(configMapYAML)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create ConfigMap '%s': %w (output: %s)", component.Name+"-config", err, string(output))
	}

	return nil
}

// deleteConfigMap deletes a ConfigMap from the cluster
func deleteConfigMap(ctx context.Context, namespace, name string) error {
	// Delete ConfigMap using kubectl
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "configmap", name, "-n", namespace, "--ignore-not-found")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete ConfigMap '%s' in namespace '%s': %w (output: %s)", name, namespace, err, string(output))
	}

	return nil
}
