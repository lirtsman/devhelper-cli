package kindenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// configmap_test.go provides tests for ConfigMap generation functionality

// T071: Test for ConfigMap generation
func TestGenerateConfigMapYAML_SingleFile(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		ConfigFiles: []ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "server:\n  port: 8080",
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateConfigMapYAML(component)
	require.NoError(t, err)
	require.NotEmpty(t, yamlStr)

	// Verify it's valid YAML
	var configMap map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &configMap)
	require.NoError(t, err)

	// Verify basic structure
	assert.Equal(t, "v1", configMap["apiVersion"])
	assert.Equal(t, "ConfigMap", configMap["kind"])

	metadata := configMap["metadata"].(map[string]interface{})
	assert.Equal(t, "test-app-config", metadata["name"])
	assert.Equal(t, "default", metadata["namespace"])

	// Verify labels
	labels := metadata["labels"].(map[string]interface{})
	assert.Equal(t, "test-app", labels["app"])
	assert.Equal(t, "kindenv", labels["managed-by"])
	assert.Equal(t, "custom", labels["component-type"])

	// Verify data
	data := configMap["data"].(map[string]interface{})
	assert.Equal(t, "server:\n  port: 8080", data["app.yaml"])
}

// T072: Test for multiple config files
func TestGenerateConfigMapYAML_MultipleFiles(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		ConfigFiles: []ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "server:\n  port: 8080",
			},
			{
				Name:     "logback.xml",
				Path:     "/config/logback.xml",
				Contents: "<configuration>\n  <appender name=\"STDOUT\"/>\n</configuration>",
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateConfigMapYAML(component)
	require.NoError(t, err)
	require.NotEmpty(t, yamlStr)

	// Verify it's valid YAML
	var configMap map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &configMap)
	require.NoError(t, err)

	// Verify data contains both files
	data := configMap["data"].(map[string]interface{})
	assert.Equal(t, "server:\n  port: 8080", data["app.yaml"])
	assert.Equal(t, "<configuration>\n  <appender name=\"STDOUT\"/>\n</configuration>", data["logback.xml"])
	assert.Len(t, data, 2)
}

// Test for empty config files (should return empty string)
func TestGenerateConfigMapYAML_NoConfigFiles(t *testing.T) {
	component := &CustomComponent{
		Name:        "test-app",
		Image:       "nginx:latest",
		Namespace:   "default",
		ConfigFiles: []ConfigFile{},
	}
	component.SetDefaults()

	yamlStr, err := generateConfigMapYAML(component)
	require.NoError(t, err)
	assert.Empty(t, yamlStr)
}

// Test for ConfigMap with custom labels
func TestGenerateConfigMapYAML_WithCustomLabels(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Labels: map[string]string{
			"tier":      "backend",
			"component": "api",
		},
		ConfigFiles: []ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "key: value",
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateConfigMapYAML(component)
	require.NoError(t, err)

	var configMap map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &configMap)
	require.NoError(t, err)

	metadata := configMap["metadata"].(map[string]interface{})
	labels := metadata["labels"].(map[string]interface{})

	// Verify auto-generated labels
	assert.Equal(t, "test-app", labels["app"])
	assert.Equal(t, "kindenv", labels["managed-by"])
	assert.Equal(t, "custom", labels["component-type"])

	// Verify custom labels
	assert.Equal(t, "backend", labels["tier"])
	assert.Equal(t, "api", labels["component"])
}
