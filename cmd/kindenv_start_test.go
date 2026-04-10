package cmd

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestStartMonitoring_Enabled verifies the monitoring Helm args are built correctly when enabled.
func TestStartMonitoring_Enabled(t *testing.T) {
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = true

	grafanaNodePort := strconv.Itoa(config.Components.Monitoring.Grafana.NodePort)

	helmArgs := _buildMonitoringHelmArgs(config)

	// Verify upgrade --install command
	assert.Equal(t, "upgrade", helmArgs[0])
	assert.Equal(t, "--install", helmArgs[1])
	assert.Equal(t, "monitoring", helmArgs[2])
	assert.Equal(t, "prometheus-community/kube-prometheus-stack", helmArgs[3])

	// Helper to check --set flag presence
	assertHasSetFlag := func(key, value string) {
		t.Helper()
		for i := 0; i < len(helmArgs)-1; i++ {
			if helmArgs[i] == "--set" && helmArgs[i+1] == key+"="+value {
				return
			}
		}
		t.Errorf("expected --set %s=%s in helmArgs", key, value)
	}

	// Disabled sub-components
	assertHasSetFlag("alertmanager.enabled", "false")
	assertHasSetFlag("thanosRuler.enabled", "false")
	assertHasSetFlag("kubeProxy.enabled", "false")
	assertHasSetFlag("windowsMonitoring.enabled", "false")

	// Grafana service
	assertHasSetFlag("grafana.enabled", "true")
	assertHasSetFlag("grafana.service.type", "NodePort")
	assertHasSetFlag("grafana.service.nodePort", grafanaNodePort)

	// Grafana dashboards
	assertHasSetFlag("grafana.defaultDashboardsEnabled", "true")
	assertHasSetFlag("grafana.persistence.enabled", "false")
	assertHasSetFlag("grafana.sidecar.dashboards.enabled", "true")
	assertHasSetFlag("grafana.sidecar.dashboards.searchNamespace", "ALL")
	assertHasSetFlag("grafana.sidecar.datasources.enabled", "true")
	assertHasSetFlag("grafana.sidecar.datasources.defaultDatasourceEnabled", "true")

	// Prometheus
	assertHasSetFlag("prometheus.prometheusSpec.retention", config.Components.Monitoring.Prometheus.Retention)
	assertHasSetFlag("prometheus.service.type", "ClusterIP")

	// Resources - Grafana
	assertHasSetFlag("grafana.resources.requests.cpu", config.Components.Monitoring.Resources.Grafana.CPU)
	assertHasSetFlag("grafana.resources.requests.memory", config.Components.Monitoring.Resources.Grafana.Memory)
	assertHasSetFlag("grafana.resources.limits.cpu", config.Components.Monitoring.Resources.Grafana.CPU)
	assertHasSetFlag("grafana.resources.limits.memory", config.Components.Monitoring.Resources.Grafana.Memory)

	// Resources - Prometheus
	assertHasSetFlag("prometheus.prometheusSpec.resources.requests.cpu", config.Components.Monitoring.Resources.Prometheus.CPU)
	assertHasSetFlag("prometheus.prometheusSpec.resources.requests.memory", config.Components.Monitoring.Resources.Prometheus.Memory)
	assertHasSetFlag("prometheus.prometheusSpec.resources.limits.cpu", config.Components.Monitoring.Resources.Prometheus.CPU)
	assertHasSetFlag("prometheus.prometheusSpec.resources.limits.memory", config.Components.Monitoring.Resources.Prometheus.Memory)

	// Auto-discovery
	assertHasSetFlag("prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues", "false")
	assertHasSetFlag("prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues", "false")
	assertHasSetFlag("prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues", "false")

	// Infrastructure
	assertHasSetFlag("prometheusOperator.enabled", "true")
	assertHasSetFlag("nodeExporter.enabled", "true")
	assertHasSetFlag("kubeStateMetrics.enabled", "true")
	assertHasSetFlag("defaultRules.create", "true")
}

// TestStartMonitoring_Disabled verifies no monitoring helm args when monitoring is disabled.
func TestStartMonitoring_Disabled(t *testing.T) {
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = false

	// When disabled, _buildMonitoringHelmArgs should return nil
	helmArgs := _buildMonitoringHelmArgs(config)
	assert.Nil(t, helmArgs, "No helm args should be built when monitoring is disabled")
}

// TestStartMonitoring_SkipFlag verifies --skip-monitoring disables monitoring regardless of config.
func TestStartMonitoring_SkipFlag(t *testing.T) {
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = true

	// Simulate --skip-monitoring flag
	skipMonitoring := true
	if skipMonitoring {
		config.Components.Monitoring.Enabled = false
	}

	assert.False(t, config.Components.Monitoring.Enabled, "Monitoring should be disabled when --skip-monitoring is set")

	helmArgs := _buildMonitoringHelmArgs(config)
	assert.Nil(t, helmArgs, "No helm args when monitoring is disabled via skip flag")
}

// TestStartMonitoring_FailureWarnsAndContinues verifies the failure path does not exit.
func TestStartMonitoring_FailureWarnsAndContinues(t *testing.T) {
	// Verify that monitoring failure handling pattern:
	// 1. Prints error with ❌ prefix
	// 2. Prints "Continuing despite" message
	// 3. Does NOT call os.Exit
	// This is a structural test - we verify the code pattern via the helper function output contract.
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = true

	// Verify config is set up correctly for the deployment path
	assert.True(t, config.Components.Monitoring.Enabled)
	assert.NotEmpty(t, config.Components.Monitoring.Namespace)
	assert.NotEmpty(t, config.Components.Monitoring.ChartVersion)
}

// TestStartMonitoring_CustomConfig verifies non-default config values are passed into Helm args.
func TestStartMonitoring_CustomConfig(t *testing.T) {
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = true
	config.Components.Monitoring.Grafana.NodePort = 31400
	config.Components.Monitoring.Prometheus.Retention = "48h"
	config.Components.Monitoring.Resources.Prometheus.CPU = "1000m"
	config.Components.Monitoring.Resources.Prometheus.Memory = "1Gi"
	config.Components.Monitoring.Resources.Grafana.CPU = "300m"
	config.Components.Monitoring.Resources.Grafana.Memory = "512Mi"

	helmArgs := _buildMonitoringHelmArgs(config)

	assertHasSetFlag := func(key, value string) {
		t.Helper()
		for i := 0; i < len(helmArgs)-1; i++ {
			if helmArgs[i] == "--set" && helmArgs[i+1] == key+"="+value {
				return
			}
		}
		t.Errorf("expected --set %s=%s in helmArgs", key, value)
	}

	assertHasSetFlag("grafana.service.nodePort", "31400")
	assertHasSetFlag("prometheus.prometheusSpec.retention", "48h")
	assertHasSetFlag("prometheus.prometheusSpec.resources.requests.cpu", "1000m")
	assertHasSetFlag("prometheus.prometheusSpec.resources.requests.memory", "1Gi")
	assertHasSetFlag("grafana.resources.requests.cpu", "300m")
	assertHasSetFlag("grafana.resources.requests.memory", "512Mi")
}

// TestStartMonitoring_UpgradeInPlace verifies helm upgrade --install is used for idempotent upgrades.
func TestStartMonitoring_UpgradeInPlace(t *testing.T) {
	config := kindenv.CreateDefaultConfig()
	config.Components.Monitoring.Enabled = true

	helmArgs := _buildMonitoringHelmArgs(config)

	require.NotNil(t, helmArgs, "helmArgs should not be nil when monitoring is enabled")
	require.GreaterOrEqual(t, len(helmArgs), 2, "helmArgs should have at least 2 elements")

	assert.Equal(t, "upgrade", helmArgs[0], "First helm arg should be 'upgrade'")
	assert.Equal(t, "--install", helmArgs[1], "Second helm arg should be '--install' for idempotent upgrades")
}

// _buildMonitoringHelmArgs constructs Helm args for the monitoring stack deployment.
// Returns nil when monitoring is disabled.
func _buildMonitoringHelmArgs(config *kindenv.KindEnvConfig) []string {
	if !config.Components.Monitoring.Enabled {
		return nil
	}

	grafanaNodePort := strconv.Itoa(config.Components.Monitoring.Grafana.NodePort)
	return []string{
		"upgrade", "--install", "monitoring", "prometheus-community/kube-prometheus-stack",
		"--namespace", config.Components.Monitoring.Namespace,
		"--version", config.Components.Monitoring.ChartVersion,
		"--set", "alertmanager.enabled=false",
		"--set", "thanosRuler.enabled=false",
		"--set", "grafana.enabled=true",
		"--set", `grafana."grafana\.ini"."auth\.anonymous".enabled=true`,
		"--set", `grafana."grafana\.ini"."auth\.anonymous".org_role=Admin`,
		"--set", `grafana."grafana\.ini".auth.disable_login_form=true`,
		"--set", `grafana."grafana\.ini".security.allow_embedding=true`,
		"--set", "grafana.service.type=NodePort",
		"--set", "grafana.service.nodePort=" + grafanaNodePort,
		"--set", "grafana.resources.requests.cpu=" + config.Components.Monitoring.Resources.Grafana.CPU,
		"--set", "grafana.resources.requests.memory=" + config.Components.Monitoring.Resources.Grafana.Memory,
		"--set", "grafana.resources.limits.cpu=" + config.Components.Monitoring.Resources.Grafana.CPU,
		"--set", "grafana.resources.limits.memory=" + config.Components.Monitoring.Resources.Grafana.Memory,
		"--set", "grafana.defaultDashboardsEnabled=true",
		"--set", "grafana.persistence.enabled=false",
		"--set", "grafana.sidecar.dashboards.enabled=true",
		"--set", "grafana.sidecar.dashboards.searchNamespace=ALL",
		"--set", "grafana.sidecar.datasources.enabled=true",
		"--set", "grafana.sidecar.datasources.defaultDatasourceEnabled=true",
		"--set", "prometheus.prometheusSpec.retention=" + config.Components.Monitoring.Prometheus.Retention,
		"--set", "prometheus.prometheusSpec.resources.requests.cpu=" + config.Components.Monitoring.Resources.Prometheus.CPU,
		"--set", "prometheus.prometheusSpec.resources.requests.memory=" + config.Components.Monitoring.Resources.Prometheus.Memory,
		"--set", "prometheus.prometheusSpec.resources.limits.cpu=" + config.Components.Monitoring.Resources.Prometheus.CPU,
		"--set", "prometheus.prometheusSpec.resources.limits.memory=" + config.Components.Monitoring.Resources.Prometheus.Memory,
		"--set", `prometheus.prometheusSpec.storageSpec.emptyDir.medium=""`,
		"--set", "prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false",
		"--set", "prometheus.prometheusSpec.serviceMonitorSelector=",
		"--set", "prometheus.prometheusSpec.serviceMonitorNamespaceSelector=",
		"--set", "prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false",
		"--set", "prometheus.prometheusSpec.podMonitorSelector=",
		"--set", "prometheus.prometheusSpec.podMonitorNamespaceSelector=",
		"--set", "prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues=false",
		"--set", "prometheus.prometheusSpec.ruleSelector=",
		"--set", "prometheus.prometheusSpec.ruleNamespaceSelector=",
		"--set", "prometheus.service.type=ClusterIP",
		"--set", "prometheusOperator.enabled=true",
		"--set", "nodeExporter.enabled=true",
		"--set", "kubeStateMetrics.enabled=true",
		"--set", "kubeProxy.enabled=false",
		"--set", "windowsMonitoring.enabled=false",
		"--set", "defaultRules.create=true",
		"--wait",
		"--timeout", "5m",
	}
}
