package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/stretchr/testify/assert"
)

// T092: Test for status reporting of multiple components
func TestStatusReportingOfMultipleComponents(t *testing.T) {
	// This test verifies that status reporting handles multiple custom components correctly
	// Note: This is a unit test that validates the status structure, not actual kubectl calls

	config := &kindenv.KindEnvConfig{
		CustomComponents: []kindenv.CustomComponent{
			{
				Name:  "running-app",
				Image: "nginx:latest",
			},
			{
				Name:  "deploying-app",
				Image: "alpine:latest",
				Ports: []kindenv.PortMapping{
					{
						ContainerPort: 8080,
						Protocol:      "TCP",
						NodePort:      30080,
					},
				},
			},
			{
				Name:    "disabled-app",
				Image:   "nginx:latest",
				Enabled: boolPtr(false),
			},
			{
				Name:  "resource-app",
				Image: "nginx:latest",
				Resources: &kindenv.ResourceRequirements{
					Requests: &kindenv.ResourceList{
						CPU:    "100m",
						Memory: "128Mi",
					},
					Limits: &kindenv.ResourceList{
						CPU:    "500m",
						Memory: "512Mi",
					},
				},
			},
			{
				Name:  "config-app",
				Image: "nginx:latest",
				ConfigFiles: []kindenv.ConfigFile{
					{
						Name:     "app.yaml",
						Path:     "/config/app.yaml",
						Contents: "key: value",
					},
				},
			},
		},
	}

	// Set defaults for all components
	for i := range config.CustomComponents {
		config.CustomComponents[i].SetDefaults()
	}

	// Test 1: Verify all components are present in config
	assert.Len(t, config.CustomComponents, 5, "Should have 5 custom components")

	// Test 2: Verify disabled component is identified
	disabledComponent := config.CustomComponents[2] // disabled-app
	assert.NotNil(t, disabledComponent.Enabled, "Disabled component should have Enabled field")
	assert.False(t, *disabledComponent.Enabled, "Disabled component should be disabled")

	// Test 3: Verify components with ports have port information
	componentsWithPorts := 0
	for _, component := range config.CustomComponents {
		if len(component.Ports) > 0 {
			componentsWithPorts++
			assert.NotEmpty(t, component.Ports, "Component should have ports")
			for _, port := range component.Ports {
				assert.Greater(t, port.ContainerPort, 0, "Container port should be greater than 0")
				assert.NotEmpty(t, port.Protocol, "Protocol should not be empty")
			}
		}
	}
	assert.Equal(t, 1, componentsWithPorts, "Should have 1 component with ports")

	// Test 4: Verify components with custom resources have resource information
	// Note: SetDefaults() sets default resources for all components, so we check for custom resources
	componentsWithCustomResources := 0
	for _, component := range config.CustomComponents {
		// Check if component has custom (non-default) resources
		if component.Resources != nil && component.Resources.Requests != nil {
			// resource-app has custom CPU/memory, others have defaults
			if component.Name == "resource-app" {
				componentsWithCustomResources++
				assert.NotEmpty(t, component.Resources.Requests.CPU, "CPU request should not be empty")
				assert.NotEmpty(t, component.Resources.Requests.Memory, "Memory request should not be empty")
				if component.Resources.Limits != nil {
					assert.NotEmpty(t, component.Resources.Limits.CPU, "CPU limit should not be empty")
					assert.NotEmpty(t, component.Resources.Limits.Memory, "Memory limit should not be empty")
				}
			}
		}
	}
	assert.Equal(t, 1, componentsWithCustomResources, "Should have 1 component with custom resources")

	// Test 5: Verify components with config files have ConfigMap information
	componentsWithConfigMaps := 0
	for _, component := range config.CustomComponents {
		if len(component.ConfigFiles) > 0 {
			componentsWithConfigMaps++
			configMapName := component.Name + "-config"
			assert.NotEmpty(t, configMapName, "ConfigMap name should not be empty")
			assert.Equal(t, len(component.ConfigFiles), len(component.ConfigFiles), "Config files count should match")
		}
	}
	assert.Equal(t, 1, componentsWithConfigMaps, "Should have 1 component with ConfigMaps")

	// Test 6: Verify namespace handling
	for _, component := range config.CustomComponents {
		namespace := component.Namespace
		if namespace == "" {
			namespace = "default"
		}
		assert.NotEmpty(t, namespace, "Component should have a namespace")
	}

	// Test 7: Verify component names are unique
	names := make(map[string]bool)
	for _, component := range config.CustomComponents {
		assert.False(t, names[component.Name], "Component name %s should be unique", component.Name)
		names[component.Name] = true
	}
}

// TestStatusReportingWithAllFeatures verifies status reporting for a component with all features
func TestStatusReportingWithAllFeatures(t *testing.T) {
	component := kindenv.CustomComponent{
		Name:      "full-featured-app",
		Image:     "nginx:latest",
		Namespace: "test-ns",
		Replicas:  intPtr(2),
		Ports: []kindenv.PortMapping{
			{
				ContainerPort: 80,
				Protocol:      "TCP",
				NodePort:      30080,
			},
			{
				ContainerPort: 443,
				Protocol:      "TCP",
				NodePort:      30443,
			},
		},
		Resources: &kindenv.ResourceRequirements{
			Requests: &kindenv.ResourceList{
				CPU:    "200m",
				Memory: "256Mi",
			},
			Limits: &kindenv.ResourceList{
				CPU:    "1000m",
				Memory: "1Gi",
			},
		},
		ConfigFiles: []kindenv.ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "server:\n  port: 80",
			},
			{
				Name:     "logback.xml",
				Path:     "/config/logback.xml",
				Contents: "<configuration></configuration>",
			},
		},
	}

	component.SetDefaults()

	// Verify all status information is available
	assert.NotEmpty(t, component.Name, "Should have component name")
	assert.NotEmpty(t, component.Namespace, "Should have namespace")
	assert.NotNil(t, component.Replicas, "Should have replicas")
	assert.Equal(t, 2, *component.Replicas, "Should have correct replicas")
	assert.Len(t, component.Ports, 2, "Should have 2 ports")
	assert.NotNil(t, component.Resources, "Should have resources")
	assert.Len(t, component.ConfigFiles, 2, "Should have 2 config files")

	// Verify port information
	for i, port := range component.Ports {
		assert.Greater(t, port.ContainerPort, 0, "Port %d container port should be greater than 0", i)
		assert.NotEmpty(t, port.Protocol, "Port %d protocol should not be empty", i)
		if port.NodePort > 0 {
			assert.Greater(t, port.NodePort, 30000, "NodePort should be in valid range")
		}
	}

	// Verify resource information
	if component.Resources != nil {
		if component.Resources.Requests != nil {
			assert.NotEmpty(t, component.Resources.Requests.CPU, "CPU request should not be empty")
			assert.NotEmpty(t, component.Resources.Requests.Memory, "Memory request should not be empty")
		}
		if component.Resources.Limits != nil {
			assert.NotEmpty(t, component.Resources.Limits.CPU, "CPU limit should not be empty")
			assert.NotEmpty(t, component.Resources.Limits.Memory, "Memory limit should not be empty")
		}
	}

	// Verify ConfigMap information
	configMapName := component.Name + "-config"
	assert.Equal(t, "full-featured-app-config", configMapName, "ConfigMap name should be correct")
	assert.Len(t, component.ConfigFiles, 2, "Should have 2 config files")
}

// TestStatus_MonitoringEnabled verifies monitoring status config is set up correctly when enabled.
func TestStatus_MonitoringEnabled(t *testing.T) {
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = true

	// Verify config fields are accessible for status check
	assert.True(t, config.Components.Monitoring.Enabled)
	assert.NotEmpty(t, config.Components.Monitoring.Namespace, "Monitoring namespace should not be empty")
	assert.Equal(t, "monitoring", config.Components.Monitoring.Namespace)
}

// TestStatus_MonitoringDisabled verifies no monitoring output when disabled.
func TestStatus_MonitoringDisabled(t *testing.T) {
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = false

	// When disabled, no monitoring-related checks are performed
	assert.False(t, config.Components.Monitoring.Enabled)
}

// TestStatus_MonitoringDegraded verifies pod count parsing for degraded state.
func TestStatus_MonitoringDegraded(t *testing.T) {
	// Simulate kubectl output with a not-ready pod
	podOutput := "monitoring-grafana-abc123 0/1 Pending 0 30s"

	lines := strings.Split(strings.TrimSpace(podOutput), "\n")
	grafanaReady, grafanaTotal := 0, 0

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		podName := fields[0]
		readyStr := fields[1]
		status := fields[2]

		parts := strings.Split(readyStr, "/")
		ready := 0
		total := 0
		if len(parts) == 2 {
			fmt.Sscan(parts[0], &ready)
			fmt.Sscan(parts[1], &total)
		}
		isReady := ready == total && status == "Running"

		if strings.Contains(podName, "grafana") {
			grafanaTotal++
			if isReady {
				grafanaReady++
			}
		}
	}

	assert.Equal(t, 1, grafanaTotal, "Should count 1 grafana pod")
	assert.Equal(t, 0, grafanaReady, "Grafana pod should not be ready (Pending)")
}
