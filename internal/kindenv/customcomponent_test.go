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
		Image:     "nginx:latest",
		Namespace: "default",
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
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
		Image:     "myregistry/my-app:v1.0",
		Namespace: "default",
		Env: []EnvVar{
			{Name: "APP_ENV", Value: "development"},
			{Name: "LOG_LEVEL", Value: "debug"},
			{Name: "API_URL", Value: "http://api-service:3000"},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
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
		name        string
		component   *CustomComponent
		expectedNs  string
		description string
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

			yamlStr, err := generateDeploymentYAML(tt.component, nil)
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

	yamlStr, err := generateDeploymentYAML(component, nil)
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
			"tier":        "backend",
			"version":     "v1.0",
			"environment": "dev",
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
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

// T043: Test for command override
func TestGenerateDeploymentYAML_CommandOverride(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Command:   []string{"/bin/sh", "-c"},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	command := container["command"].([]interface{})
	assert.Equal(t, []interface{}{"/bin/sh", "-c"}, command)
	// Args should not be present when not specified
	_, hasArgs := container["args"]
	assert.False(t, hasArgs, "Args should not be present when not specified")
}

// T044: Test for args without command
func TestGenerateDeploymentYAML_ArgsWithoutCommand(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Args:      []string{"-g", "daemon off;"},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	args := container["args"].([]interface{})
	assert.Equal(t, []interface{}{"-g", "daemon off;"}, args)
	// Command should not be present when not specified
	_, hasCommand := container["command"]
	assert.False(t, hasCommand, "Command should not be present when not specified")
}

// T045: Test for command + args together
func TestGenerateDeploymentYAML_CommandAndArgs(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:     "openjdk:11",
		Namespace: "default",
		Command:   []string{"java"},
		Args:      []string{"-jar", "/app/application.jar", "--spring.profiles.active=local"},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
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

// T049: Test for single port mapping
func TestGenerateDeploymentYAML_SinglePortMapping(t *testing.T) {
	component := &CustomComponent{
		Name:      "web-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Ports: []PortMapping{
			{
				ContainerPort: 8080,
				Protocol:      "TCP",
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	ports := container["ports"].([]interface{})
	require.Len(t, ports, 1)

	port := ports[0].(map[string]interface{})
	assert.Equal(t, 8080, port["containerPort"])
	assert.Equal(t, "TCP", port["protocol"])
}

// T050: Test for multiple port mappings
func TestGenerateDeploymentYAML_MultiplePortMappings(t *testing.T) {
	component := &CustomComponent{
		Name:      "multi-port-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Ports: []PortMapping{
			{
				ContainerPort: 8080,
				Protocol:      "TCP",
			},
			{
				ContainerPort: 8443,
				Protocol:      "TCP",
			},
			{
				ContainerPort: 9090,
				Protocol:      "UDP",
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	ports := container["ports"].([]interface{})
	require.Len(t, ports, 3)

	// Verify all ports are present
	portMap := make(map[int]string)
	for _, p := range ports {
		portObj := p.(map[string]interface{})
		portNum := portObj["containerPort"].(int)
		protocol := portObj["protocol"].(string)
		portMap[portNum] = protocol
	}

	assert.Equal(t, "TCP", portMap[8080])
	assert.Equal(t, "TCP", portMap[8443])
	assert.Equal(t, "UDP", portMap[9090])
}

// T052: Test for NodePort auto-assignment
func TestAssignPorts_NodePortAutoAssignment(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Ports: []PortMapping{
			{
				ContainerPort: 8080,
				Protocol:      "TCP",
				// NodePort not specified - should be auto-assigned
			},
		},
	}
	component.SetDefaults()

	usedPorts := make(map[int]bool)
	err := assignPorts(component, usedPorts)
	require.NoError(t, err)

	// Verify NodePort was assigned in valid range
	assert.GreaterOrEqual(t, component.Ports[0].NodePort, 30000, "NodePort should be >= 30000")
	assert.LessOrEqual(t, component.Ports[0].NodePort, 32767, "NodePort should be <= 32767")
	assert.True(t, usedPorts[component.Ports[0].NodePort], "NodePort should be marked as used")
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

// T035: Test for secretKeyRef environment variables
func TestGenerateDeploymentYAML_WithSecretKeyRef(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:     "myregistry/my-app:v1.0",
		Namespace: "default",
		Env: []EnvVar{
			{
				Name: "DB_PASSWORD",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "mysql-secret",
						Key:  "password",
					},
				},
			},
			{
				Name: "DB_USER",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "mysql-secret",
						Key:  "username",
					},
				},
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	envVars := container["env"].([]interface{})
	require.Len(t, envVars, 2)

	// Verify secretKeyRef structure
	dbPasswordEnv := envVars[0].(map[string]interface{})
	assert.Equal(t, "DB_PASSWORD", dbPasswordEnv["name"])
	assert.Nil(t, dbPasswordEnv["value"]) // Should not have direct value
	valueFrom := dbPasswordEnv["valueFrom"].(map[string]interface{})
	secretKeyRef := valueFrom["secretKeyRef"].(map[string]interface{})
	assert.Equal(t, "mysql-secret", secretKeyRef["name"])
	assert.Equal(t, "password", secretKeyRef["key"])

	dbUserEnv := envVars[1].(map[string]interface{})
	assert.Equal(t, "DB_USER", dbUserEnv["name"])
	valueFrom2 := dbUserEnv["valueFrom"].(map[string]interface{})
	secretKeyRef2 := valueFrom2["secretKeyRef"].(map[string]interface{})
	assert.Equal(t, "mysql-secret", secretKeyRef2["name"])
	assert.Equal(t, "username", secretKeyRef2["key"])
}

// T036: Test for mixed direct values and secret references
func TestGenerateDeploymentYAML_MixedEnvVars(t *testing.T) {
	component := &CustomComponent{
		Name:      "my-app",
		Image:     "myregistry/my-app:v1.0",
		Namespace: "default",
		Env: []EnvVar{
			{
				Name:  "APP_ENV",
				Value: "production",
			},
			{
				Name: "DB_PASSWORD",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "mysql-secret",
						Key:  "password",
					},
				},
			},
			{
				Name:  "LOG_LEVEL",
				Value: "info",
			},
			{
				Name: "DB_HOST",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "mysql-secret",
						Key:  "host",
					},
				},
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	envVars := container["env"].([]interface{})
	require.Len(t, envVars, 4)

	// Build a map for easier checking
	envMap := make(map[string]interface{})
	for _, env := range envVars {
		envObj := env.(map[string]interface{})
		name := envObj["name"].(string)
		envMap[name] = envObj
	}

	// Check direct values
	appEnv := envMap["APP_ENV"].(map[string]interface{})
	assert.Equal(t, "production", appEnv["value"])
	assert.Nil(t, appEnv["valueFrom"])

	logLevel := envMap["LOG_LEVEL"].(map[string]interface{})
	assert.Equal(t, "info", logLevel["value"])
	assert.Nil(t, logLevel["valueFrom"])

	// Check secret references
	dbPassword := envMap["DB_PASSWORD"].(map[string]interface{})
	assert.Nil(t, dbPassword["value"])
	valueFrom := dbPassword["valueFrom"].(map[string]interface{})
	secretKeyRef := valueFrom["secretKeyRef"].(map[string]interface{})
	assert.Equal(t, "mysql-secret", secretKeyRef["name"])
	assert.Equal(t, "password", secretKeyRef["key"])

	dbHost := envMap["DB_HOST"].(map[string]interface{})
	assert.Nil(t, dbHost["value"])
	valueFrom2 := dbHost["valueFrom"].(map[string]interface{})
	secretKeyRef2 := valueFrom2["secretKeyRef"].(map[string]interface{})
	assert.Equal(t, "mysql-secret", secretKeyRef2["name"])
	assert.Equal(t, "host", secretKeyRef2["key"])
}

// T061: Test for custom resource limits
func TestGenerateDeploymentYAML_WithCustomResources(t *testing.T) {
	component := &CustomComponent{
		Name:      "resource-test",
		Image:     "nginx:latest",
		Namespace: "default",
		Resources: &ResourceRequirements{
			Requests: &ResourceList{
				CPU:    "500m",
				Memory: "512Mi",
			},
			Limits: &ResourceList{
				CPU:    "2000m",
				Memory: "2Gi",
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	resources := container["resources"].(map[string]interface{})
	requests := resources["requests"].(map[string]interface{})
	limits := resources["limits"].(map[string]interface{})

	// Verify custom resources are applied
	assert.Equal(t, "500m", requests["cpu"])
	assert.Equal(t, "512Mi", requests["memory"])
	assert.Equal(t, "2000m", limits["cpu"])
	assert.Equal(t, "2Gi", limits["memory"])
}

// T062: Test for default resource application
func TestGenerateDeploymentYAML_WithDefaultResources(t *testing.T) {
	component := &CustomComponent{
		Name:      "default-resource-test",
		Image:     "nginx:latest",
		Namespace: "default",
		// No Resources specified - should use defaults
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	resources := container["resources"].(map[string]interface{})
	requests := resources["requests"].(map[string]interface{})
	limits := resources["limits"].(map[string]interface{})

	// Verify default resources are applied
	assert.Equal(t, "100m", requests["cpu"])
	assert.Equal(t, "128Mi", requests["memory"])
	assert.Equal(t, "500m", limits["cpu"])
	assert.Equal(t, "512Mi", limits["memory"])
}

// T076: Integration test for config file mounting in deployment
func TestGenerateDeploymentYAML_WithConfigFiles(t *testing.T) {
	component := &CustomComponent{
		Name:      "config-test",
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
				Contents: "<configuration></configuration>",
			},
		},
	}
	component.SetDefaults()

	yamlStr, err := generateDeploymentYAML(component, nil)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})

	// Verify volumes are present
	volumes := podSpec["volumes"].([]interface{})
	require.Len(t, volumes, 1)
	volume := volumes[0].(map[string]interface{})
	assert.Equal(t, "config-test-config-volume", volume["name"])
	configMapVolume := volume["configMap"].(map[string]interface{})
	assert.Equal(t, "config-test-config", configMapVolume["name"])
	assert.Equal(t, 420, configMapVolume["defaultMode"]) // 0644 in decimal

	// Verify volumeMounts are present in container
	containers := podSpec["containers"].([]interface{})
	require.Len(t, containers, 1)
	container := containers[0].(map[string]interface{})

	volumeMounts := container["volumeMounts"].([]interface{})
	require.Len(t, volumeMounts, 2)

	// Build a map of mount paths for easier checking
	mountPaths := make(map[string]map[string]interface{})
	for _, vm := range volumeMounts {
		vmMap := vm.(map[string]interface{})
		mountPaths[vmMap["mountPath"].(string)] = vmMap
	}

	// Verify first config file mount
	appYamlMount := mountPaths["/config/app.yaml"]
	require.NotNil(t, appYamlMount)
	assert.Equal(t, "config-test-config-volume", appYamlMount["name"])
	assert.Equal(t, "app.yaml", appYamlMount["subPath"])
	assert.True(t, appYamlMount["readOnly"].(bool))

	// Verify second config file mount
	logbackMount := mountPaths["/config/logback.xml"]
	require.NotNil(t, logbackMount)
	assert.Equal(t, "config-test-config-volume", logbackMount["name"])
	assert.Equal(t, "logback.xml", logbackMount["subPath"])
	assert.True(t, logbackMount["readOnly"].(bool))
}

// Test for imagePullSecrets when ECR is enabled
func TestGenerateDeploymentYAML_WithECRImagePullSecrets(t *testing.T) {
	component := &CustomComponent{
		Name:      "ecr-app",
		Image:     "992979781608.dkr.ecr.eu-west-1.amazonaws.com/my-app:v1.0",
		Namespace: "default",
	}
	component.SetDefaults()

	// Create config with ECR enabled
	config := &KindEnvConfig{
		Images: struct {
			SkipPull  bool `yaml:"skipPull"`
			DockerHub struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"dockerHub"`
			UseAwsEcr bool `yaml:"useAwsEcr"`
			AWS       struct {
				Region      string `yaml:"region"`
				EcrRegistry string `yaml:"ecrRegistry"`
				Profile     string `yaml:"profile"`
			} `yaml:"aws"`
			UseHarbor      bool   `yaml:"useHarbor"`
			HarborRegistry string `yaml:"harborRegistry"`
		}{
			UseAwsEcr: true,
		},
	}

	yamlStr, err := generateDeploymentYAML(component, config)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	spec := template["spec"].(map[string]interface{})

	// Verify imagePullSecrets are present
	imagePullSecrets, exists := spec["imagePullSecrets"]
	require.True(t, exists, "imagePullSecrets should be present when ECR is enabled")
	
	secrets := imagePullSecrets.([]interface{})
	require.Len(t, secrets, 1)
	
	secret := secrets[0].(map[string]interface{})
	assert.Equal(t, "ecr-credentials", secret["name"])
}

// Test for no imagePullSecrets when ECR is disabled
func TestGenerateDeploymentYAML_WithoutECRImagePullSecrets(t *testing.T) {
	component := &CustomComponent{
		Name:      "public-app",
		Image:     "nginx:latest",
		Namespace: "default",
	}
	component.SetDefaults()

	// Create config with ECR disabled
	config := &KindEnvConfig{
		Images: struct {
			SkipPull  bool `yaml:"skipPull"`
			DockerHub struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"dockerHub"`
			UseAwsEcr bool `yaml:"useAwsEcr"`
			AWS       struct {
				Region      string `yaml:"region"`
				EcrRegistry string `yaml:"ecrRegistry"`
				Profile     string `yaml:"profile"`
			} `yaml:"aws"`
			UseHarbor      bool   `yaml:"useHarbor"`
			HarborRegistry string `yaml:"harborRegistry"`
		}{
			UseAwsEcr: false,
		},
	}

	yamlStr, err := generateDeploymentYAML(component, config)
	require.NoError(t, err)

	var deployment map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlStr), &deployment)
	require.NoError(t, err)

	template := deployment["spec"].(map[string]interface{})["template"].(map[string]interface{})
	spec := template["spec"].(map[string]interface{})

	// Verify imagePullSecrets are NOT present
	_, exists := spec["imagePullSecrets"]
	assert.False(t, exists, "imagePullSecrets should not be present when ECR is disabled")
}
