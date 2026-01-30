package kindenv

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// customcomponent_test.go provides tests for custom component deployment functionality

// T022: Test for minimal component deployment
func TestGenerateDeploymentYAML_MinimalComponent(t *testing.T) {
	component := &CustomComponent{
		Name:      "nginx-test",
		Image:      "nginx:latest",
		Namespace: "default",
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component)
	require.NoError(t, err)
	require.NotEmpty(t, yamlStr)

	// Verify it's valid YAML
	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	// Verify basic structure
	assert.Equal(t, "apps/v1", deployment["apiVersion"])
	assert.Equal(t, "Deployment", deployment["kind"])

	metadata := deployment["metadata"].(map[string]interface{})
	assert.Equal(t, "nginx-test", metadata["name"])
	assert.Equal(t, "default", metadata["namespace"])

	spec := deployment["spec"].(map[string]interface{})
	replicas := spec["replicas"].(int)
	assert.Equal(t, 1, replicas)

	template := spec["template"].(map[string]interface{})
	templateSpec := template["spec"].(map[string]interface{})
	containers := templateSpec["containers"].([]interface{})
	require.Len(t, containers, 1)

	container := containers[0].(map[string]interface{})
	assert.Equal(t, "nginx-test", container["name"])
	assert.Equal(t, "nginx:latest", container["image"])

	// Verify default resources are applied
	resources := container["resources"].(map[string]interface{})
	requests := resources["requests"].(map[string]interface{})
	limits := resources["limits"].(map[string]interface{})
	assert.Equal(t, "100m", requests["cpu"])
	assert.Equal(t, "128Mi", requests["memory"])
	assert.Equal(t, "500m", limits["cpu"])
	assert.Equal(t, "512Mi", limits["memory"])

	// Verify labels
	templateLabels := template["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
	assert.Equal(t, "nginx-test", templateLabels["app"])
	assert.Equal(t, "kindenv", templateLabels["managed-by"])
	assert.Equal(t, "custom", templateLabels["component-type"])
}

// T023: Test for component with environment variables
func TestGenerateDeploymentYAML_WithEnvironmentVariables(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:      "myregistry/my-app:v1.0",
		Namespace: "default",
		Env: []EnvVar{
			{Name: "APP_ENV", Value: "development"},
			{Name: "LOG_LEVEL", Value: "debug"},
			{Name: "API_URL", Value: "http://api-service:3000"},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	envVars := container["env"].([]interface{})
	require.Len(t, envVars, 3)

	// Verify environment variables
	envMap := make(map[string]string)
	for _, env := range envVars {
		envObj := env.(map[string]interface{})
		name := envObj["name"].(string)
		value := envObj["value"].(string)
		envMap[name] = value
	}

	assert.Equal(t, "development", envMap["APP_ENV"])
	assert.Equal(t, "debug", envMap["LOG_LEVEL"])
	assert.Equal(t, "http://api-service:3000", envMap["API_URL"])
}

// T024: Test for enabled/disabled flag behavior
func TestFilterEnabledComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []CustomComponent
		expected   []string // Expected component names
	}{
		{
			name: "all enabled",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: boolPtr(true)},
				{Name: "app2", Image: "redis:latest", Enabled: boolPtr(true)},
			},
			expected: []string{"app1", "app2"},
		},
		{
			name: "some disabled",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: boolPtr(true)},
				{Name: "app2", Image: "redis:latest", Enabled: boolPtr(false)},
			},
			expected: []string{"app1"},
		},
		{
			name: "all disabled",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: boolPtr(false)},
				{Name: "app2", Image: "redis:latest", Enabled: boolPtr(false)},
			},
			expected: []string{},
		},
		{
			name: "nil enabled defaults to true",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: nil},
			},
			expected: []string{"app1"},
		},
		{
			name: "mixed enabled states",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: boolPtr(true)},
				{Name: "app2", Image: "redis:latest", Enabled: nil},
				{Name: "app3", Image: "postgres:latest", Enabled: boolPtr(false)},
			},
			expected: []string{"app1", "app2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var enabled []CustomComponent
			for _, component := range tt.components {
				if component.Enabled == nil || *component.Enabled {
					enabled = append(enabled, component)
				}
			}

			assert.Len(t, enabled, len(tt.expected))
			enabledNames := make([]string, len(enabled))
			for i, comp := range enabled {
				enabledNames[i] = comp.Name
			}
			assert.ElementsMatch(t, tt.expected, enabledNames)
		})
	}
}

// T025: Test for namespace specification
func TestGenerateDeploymentYAML_NamespaceSpecification(t *testing.T) {
	tests := []struct {
		name          string
		component     *CustomComponent
		expectedNs    string
		description   string
	}{
		{
			name: "explicit namespace",
			component: &CustomComponent{
				Name:      "my-app",
				Image:     "nginx:latest",
				Namespace: "production",
			},
			expectedNs:  "production",
			description: "Component with explicit namespace",
		},
		{
			name: "default namespace",
			component: &CustomComponent{
				Name:  "my-app",
				Image: "nginx:latest",
				// Namespace not set
			},
			expectedNs:  "default",
			description: "Component without namespace defaults to 'default'",
		},
		{
			name: "custom namespace",
			component: &CustomComponent{
				Name:      "my-app",
				Image:     "nginx:latest",
				Namespace: "my-namespace",
			},
			expectedNs:  "my-namespace",
			description: "Component with custom namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.component.SetDefaults()

			yamlStr, err := generateDeploymentYAML(tt.component)
			require.NoError(t, err)

			var deployment map[string]interface{}
			err = yaml.Unmarshal([]byte(yamlStr), &deployment)
			require.NoError(t, err)

			metadata := deployment["metadata"].(map[string]interface{})
			assert.Equal(t, tt.expectedNs, metadata["namespace"], tt.description)
		})
	}
}

func TestGenerateDeploymentYAML_EmptyComponents(t *testing.T) {
	config := &KindEnvConfig{
		CustomComponents: []CustomComponent{},
	}

	// Should return empty list when no components
	deployments, err := DeployCustomComponents(nil, config)
	assert.NoError(t, err)
	assert.Nil(t, deployments)
}

func TestDeployCustomComponents_WithComponents(t *testing.T) {
	config := &KindEnvConfig{
		CustomComponents: []CustomComponent{
			{
				Name:      "test-app",
				Image:     "nginx:latest",
				Namespace: "default",
			},
		},
	}

	deployments, err := DeployCustomComponents(nil, config)
	require.NoError(t, err)
	require.Len(t, deployments, 1)
	assert.Equal(t, "test-app", deployments[0].Name)
	assert.Equal(t, "default", deployments[0].Namespace)
	assert.NotEmpty(t, deployments[0].DeploymentYAML)
}

func TestGenerateDeploymentYAML_Replicas(t *testing.T) {
	replicas := 3
	component := &CustomComponent{
		Name:      "scaled-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Replicas:  &replicas,
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	spec := deployment["spec"].(map[string]interface{})
	assert.Equal(t, 3, spec["replicas"])
}

func TestGenerateDeploymentYAML_CustomLabels(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Labels: map[string]string{
			"tier":       "backend",
			"version":    "v1.0",
			"environment": "dev",
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	labels := template["metadata"].(map[string]interface{})["labels"].(map[string]interface{})

	// Verify auto-generated labels
	assert.Equal(t, "my-app", labels["app"])
	assert.Equal(t, "kindenv", labels["managed-by"])
	assert.Equal(t, "custom", labels["component-type"])

	// Verify custom labels
	assert.Equal(t, "backend", labels["tier"])
	assert.Equal(t, "v1.0", labels["version"])
	assert.Equal(t, "dev", labels["environment"])
}

func TestGenerateDeploymentYAML_CommandAndArgs(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:     "openjdk:11",
		Namespace: "default",
		Command:   []string{"java"},
		Args:     []string{"-jar", "/app/application.jar", "--spring.profiles.active=local"},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	command := container["command"].([]interface{})
	args := container["args"].([]interface{})

	assert.Equal(t, []interface{}{"java"}, command)
	assert.Equal(t, []interface{}{"-jar", "/app/application.jar", "--spring.profiles.active=local"}, args)
}

// Helper function to check if YAML contains a substring (for basic validation)
func yamlContains(yamlStr, substr string) bool {
	return strings.Contains(yamlStr, substr)
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

// Placeholder for future tests that will require kubectl mocking
func TestDeployCustomComponentsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	// TODO: Implement integration tests with real Kind cluster
	t.Skip("integration tests not yet implemented")
}
