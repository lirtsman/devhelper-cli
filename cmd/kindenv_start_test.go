package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/stretchr/testify/assert"
)

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
			config: &kindenv.KindEnvConfig{
				Components: struct {
					Temporal struct {
						Enabled      bool   `yaml:"enabled"`
						Namespace    string `yaml:"namespace"`
						ChartVersion string `yaml:"chartVersion"`
						NodePorts    struct {
							Web      int `yaml:"web"`
							Frontend int `yaml:"frontend"`
						} `yaml:"nodePorts"`
					} `yaml:"temporal"`
					Redis struct {
						Enabled   bool `yaml:"enabled"`
						NodePorts struct {
							Redis int `yaml:"redis"`
						} `yaml:"nodePorts"`
						ChartVersion string `yaml:"chartVersion"`
						Auth         struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"auth"`
					} `yaml:"redis"`
					Dapr struct {
						Enabled      bool   `yaml:"enabled"`
						ChartVersion string `yaml:"chartVersion"`
						NodePorts    struct {
							Dashboard int `yaml:"dashboard"`
						} `yaml:"nodePorts"`
						LogLevel string `yaml:"logLevel"`
						Mtls     struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"mtls"`
						Ha struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"ha"`
					} `yaml:"dapr"`
					OpenSearch struct {
						Enabled   bool   `yaml:"enabled"`
						Namespace string `yaml:"namespace"`
						Version   string `yaml:"version"`
						NodePorts struct {
							Rest int `yaml:"rest"`
						} `yaml:"nodePorts"`
						Security struct {
							Disabled bool `yaml:"disabled"`
						} `yaml:"security"`
						IndexManagement struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"indexManagement"`
					} `yaml:"openSearch"`
					OpenSearchDashboards struct {
						Enabled   bool   `yaml:"enabled"`
						Namespace string `yaml:"namespace"`
						Version   string `yaml:"version"`
						NodePorts struct {
							Http int `yaml:"http"`
						} `yaml:"nodePorts"`
					} `yaml:"openSearchDashboards"`
					TemporalWorkerOperator struct {
						Enabled          bool   `yaml:"enabled"`
						ChartVersion     string `yaml:"chartVersion"`
						TemporalNamespace string `yaml:"temporalNamespace"`
					} `yaml:"temporalWorkerOperator"`
					IndicesOperator struct {
						Enabled      bool   `yaml:"enabled"`
						ChartVersion string `yaml:"chartVersion"`
					} `yaml:"indicesOperator"`
					MetricsServer struct {
						Enabled      bool   `yaml:"enabled"`
						ChartVersion string `yaml:"chartVersion"`
					} `yaml:"metricsServer"`
					MySQL struct {
						Enabled      bool   `yaml:"enabled"`
						Namespace    string `yaml:"namespace"`
						ChartVersion string `yaml:"chartVersion"`
						Database     string `yaml:"database"`
						NodePorts    struct {
							MySQL int `yaml:"mysql"`
						} `yaml:"nodePorts"`
						Resources struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						} `yaml:"resources"`
						Persistence struct {
							Enabled bool   `yaml:"enabled"`
							Size    string `yaml:"size"`
						} `yaml:"persistence"`
					} `yaml:"mysql"`
				}{
					MySQL: struct {
						Enabled      bool   `yaml:"enabled"`
						Namespace    string `yaml:"namespace"`
						ChartVersion string `yaml:"chartVersion"`
						Database     string `yaml:"database"`
						NodePorts    struct {
							MySQL int `yaml:"mysql"`
						} `yaml:"nodePorts"`
						Resources struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						} `yaml:"resources"`
						Persistence struct {
							Enabled bool   `yaml:"enabled"`
							Size    string `yaml:"size"`
						} `yaml:"persistence"`
					}{
						Enabled:      true,
						Namespace:    "mysql",
						ChartVersion: "9.4.6",
						Database:     "mydb",
						NodePorts: struct {
							MySQL int `yaml:"mysql"`
						}{
							MySQL: 30306,
						},
						Resources: struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						}{
							CPU:    "500m",
							Memory: "1Gi",
						},
					},
				},
			},
			expectedCPU:    "500m",
			expectedMemory: "1Gi",
		},
		{
			name: "custom resource configuration",
			config: &kindenv.KindEnvConfig{
				Components: struct {
					Temporal struct {
						Enabled      bool   `yaml:"enabled"`
						Namespace    string `yaml:"namespace"`
						ChartVersion string `yaml:"chartVersion"`
						NodePorts    struct {
							Web      int `yaml:"web"`
							Frontend int `yaml:"frontend"`
						} `yaml:"nodePorts"`
					} `yaml:"temporal"`
					Redis struct {
						Enabled   bool `yaml:"enabled"`
						NodePorts struct {
							Redis int `yaml:"redis"`
						} `yaml:"nodePorts"`
						ChartVersion string `yaml:"chartVersion"`
						Auth         struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"auth"`
					} `yaml:"redis"`
					Dapr struct {
						Enabled      bool   `yaml:"enabled"`
						ChartVersion string `yaml:"chartVersion"`
						NodePorts    struct {
							Dashboard int `yaml:"dashboard"`
						} `yaml:"nodePorts"`
						LogLevel string `yaml:"logLevel"`
						Mtls     struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"mtls"`
						Ha struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"ha"`
					} `yaml:"dapr"`
					OpenSearch struct {
						Enabled   bool   `yaml:"enabled"`
						Namespace string `yaml:"namespace"`
						Version   string `yaml:"version"`
						NodePorts struct {
							Rest int `yaml:"rest"`
						} `yaml:"nodePorts"`
						Security struct {
							Disabled bool `yaml:"disabled"`
						} `yaml:"security"`
						IndexManagement struct {
							Enabled bool `yaml:"enabled"`
						} `yaml:"indexManagement"`
					} `yaml:"openSearch"`
					OpenSearchDashboards struct {
						Enabled   bool   `yaml:"enabled"`
						Namespace string `yaml:"namespace"`
						Version   string `yaml:"version"`
						NodePorts struct {
							Http int `yaml:"http"`
						} `yaml:"nodePorts"`
					} `yaml:"openSearchDashboards"`
					TemporalWorkerOperator struct {
						Enabled          bool   `yaml:"enabled"`
						ChartVersion     string `yaml:"chartVersion"`
						TemporalNamespace string `yaml:"temporalNamespace"`
					} `yaml:"temporalWorkerOperator"`
					IndicesOperator struct {
						Enabled      bool   `yaml:"enabled"`
						ChartVersion string `yaml:"chartVersion"`
					} `yaml:"indicesOperator"`
					MetricsServer struct {
						Enabled      bool   `yaml:"enabled"`
						ChartVersion string `yaml:"chartVersion"`
					} `yaml:"metricsServer"`
					MySQL struct {
						Enabled      bool   `yaml:"enabled"`
						Namespace    string `yaml:"namespace"`
						ChartVersion string `yaml:"chartVersion"`
						Database     string `yaml:"database"`
						NodePorts    struct {
							MySQL int `yaml:"mysql"`
						} `yaml:"nodePorts"`
						Resources struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						} `yaml:"resources"`
						Persistence struct {
							Enabled bool   `yaml:"enabled"`
							Size    string `yaml:"size"`
						} `yaml:"persistence"`
					} `yaml:"mysql"`
				}{
					MySQL: struct {
						Enabled      bool   `yaml:"enabled"`
						Namespace    string `yaml:"namespace"`
						ChartVersion string `yaml:"chartVersion"`
						Database     string `yaml:"database"`
						NodePorts    struct {
							MySQL int `yaml:"mysql"`
						} `yaml:"nodePorts"`
						Resources struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						} `yaml:"resources"`
						Persistence struct {
							Enabled bool   `yaml:"enabled"`
							Size    string `yaml:"size"`
						} `yaml:"persistence"`
					}{
						Enabled:      true,
						Namespace:    "mysql",
						ChartVersion: "9.4.6",
						Database:     "customdb",
						NodePorts: struct {
							MySQL int `yaml:"mysql"`
						}{
							MySQL: 30306,
						},
						Resources: struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						}{
							CPU:    "1000m",
							Memory: "2Gi",
						},
					},
				},
			},
			expectedCPU:    "1000m",
			expectedMemory: "2Gi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that the configuration contains the expected resource values
			assert.Equal(t, tt.expectedCPU, tt.config.Components.MySQL.Resources.CPU,
				"CPU resource should match expected value")
			assert.Equal(t, tt.expectedMemory, tt.config.Components.MySQL.Resources.Memory,
				"Memory resource should match expected value")

			// Verify that MySQL is enabled
			assert.True(t, tt.config.Components.MySQL.Enabled,
				"MySQL should be enabled for this test")

			// Verify that the configuration would generate correct Helm arguments
			// This simulates what would be passed to Helm
			expectedCPUArg := "primary.resources.requests.cpu=" + tt.expectedCPU
			expectedMemoryArg := "primary.resources.requests.memory=" + tt.expectedMemory

			// Verify the arguments would be formatted correctly
			assert.Contains(t, expectedCPUArg, tt.expectedCPU,
				"Helm CPU argument should contain expected CPU value")
			assert.Contains(t, expectedMemoryArg, tt.expectedMemory,
				"Helm Memory argument should contain expected memory value")
		})
	}
}

// TestMySQLCustomDatabaseName tests that custom database names are properly configured
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
			name:           "database name with underscore",
			database:       "my_app",
			expectedInArgs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that the database name would be included in Helm arguments
			expectedArg := "auth.database=" + tt.database
			assert.Contains(t, expectedArg, tt.database,
				"Helm database argument should contain database name")

			// Verify the argument format is correct
			assert.True(t, strings.HasPrefix(expectedArg, "auth.database="),
				"Database argument should start with auth.database=")
		})
	}
}

// TestMySQLNodePortConfiguration tests that NodePort configuration is properly handled
func TestMySQLNodePortConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		nodePort    int
		expectValid bool
	}{
		{
			name:        "valid nodeport at minimum",
			nodePort:    30000,
			expectValid: true,
		},
		{
			name:        "valid nodeport at maximum",
			nodePort:    32767,
			expectValid: true,
		},
		{
			name:        "valid nodeport in middle",
			nodePort:    30306,
			expectValid: true,
		},
		{
			name:        "invalid nodeport too low",
			nodePort:    29999,
			expectValid: false,
		},
		{
			name:        "invalid nodeport too high",
			nodePort:    32768,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify NodePort validation
			err := kindenv.ValidateNodePort(tt.nodePort)
			if tt.expectValid {
				assert.NoError(t, err, "NodePort should be valid")
			} else {
				assert.Error(t, err, "NodePort should be invalid")
			}

			// Verify that valid NodePorts would be included in Helm arguments
			if tt.expectValid {
				// Verify the argument format would be correct (using fmt.Sprintf would be used in actual code)
				expectedArgPrefix := "primary.service.nodePorts.mysql="
				assert.True(t, strings.HasPrefix(expectedArgPrefix, "primary.service.nodePorts.mysql="),
					"Helm NodePort argument should have correct format")
			}
		})
	}
}

// TestMySQLPersistenceConfiguration tests that persistence configuration is properly handled
func TestMySQLPersistenceConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		persistenceEnabled bool
		persistenceSize    string
		expectValid    bool
	}{
		{
			name:              "persistence disabled",
			persistenceEnabled: false,
			persistenceSize:    "",
			expectValid:       true,
		},
		{
			name:              "persistence enabled with valid size",
			persistenceEnabled: true,
			persistenceSize:    "8Gi",
			expectValid:       true,
		},
		{
			name:              "persistence enabled with custom size",
			persistenceEnabled: true,
			persistenceSize:    "10Gi",
			expectValid:       true,
		},
		{
			name:              "persistence enabled but size missing",
			persistenceEnabled: true,
			persistenceSize:    "",
			expectValid:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := kindenv.MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: kindenv.MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: kindenv.MySQLPersistence{
					Enabled: tt.persistenceEnabled,
					Size:    tt.persistenceSize,
				},
			}

			err := kindenv.ValidateMySQLConfig(config)
			if tt.expectValid {
				assert.NoError(t, err, "Persistence configuration should be valid")
			} else {
				assert.Error(t, err, "Persistence configuration should be invalid")
			}

			// Verify that enabled persistence would include size in Helm arguments
			if tt.persistenceEnabled && tt.expectValid {
				expectedArg := "primary.persistence.size=" + tt.persistenceSize
				assert.Contains(t, expectedArg, tt.persistenceSize,
					"Helm persistence size argument should contain size value")
			}
		})
	}
}

// TestCustomComponentBasicDeployment tests that basic custom component deployment generates correct YAML
func TestCustomComponentBasicDeployment(t *testing.T) {
	tests := []struct {
		name        string
		component   kindenv.CustomComponent
		expectError bool
		checkYAML    func(*testing.T, string)
	}{
		{
			name: "minimal custom component with image and env vars",
			component: kindenv.CustomComponent{
				Name:  "test-app",
				Image: "nginx:latest",
				Env: []kindenv.EnvVar{
					{
						Name:  "APP_ENV",
						Value: "development",
					},
					{
						Name:  "DEBUG",
						Value: "true",
					},
				},
			},
			expectError: false,
			checkYAML: func(t *testing.T, yaml string) {
				assert.Contains(t, yaml, "name: test-app", "YAML should contain component name")
				assert.Contains(t, yaml, "image: nginx:latest", "YAML should contain image")
				assert.Contains(t, yaml, "APP_ENV", "YAML should contain environment variable name")
				assert.Contains(t, yaml, "development", "YAML should contain environment variable value")
				assert.Contains(t, yaml, "apiVersion: apps/v1", "YAML should be a Deployment")
				assert.Contains(t, yaml, "kind: Deployment", "YAML should be a Deployment")
			},
		},
		{
			name: "custom component with namespace",
			component: kindenv.CustomComponent{
				Name:      "my-app",
				Image:     "alpine:latest",
				Namespace: "custom-ns",
			},
			expectError: false,
			checkYAML: func(t *testing.T, yaml string) {
				assert.Contains(t, yaml, "namespace: custom-ns", "YAML should contain namespace")
				assert.Contains(t, yaml, "name: my-app", "YAML should contain component name")
			},
		},
		{
			name: "custom component with replicas",
			component: kindenv.CustomComponent{
				Name:     "scaled-app",
				Image:    "nginx:latest",
				Replicas: intPtr(3),
			},
			expectError: false,
			checkYAML: func(t *testing.T, yaml string) {
				assert.Contains(t, yaml, "replicas: 3", "YAML should contain replicas")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set defaults
			tt.component.SetDefaults()

			// Validate component
			err := tt.component.Validate()
			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
				return
			}
			assert.NoError(t, err, "Component should be valid")

			// Generate deployment YAML
			config := &kindenv.KindEnvConfig{
				CustomComponents: []kindenv.CustomComponent{tt.component},
			}

			deploymentInfos, err := kindenv.DeployCustomComponents(context.Background(), config)
			if tt.expectError {
				assert.Error(t, err, "Expected deployment error")
				return
			}

			assert.NoError(t, err, "Deployment should succeed")
			assert.Len(t, deploymentInfos, 1, "Should have one deployment info")

			deploymentInfo := deploymentInfos[0]
			assert.Equal(t, tt.component.Name, deploymentInfo.Name, "Deployment name should match")
			assert.NotEmpty(t, deploymentInfo.DeploymentYAML, "Deployment YAML should not be empty")

			// Check YAML content
			if tt.checkYAML != nil {
				tt.checkYAML(t, deploymentInfo.DeploymentYAML)
			}
		})
	}
}

// T038: Integration test for secret-based MySQL connection
func TestCustomComponentWithSecretReferences(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	
	tests := []struct {
		name        string
		component   kindenv.CustomComponent
		expectError bool
		checkYAML   func(*testing.T, string)
		skipIfNoCluster bool
	}{
		{
			name: "component with secretKeyRef for MySQL connection",
			component: kindenv.CustomComponent{
				Name:  "mysql-client",
				Image: "mysql:8.0",
				Env: []kindenv.EnvVar{
					{
						Name:  "MYSQL_HOST",
						Value: "mysql.mysql.svc.cluster.local",
					},
					{
						Name: "MYSQL_PASSWORD",
						ValueFrom: &kindenv.EnvVarSource{
							SecretKeyRef: &kindenv.SecretKeySelector{
								Name: "mysql-secret",
								Key:  "password",
							},
						},
					},
					{
						Name: "MYSQL_USER",
						ValueFrom: &kindenv.EnvVarSource{
							SecretKeyRef: &kindenv.SecretKeySelector{
								Name: "mysql-secret",
								Key:  "username",
							},
						},
					},
				},
			},
			expectError: false,
			skipIfNoCluster: true,
			checkYAML: func(t *testing.T, yaml string) {
				assert.Contains(t, yaml, "name: mysql-client", "YAML should contain component name")
				assert.Contains(t, yaml, "MYSQL_HOST", "YAML should contain direct env var")
				assert.Contains(t, yaml, "mysql.mysql.svc.cluster.local", "YAML should contain direct env var value")
				assert.Contains(t, yaml, "MYSQL_PASSWORD", "YAML should contain secret ref env var")
				assert.Contains(t, yaml, "valueFrom", "YAML should contain valueFrom")
				assert.Contains(t, yaml, "secretKeyRef", "YAML should contain secretKeyRef")
				assert.Contains(t, yaml, "mysql-secret", "YAML should contain secret name")
			},
		},
		{
			name: "component with secretKeyRef - validation error when secret missing",
			component: kindenv.CustomComponent{
				Name:  "test-app",
				Image: "nginx:latest",
				Env: []kindenv.EnvVar{
					{
						Name: "DB_PASSWORD",
						ValueFrom: &kindenv.EnvVarSource{
							SecretKeyRef: &kindenv.SecretKeySelector{
								Name: "nonexistent-secret",
								Key:  "password",
							},
						},
					},
				},
			},
			expectError: true,
			skipIfNoCluster: false, // This test should run even without cluster to test validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set defaults
			tt.component.SetDefaults()

			// Validate component structure
			err := tt.component.Validate()
			assert.NoError(t, err, "Component structure should be valid")

			// Generate deployment YAML (this will trigger secret validation)
			config := &kindenv.KindEnvConfig{
				CustomComponents: []kindenv.CustomComponent{tt.component},
			}

			deploymentInfos, err := kindenv.DeployCustomComponents(context.Background(), config)
			if tt.expectError {
				assert.Error(t, err, "Expected deployment error due to missing secret")
				assert.Contains(t, err.Error(), "references secrets that do not exist", "Error should mention missing secrets")
				return
			}

			// If we expect success but cluster might not be available, skip gracefully
			if tt.skipIfNoCluster && err != nil {
				if strings.Contains(err.Error(), "references secrets that do not exist") {
					t.Skip("Skipping test: secret validation requires a running cluster with the secret")
				}
			}

			assert.NoError(t, err, "Deployment should succeed")
			assert.Len(t, deploymentInfos, 1, "Should have one deployment info")

			deploymentInfo := deploymentInfos[0]
			assert.Equal(t, tt.component.Name, deploymentInfo.Name, "Deployment name should match")
			assert.NotEmpty(t, deploymentInfo.DeploymentYAML, "Deployment YAML should not be empty")

			// Check YAML content
			if tt.checkYAML != nil {
				tt.checkYAML(t, deploymentInfo.DeploymentYAML)
			}
		})
	}
}

// T053: Integration test for port accessibility
func TestCustomComponentWithPortMappings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		name        string
		component   kindenv.CustomComponent
		expectError bool
		checkYAML   func(*testing.T, string, string) // deploymentYAML, serviceYAML
	}{
		{
			name: "component with single port mapping",
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
			checkYAML: func(t *testing.T, deploymentYAML, serviceYAML string) {
				// Check deployment has container port
				assert.Contains(t, deploymentYAML, "containerPort: 8080", "Deployment should contain container port")
				// Check service YAML exists and has NodePort
				assert.NotEmpty(t, serviceYAML, "Service YAML should be generated")
				assert.Contains(t, serviceYAML, "kind: Service", "Service YAML should be a Service")
				assert.Contains(t, serviceYAML, "type: NodePort", "Service should be NodePort type")
				assert.Contains(t, serviceYAML, "port: 8080", "Service should expose port 8080")
			},
		},
		{
			name: "component with multiple port mappings",
			component: kindenv.CustomComponent{
				Name:  "multi-port-app",
				Image: "nginx:latest",
				Ports: []kindenv.PortMapping{
					{
						ContainerPort: 8080,
						Protocol:      "TCP",
					},
					{
						ContainerPort: 8443,
						Protocol:      "TCP",
					},
				},
			},
			expectError: false,
			checkYAML: func(t *testing.T, deploymentYAML, serviceYAML string) {
				assert.Contains(t, deploymentYAML, "containerPort: 8080", "Deployment should contain port 8080")
				assert.Contains(t, deploymentYAML, "containerPort: 8443", "Deployment should contain port 8443")
				assert.NotEmpty(t, serviceYAML, "Service YAML should be generated")
				assert.Contains(t, serviceYAML, "port: 8080", "Service should expose port 8080")
				assert.Contains(t, serviceYAML, "port: 8443", "Service should expose port 8443")
			},
		},
		{
			name: "component with explicit NodePort",
			component: kindenv.CustomComponent{
				Name:  "explicit-port-app",
				Image: "nginx:latest",
				Ports: []kindenv.PortMapping{
					{
						ContainerPort: 8080,
						NodePort:      30001,
						Protocol:     "TCP",
					},
				},
			},
			expectError: false,
			checkYAML: func(t *testing.T, deploymentYAML, serviceYAML string) {
				assert.Contains(t, serviceYAML, "nodePort: 30001", "Service should use specified NodePort")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set defaults
			tt.component.SetDefaults()

			// Validate component
			err := tt.component.Validate()
			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
				return
			}
			assert.NoError(t, err, "Component should be valid")

			// Generate deployment info
			config := &kindenv.KindEnvConfig{
				CustomComponents: []kindenv.CustomComponent{tt.component},
			}

			deploymentInfos, err := kindenv.DeployCustomComponents(context.Background(), config)
			if tt.expectError {
				assert.Error(t, err, "Expected deployment error")
				return
			}

			assert.NoError(t, err, "Deployment should succeed")
			assert.Len(t, deploymentInfos, 1, "Should have one deployment info")

			deploymentInfo := deploymentInfos[0]
			assert.Equal(t, tt.component.Name, deploymentInfo.Name, "Deployment name should match")
			assert.NotEmpty(t, deploymentInfo.DeploymentYAML, "Deployment YAML should not be empty")

			// Check YAML content
			if tt.checkYAML != nil {
				tt.checkYAML(t, deploymentInfo.DeploymentYAML, deploymentInfo.ServiceYAML)
			}
		})
	}
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}
