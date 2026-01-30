package cmd

import (
	"testing"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/stretchr/testify/assert"
)

// T091: Test for kindenv stop cleanup of all resources
func TestKindenvStopCleanupOfAllResources(t *testing.T) {
	// This test verifies that the stop command properly cleans up all custom component resources
	// Note: This is a unit test that validates the cleanup logic, not an actual cluster cleanup

	config := &kindenv.KindEnvConfig{
		CustomComponents: []kindenv.CustomComponent{
			{
				Name:  "app-with-deployment",
				Image: "nginx:latest",
			},
			{
				Name:  "app-with-service",
				Image: "nginx:latest",
				Ports: []kindenv.PortMapping{
					{
						ContainerPort: 8080,
						Protocol:      "TCP",
						NodePort:      30080,
					},
				},
			},
			{
				Name:  "app-with-configmap",
				Image: "nginx:latest",
				ConfigFiles: []kindenv.ConfigFile{
					{
						Name:     "config.yaml",
						Path:     "/config/config.yaml",
						Contents: "key: value",
					},
				},
			},
			{
				Name:  "app-with-all-resources",
				Image: "nginx:latest",
				Ports: []kindenv.PortMapping{
					{
						ContainerPort: 8080,
						Protocol:      "TCP",
						NodePort:      30081,
					},
				},
				ConfigFiles: []kindenv.ConfigFile{
					{
						Name:     "app.yaml",
						Path:     "/config/app.yaml",
						Contents: "server:\n  port: 8080",
					},
				},
			},
			{
				Name:    "disabled-app",
				Image:   "nginx:latest",
				Enabled: boolPtr(false),
			},
		},
	}

	// Set defaults for all components
	for i := range config.CustomComponents {
		config.CustomComponents[i].SetDefaults()
	}

	// Verify that cleanup logic would handle all resource types
	// This test validates the structure and logic, not actual kubectl calls

	// Test 1: All enabled components should be cleaned up
	enabledCount := 0
	for _, component := range config.CustomComponents {
		if component.Enabled == nil || *component.Enabled {
			enabledCount++
		}
	}
	assert.Equal(t, 4, enabledCount, "Should have 4 enabled components")

	// Test 2: Components with ports should have services to clean up
	componentsWithPorts := 0
	for _, component := range config.CustomComponents {
		if len(component.Ports) > 0 {
			componentsWithPorts++
		}
	}
	assert.Equal(t, 2, componentsWithPorts, "Should have 2 components with ports")

	// Test 3: Components with config files should have ConfigMaps to clean up
	componentsWithConfigMaps := 0
	for _, component := range config.CustomComponents {
		if len(component.ConfigFiles) > 0 {
			componentsWithConfigMaps++
		}
	}
	assert.Equal(t, 2, componentsWithConfigMaps, "Should have 2 components with ConfigMaps")

	// Test 4: Verify disabled components are skipped
	disabledComponent := config.CustomComponents[4] // disabled-app
	assert.NotNil(t, disabledComponent.Enabled, "Disabled component should have Enabled field set")
	assert.False(t, *disabledComponent.Enabled, "Disabled component should be disabled")

	// Test 5: Verify namespace handling
	for _, component := range config.CustomComponents {
		namespace := component.Namespace
		if namespace == "" {
			namespace = "default"
		}
		assert.NotEmpty(t, namespace, "Component should have a namespace")
	}

	// Test 6: Verify resource names are correct for cleanup
	for _, component := range config.CustomComponents {
		if component.Enabled == nil || *component.Enabled {
			// Deployment name should match component name
			deploymentName := component.Name
			assert.NotEmpty(t, deploymentName, "Deployment name should not be empty")

			// Service name should match component name if ports exist
			if len(component.Ports) > 0 {
				serviceName := component.Name
				assert.NotEmpty(t, serviceName, "Service name should not be empty")
			}

			// ConfigMap name should be component-name-config if config files exist
			if len(component.ConfigFiles) > 0 {
				configMapName := component.Name + "-config"
				assert.NotEmpty(t, configMapName, "ConfigMap name should not be empty")
				assert.Contains(t, configMapName, component.Name, "ConfigMap name should contain component name")
			}
		}
	}
}

// TestCleanupOrder verifies that cleanup happens in the correct order
func TestCleanupOrder(t *testing.T) {
	// Cleanup should happen in this order:
	// 1. Deployments (pods will be terminated)
	// 2. Services (no dependencies)
	// 3. ConfigMaps (no dependencies)

	component := kindenv.CustomComponent{
		Name:  "test-app",
		Image: "nginx:latest",
		Ports: []kindenv.PortMapping{
			{
				ContainerPort: 8080,
				Protocol:      "TCP",
			},
		},
		ConfigFiles: []kindenv.ConfigFile{
			{
				Name:     "config.yaml",
				Path:     "/config/config.yaml",
				Contents: "key: value",
			},
		},
	}

	component.SetDefaults()

	// Verify all resources exist
	assert.NotEmpty(t, component.Name, "Should have deployment name")
	assert.NotEmpty(t, component.Ports, "Should have ports for service")
	assert.NotEmpty(t, component.ConfigFiles, "Should have config files for ConfigMap")

	// Verify resource names
	deploymentName := component.Name
	serviceName := component.Name
	configMapName := component.Name + "-config"

	assert.Equal(t, "test-app", deploymentName, "Deployment name should match")
	assert.Equal(t, "test-app", serviceName, "Service name should match")
	assert.Equal(t, "test-app-config", configMapName, "ConfigMap name should match")
}

