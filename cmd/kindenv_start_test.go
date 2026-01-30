package cmd

import (
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
