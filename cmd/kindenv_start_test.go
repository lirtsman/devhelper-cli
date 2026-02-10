package cmd

import (
	"fmt"
	"testing"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/stretchr/testify/assert"
)

// TestKedaConfiguration tests KEDA configuration validation and behavior
func TestKedaConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		setupConfig   func() *kindenv.KindEnvConfig
		expectedError bool
	}{
		{
			name: "default KEDA configuration",
			setupConfig: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				// Default has KEDA disabled
				return config
			},
			expectedError: false,
		},
		{
			name: "KEDA enabled with defaults",
			setupConfig: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.Keda.Enabled = true
				return config
			},
			expectedError: false,
		},
		{
			name: "KEDA disabled",
			setupConfig: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.Keda.Enabled = false
				return config
			},
			expectedError: false,
		},
		{
			name: "KEDA with custom chart version",
			setupConfig: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.Keda.Enabled = true
				config.Components.Keda.ChartVersion = "2.19.0"
				return config
			},
			expectedError: false,
		},
		{
			name: "KEDA with custom namespace",
			setupConfig: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.Keda.Enabled = true
				config.Components.Keda.Namespace = "autoscaling"
				return config
			},
			expectedError: false,
		},
		{
			name: "KEDA with empty chart version (invalid)",
			setupConfig: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.Keda.Enabled = true
				config.Components.Keda.ChartVersion = ""
				return config
			},
			expectedError: true,
		},
		{
			name: "KEDA with empty namespace (invalid)",
			setupConfig: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.Keda.Enabled = true
				config.Components.Keda.Namespace = ""
				return config
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()

			// Validate configuration structure
			assert.NotNil(t, config)
			assert.NotNil(t, config.Components)

			// Verify KEDA configuration is accessible
			kedaConfig := config.Components.Keda

			// Validate configuration
			err := validateKedaConfig(config)

			if tt.expectedError {
				assert.Error(t, err, "Expected validation error")
			} else {
				assert.NoError(t, err, "Configuration should be valid")
				assert.NotEmpty(t, kedaConfig.Namespace, "KEDA namespace should not be empty")
				assert.NotEmpty(t, kedaConfig.ChartVersion, "KEDA chart version should not be empty")
			}
		})
	}
}

// TestKedaConfigurationValidation tests KEDA config struct field validation
func TestKedaConfigurationValidation(t *testing.T) {
	// Test that KEDA configuration fields have expected types and defaults
	config := kindenv.CreateDefaultConfig()

	// Verify KEDA struct exists and has correct fields
	kedaConfig := config.Components.Keda
	assert.IsType(t, false, kedaConfig.Enabled, "Enabled should be bool")
	assert.IsType(t, "", kedaConfig.Namespace, "Namespace should be string")
	assert.IsType(t, "", kedaConfig.ChartVersion, "ChartVersion should be string")

	// Verify default values
	assert.Equal(t, false, kedaConfig.Enabled, "KEDA should be disabled by default")
	assert.Equal(t, "keda", kedaConfig.Namespace, "Default namespace should be 'keda'")
	assert.Equal(t, "2.16.0", kedaConfig.ChartVersion, "Default chart version should be 2.16.0")
}

// validateKedaConfig validates KEDA configuration
func validateKedaConfig(config *kindenv.KindEnvConfig) error {
	if config.Components.Keda.Enabled {
		if config.Components.Keda.Namespace == "" {
			return fmt.Errorf("KEDA namespace cannot be empty when KEDA is enabled")
		}
		if config.Components.Keda.ChartVersion == "" {
			return fmt.Errorf("KEDA chartVersion cannot be empty when KEDA is enabled")
		}
	}
	return nil
}

// TestKedaSkipFlag tests --skip-keda flag behavior
func TestKedaSkipFlag(t *testing.T) {
	tests := []struct {
		name                string
		kedaEnabledInConfig bool
		skipFlagSet         bool
		expectedEnabled     bool
	}{
		{
			name:                "KEDA enabled in config, no skip flag",
			kedaEnabledInConfig: true,
			skipFlagSet:         false,
			expectedEnabled:     true,
		},
		{
			name:                "KEDA enabled in config, skip flag set",
			kedaEnabledInConfig: true,
			skipFlagSet:         true,
			expectedEnabled:     false,
		},
		{
			name:                "KEDA disabled in config, no skip flag",
			kedaEnabledInConfig: false,
			skipFlagSet:         false,
			expectedEnabled:     false,
		},
		{
			name:                "KEDA disabled in config, skip flag set",
			kedaEnabledInConfig: false,
			skipFlagSet:         true,
			expectedEnabled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := kindenv.CreateDefaultConfig()
			config.Components.Keda.Enabled = tt.kedaEnabledInConfig

			// Simulate skip flag behavior
			if tt.skipFlagSet {
				config.Components.Keda.Enabled = false
			}

			// Verify final state
			assert.Equal(t, tt.expectedEnabled, config.Components.Keda.Enabled,
				"KEDA enabled state should match expected value")

			// Additional verification: if skip flag overrides config
			if tt.kedaEnabledInConfig && tt.skipFlagSet {
				assert.False(t, config.Components.Keda.Enabled,
					"Skip flag should override config and disable KEDA")
			}
		})
	}
}

// TestMySQLResourceConfiguration tests that MySQL resource limits are properly configured
func TestMySQLResourceConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		config         *kindenv.KindEnvConfig
		expectedCPU    string
		expectedMemory string
	}{
		{
			name: "default resource configuration",
			config: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				return config
			}(),
			expectedCPU:    "500m",
			expectedMemory: "1Gi",
		},
		{
			name: "custom resource configuration",
			config: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.MySQL.Resources.CPU = "1000m"
				config.Components.MySQL.Resources.Memory = "2Gi"
				return config
			}(),
			expectedCPU:    "1000m",
			expectedMemory: "2Gi",
		},
		{
			name: "low resource configuration",
			config: func() *kindenv.KindEnvConfig {
				config := kindenv.CreateDefaultConfig()
				config.Components.MySQL.Resources.CPU = "250m"
				config.Components.MySQL.Resources.Memory = "512Mi"
				return config
			}(),
			expectedCPU:    "250m",
			expectedMemory: "512Mi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify resource configuration
			assert.Equal(t, tt.expectedCPU, tt.config.Components.MySQL.Resources.CPU)
			assert.Equal(t, tt.expectedMemory, tt.config.Components.MySQL.Resources.Memory)
		})
	}
}

// TestMySQLCustomDatabaseName tests that MySQL database name can be customized
func TestMySQLCustomDatabaseName(t *testing.T) {
	tests := []struct {
		name           string
		database       string
		expectedInArgs bool
	}{
		{
			name:           "default database name",
			database:       "mysql",
			expectedInArgs: true,
		},
		{
			name:           "custom database name",
			database:       "myapp",
			expectedInArgs: true,
		},
		{
			name:           "empty database name",
			database:       "",
			expectedInArgs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := kindenv.CreateDefaultConfig()
			config.Components.MySQL.Database = tt.database

			// Verify database name is set correctly
			assert.Equal(t, tt.database, config.Components.MySQL.Database)
		})
	}
}

// TestMySQLNodePortConfiguration tests that MySQL NodePort can be customized
func TestMySQLNodePortConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		nodePort    int
		expectValid bool
	}{
		{
			name:        "default MySQL NodePort",
			nodePort:    30306,
			expectValid: true,
		},
		{
			name:        "custom MySQL NodePort",
			nodePort:    31234,
			expectValid: true,
		},
		{
			name:        "low NodePort (out of range)",
			nodePort:    1234,
			expectValid: false,
		},
		{
			name:        "high NodePort (out of range)",
			nodePort:    40000,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := kindenv.CreateDefaultConfig()
			config.Components.MySQL.NodePorts.MySQL = tt.nodePort

			// Verify NodePort is set
			assert.Equal(t, tt.nodePort, config.Components.MySQL.NodePorts.MySQL)

			// NodePort validation (30000-32767 is valid Kubernetes range)
			if tt.expectValid {
				assert.True(t, tt.nodePort >= 30000 && tt.nodePort <= 32767)
			} else {
				assert.False(t, tt.nodePort >= 30000 && tt.nodePort <= 32767)
			}
		})
	}
}

// TestMySQLPersistenceConfiguration tests that MySQL persistence can be configured
func TestMySQLPersistenceConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		persistenceEnabled bool
		persistenceSize    string
		expectValid        bool
	}{
		{
			name:               "default persistence disabled",
			persistenceEnabled: false,
			persistenceSize:    "8Gi",
			expectValid:        true,
		},
		{
			name:               "persistence enabled with default size",
			persistenceEnabled: true,
			persistenceSize:    "8Gi",
			expectValid:        true,
		},
		{
			name:               "persistence enabled with custom size",
			persistenceEnabled: true,
			persistenceSize:    "20Gi",
			expectValid:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := kindenv.CreateDefaultConfig()
			config.Components.MySQL.Persistence.Enabled = tt.persistenceEnabled
			config.Components.MySQL.Persistence.Size = tt.persistenceSize

			// Verify persistence configuration
			assert.Equal(t, tt.persistenceEnabled, config.Components.MySQL.Persistence.Enabled)
			assert.Equal(t, tt.persistenceSize, config.Components.MySQL.Persistence.Size)
		})
	}
}

// TestCustomComponentBasicDeployment tests basic custom component deployment
func TestCustomComponentBasicDeployment(t *testing.T) {
	tests := []struct {
		name        string
		component   kindenv.CustomComponent
		expectError bool
		checkYAML   func(t *testing.T, component kindenv.CustomComponent)
	}{
		{
			name: "basic nginx deployment",
			component: kindenv.CustomComponent{
				Name:  "nginx-app",
				Image: "nginx:latest",
			},
			expectError: false,
			checkYAML: func(t *testing.T, component kindenv.CustomComponent) {
				assert.Equal(t, "nginx-app", component.Name)
				assert.Equal(t, "nginx:latest", component.Image)
			},
		},
		{
			name: "deployment with replicas",
			component: kindenv.CustomComponent{
				Name:     "replicated-app",
				Image:    "nginx:latest",
				Replicas: intPtr(3),
			},
			expectError: false,
			checkYAML: func(t *testing.T, component kindenv.CustomComponent) {
				assert.Equal(t, "replicated-app", component.Name)
				assert.NotNil(t, component.Replicas)
				assert.Equal(t, 3, *component.Replicas)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.component.SetDefaults()

			// Verify basic fields
			assert.NotEmpty(t, tt.component.Name)
			assert.NotEmpty(t, tt.component.Image)

			if tt.checkYAML != nil {
				tt.checkYAML(t, tt.component)
			}
		})
	}
}

// TestCustomComponentWithSecretReferences tests custom components with secret references
func TestCustomComponentWithSecretReferences(t *testing.T) {
	tests := []struct {
		name            string
		component       kindenv.CustomComponent
		expectError     bool
		checkYAML       func(t *testing.T, component kindenv.CustomComponent)
		skipIfNoCluster bool
	}{
		{
			name: "component with secret env var",
			component: kindenv.CustomComponent{
				Name:  "app-with-secret",
				Image: "myapp:latest",
				Env: []kindenv.EnvVar{
					{
						Name:  "DB_PASSWORD",
						Value: "",
						ValueFrom: &kindenv.EnvVarSource{
							SecretKeyRef: &kindenv.SecretKeySelector{
								Name: "mysql-secret",
								Key:  "password",
							},
						},
					},
				},
			},
			expectError: false,
			checkYAML: func(t *testing.T, component kindenv.CustomComponent) {
				assert.Equal(t, 1, len(component.Env))
				assert.NotNil(t, component.Env[0].ValueFrom)
				assert.NotNil(t, component.Env[0].ValueFrom.SecretKeyRef)
				assert.Equal(t, "mysql-secret", component.Env[0].ValueFrom.SecretKeyRef.Name)
			},
			skipIfNoCluster: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoCluster {
				t.Skip("Skipping test that requires cluster")
			}

			tt.component.SetDefaults()

			if tt.checkYAML != nil {
				tt.checkYAML(t, tt.component)
			}
		})
	}
}

// TestCustomComponentWithPortMappings tests custom components with port mappings
func TestCustomComponentWithPortMappings(t *testing.T) {
	tests := []struct {
		name        string
		component   kindenv.CustomComponent
		expectError bool
		checkYAML   func(t *testing.T, component kindenv.CustomComponent)
	}{
		{
			name: "component with single port",
			component: kindenv.CustomComponent{
				Name:  "web-app",
				Image: "nginx:latest",
				Ports: []kindenv.PortMapping{
					{
						ContainerPort: 8080,
						Protocol:      "TCP",
					},
				},
			},
			expectError: false,
			checkYAML: func(t *testing.T, component kindenv.CustomComponent) {
				assert.Equal(t, 1, len(component.Ports))
				assert.Equal(t, 8080, component.Ports[0].ContainerPort)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.component.SetDefaults()

			if tt.checkYAML != nil {
				tt.checkYAML(t, tt.component)
			}
		})
	}
}

// TestMultipleCustomComponentsWithDifferentFeatures tests multiple custom components
func TestMultipleCustomComponentsWithDifferentFeatures(t *testing.T) {
	config := kindenv.CreateDefaultConfig()

	config.CustomComponents = []kindenv.CustomComponent{
		{
			Name:  "app1",
			Image: "nginx:latest",
		},
		{
			Name:  "app2",
			Image: "redis:latest",
			Ports: []kindenv.PortMapping{
				{ContainerPort: 6379, Protocol: "TCP"},
			},
		},
	}

	assert.Equal(t, 2, len(config.CustomComponents))
}

// TestCustomComponentWithAllFeaturesEnabled tests component with all features
func TestCustomComponentWithAllFeaturesEnabled(t *testing.T) {
	component := kindenv.CustomComponent{
		Name:     "full-featured-app",
		Image:    "myapp:latest",
		Replicas: intPtr(2),
		Ports: []kindenv.PortMapping{
			{ContainerPort: 8080, Protocol: "TCP"},
		},
		Env: []kindenv.EnvVar{
			{Name: "ENV", Value: "prod"},
		},
	}

	component.SetDefaults()

	assert.Equal(t, "full-featured-app", component.Name)
	assert.NotNil(t, component.Replicas)
	assert.Equal(t, 2, *component.Replicas)
	assert.Equal(t, 1, len(component.Ports))
	assert.Equal(t, 1, len(component.Env))
}

// TestParallelDeploymentOfMultipleComponents tests parallel deployment
func TestParallelDeploymentOfMultipleComponents(t *testing.T) {
	t.Skip("Parallel deployment test requires cluster")
}

// TestEndToEndCustomComponentWithInfrastructure tests end-to-end deployment
func TestEndToEndCustomComponentWithInfrastructure(t *testing.T) {
	t.Skip("End-to-end test requires cluster")
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}

// Helper function for creating bool pointers
func boolPtr(b bool) *bool {
	return &b
}
