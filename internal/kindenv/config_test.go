package kindenv

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateMySQLConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      MySQLConfig
		expectError bool
		errorField  string
	}{
		{
			name: "valid config with all fields",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Resources: MySQLResources{
					CPU:    "500m",
					Memory: "1Gi",
				},
				Persistence: MySQLPersistence{
					Enabled: false,
					Size:    "8Gi",
				},
			},
			expectError: false,
		},
		{
			name: "disabled MySQL should skip validation",
			config: MySQLConfig{
				Enabled: false,
			},
			expectError: false,
		},
		{
			name: "missing chart version",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
			},
			expectError: true,
			errorField:  "chartVersion",
		},
		{
			name: "missing database name",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
			},
			expectError: true,
			errorField:  "database",
		},
		{
			name: "invalid nodeport too low",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 25000,
				},
			},
			expectError: true,
			errorField:  "nodePorts.mysql",
		},
		{
			name: "invalid nodeport too high",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 40000,
				},
			},
			expectError: true,
			errorField:  "nodePorts.mysql",
		},
		{
			name: "invalid CPU format",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Resources: MySQLResources{
					CPU:    "invalid",
					Memory: "1Gi",
				},
			},
			expectError: true,
			errorField:  "resources.cpu",
		},
		{
			name: "invalid memory format",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Resources: MySQLResources{
					CPU:    "500m",
					Memory: "invalid",
				},
			},
			expectError: true,
			errorField:  "resources.memory",
		},
		{
			name: "persistence enabled but size missing",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "persistence enabled with invalid size format",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "invalid",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "valid config with persistence enabled",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Resources: MySQLResources{
					CPU:    "500m",
					Memory: "1Gi",
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "10Gi",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMySQLConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorField != "" {
					validationErr, ok := err.(*ValidationError)
					if !ok {
						t.Errorf("expected ValidationError but got %T", err)
						return
					}
					if validationErr.Field != tt.errorField {
						t.Errorf("expected error field %s but got %s", tt.errorField, validationErr.Field)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateRabbitMQConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      RabbitMQConfig
		expectError bool
		errorField  string
	}{
		{
			name: "valid config with all fields",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
				Resources: RabbitMQResources{
					CPU:    "500m",
					Memory: "1Gi",
				},
				Persistence: RabbitMQPersistence{
					Enabled: false,
					Size:    "8Gi",
				},
			},
			expectError: false,
		},
		{
			name: "disabled RabbitMQ should skip validation",
			config: RabbitMQConfig{
				Enabled: false,
			},
			expectError: false,
		},
		{
			name: "missing chart version",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
			},
			expectError: true,
			errorField:  "chartVersion",
		},
		{
			name: "invalid virtual host format",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "invalid-vhost",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
			},
			expectError: true,
			errorField:  "virtualHost",
		},
		{
			name: "valid virtual host starting with slash",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/dev",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
			},
			expectError: false,
		},
		{
			name: "invalid amqp nodeport too low",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       25000,
					Management: 31672,
				},
			},
			expectError: true,
			errorField:  "nodePorts.amqp",
		},
		{
			name: "invalid amqp nodeport too high",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       40000,
					Management: 31672,
				},
			},
			expectError: true,
			errorField:  "nodePorts.amqp",
		},
		{
			name: "invalid management nodeport too low",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 25000,
				},
			},
			expectError: true,
			errorField:  "nodePorts.management",
		},
		{
			name: "duplicate nodeports",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 30672,
				},
			},
			expectError: true,
			errorField:  "nodePorts",
		},
		{
			name: "invalid CPU format",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
				Resources: RabbitMQResources{
					CPU:    "invalid",
					Memory: "1Gi",
				},
			},
			expectError: true,
			errorField:  "resources.cpu",
		},
		{
			name: "invalid memory format",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
				Resources: RabbitMQResources{
					CPU:    "500m",
					Memory: "invalid",
				},
			},
			expectError: true,
			errorField:  "resources.memory",
		},
		{
			name: "persistence enabled but size missing",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
				Persistence: RabbitMQPersistence{
					Enabled: true,
					Size:    "",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "persistence enabled with invalid size format",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
				Persistence: RabbitMQPersistence{
					Enabled: true,
					Size:    "invalid",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "valid config with persistence enabled",
			config: RabbitMQConfig{
				Enabled:      true,
				ChartVersion: "11.0.0",
				VirtualHost:  "/",
				NodePorts: RabbitMQNodePorts{
					AMQP:       30672,
					Management: 31672,
				},
				Resources: RabbitMQResources{
					CPU:    "500m",
					Memory: "1Gi",
				},
				Persistence: RabbitMQPersistence{
					Enabled: true,
					Size:    "10Gi",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRabbitMQConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorField != "" {
					validationErr, ok := err.(*ValidationError)
					if !ok {
						t.Errorf("expected ValidationError but got %T", err)
						return
					}
					if validationErr.Field != tt.errorField {
						t.Errorf("expected error field %s but got %s", tt.errorField, validationErr.Field)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateVirtualHost(t *testing.T) {
	tests := []struct {
		name        string
		vhost       string
		expectError bool
	}{
		{
			name:        "valid virtual host root",
			vhost:       "/",
			expectError: false,
		},
		{
			name:        "valid virtual host with path",
			vhost:       "/dev",
			expectError: false,
		},
		{
			name:        "valid virtual host with underscore",
			vhost:       "/dev_env",
			expectError: false,
		},
		{
			name:        "valid virtual host alphanumeric",
			vhost:       "dev",
			expectError: false,
		},
		{
			name:        "empty virtual host (defaults to /)",
			vhost:       "",
			expectError: false,
		},
		{
			name:        "invalid virtual host with dash",
			vhost:       "dev-env",
			expectError: true,
		},
		{
			name:        "invalid virtual host starting with number",
			vhost:       "123dev",
			expectError: true,
		},
		{
			name:        "invalid virtual host with special characters",
			vhost:       "/dev@env",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVirtualHost(tt.vhost)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateDatabase(t *testing.T) {
	tests := []struct {
		name        string
		database    string
		expectError bool
	}{
		{
			name:        "valid database name",
			database:    "mydb",
			expectError: false,
		},
		{
			name:        "valid database with underscore",
			database:    "my_db",
			expectError: false,
		},
		{
			name:        "valid database starting with underscore",
			database:    "_mydb",
			expectError: false,
		},
		{
			name:        "empty database name",
			database:    "",
			expectError: true,
		},
		{
			name:        "database starting with number",
			database:    "123mydb",
			expectError: true,
		},
		{
			name:        "database with invalid characters",
			database:    "my-db",
			expectError: true,
		},
		{
			name:        "database too long",
			database:    "a" + string(make([]byte, 64)),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabase(tt.database)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateNodePort(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectError bool
	}{
		{
			name:        "valid nodeport at minimum",
			port:        30000,
			expectError: false,
		},
		{
			name:        "valid nodeport at maximum",
			port:        32767,
			expectError: false,
		},
		{
			name:        "valid nodeport in middle",
			port:        30306,
			expectError: false,
		},
		{
			name:        "invalid nodeport too low",
			port:        29999,
			expectError: true,
		},
		{
			name:        "invalid nodeport too high",
			port:        32768,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNodePort(tt.port)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateResources(t *testing.T) {
	tests := []struct {
		name        string
		cpu         string
		memory      string
		expectError bool
		errorField  string
	}{
		{
			name:        "valid CPU and memory",
			cpu:         "500m",
			memory:      "1Gi",
			expectError: false,
		},
		{
			name:        "valid CPU without m suffix",
			cpu:         "1",
			memory:      "512Mi",
			expectError: false,
		},
		{
			name:        "valid memory with different units",
			cpu:         "500m",
			memory:      "512Mi",
			expectError: false,
		},
		{
			name:        "invalid CPU format",
			cpu:         "invalid",
			memory:      "1Gi",
			expectError: true,
			errorField:  "resources.cpu",
		},
		{
			name:        "invalid memory format",
			cpu:         "500m",
			memory:      "invalid",
			expectError: true,
			errorField:  "resources.memory",
		},
		{
			name:        "empty CPU and memory should pass",
			cpu:         "",
			memory:      "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResources(tt.cpu, tt.memory)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorField != "" {
					validationErr, ok := err.(*ValidationError)
					if !ok {
						t.Errorf("expected ValidationError but got %T", err)
						return
					}
					if validationErr.Field != tt.errorField {
						t.Errorf("expected error field %s but got %s", tt.errorField, validationErr.Field)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateChartVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{
			name:        "valid semantic version",
			version:     "9.4.6",
			expectError: false,
		},
		{
			name:        "valid version with patch",
			version:     "9.4.6",
			expectError: false,
		},
		{
			name:        "empty version",
			version:     "",
			expectError: true,
		},
		{
			name:        "invalid version format",
			version:     "invalid",
			expectError: true,
		},
		{
			name:        "version without patch",
			version:     "9.4",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChartVersion(tt.version)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ValidationError",
			err:      &ValidationError{Field: "test", Value: "value", Reason: "reason"},
			expected: true,
		},
		{
			name:     "MySQLError",
			err:      &MySQLError{Operation: "test", Reason: "reason"},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidationError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestIsMySQLError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "MySQLError",
			err:      &MySQLError{Operation: "test", Reason: "reason"},
			expected: true,
		},
		{
			name:     "ValidationError",
			err:      &ValidationError{Field: "test", Value: "value", Reason: "reason"},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMySQLError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidatePersistenceConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      MySQLConfig
		expectError bool
		errorField  string
	}{
		{
			name: "persistence disabled - no size required",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: false,
					Size:    "",
				},
			},
			expectError: false,
		},
		{
			name: "persistence enabled with valid size",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "8Gi",
				},
			},
			expectError: false,
		},
		{
			name: "persistence enabled with different valid sizes",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "10Gi",
				},
			},
			expectError: false,
		},
		{
			name: "persistence enabled with Mi unit",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "512Mi",
				},
			},
			expectError: false,
		},
		{
			name: "persistence enabled but size missing",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "persistence enabled with invalid size format",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "8GB",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "persistence enabled with invalid size format - no unit",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "8",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "persistence enabled with invalid size format - lowercase",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "8gi",
				},
			},
			expectError: true,
			errorField:  "persistence.size",
		},
		{
			name: "persistence enabled with Ti unit",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "1Ti",
				},
			},
			expectError: false,
		},
		{
			name: "persistence enabled with Ki unit",
			config: MySQLConfig{
				Enabled:      true,
				ChartVersion: "9.4.6",
				Database:     "mydb",
				NodePorts: MySQLNodePorts{
					MySQL: 30306,
				},
				Persistence: MySQLPersistence{
					Enabled: true,
					Size:    "1024Ki",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMySQLConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorField != "" {
					validationErr, ok := err.(*ValidationError)
					if !ok {
						t.Errorf("expected ValidationError but got %T", err)
						return
					}
					if validationErr.Field != tt.errorField {
						t.Errorf("expected error field %s but got %s", tt.errorField, validationErr.Field)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper functions for testing CustomComponents

func createTempYAMLFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}

func cleanupTempFile(t *testing.T, path string) {
	// No-op since we use t.TempDir() which auto-cleans
}

func TestLoadConfigWithCustomComponents(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		checkFunc   func(*testing.T, *KindEnvConfig)
	}{
		{
			name: "valid custom component with minimal config",
			yamlContent: `
customComponents:
  - name: test-app
    image: nginx:latest
`,
			expectError: false,
			checkFunc: func(t *testing.T, config *KindEnvConfig) {
				if len(config.CustomComponents) != 1 {
					t.Fatalf("expected 1 custom component, got %d", len(config.CustomComponents))
				}
				cc := config.CustomComponents[0]
				if cc.Name != "test-app" {
					t.Errorf("expected name 'test-app', got '%s'", cc.Name)
				}
				if cc.Image != "nginx:latest" {
					t.Errorf("expected image 'nginx:latest', got '%s'", cc.Image)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempYAMLFile(t, tt.yamlContent)
			defer cleanupTempFile(t, tmpFile)

			config, err := LoadConfig(tmpFile)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}
}

func TestCustomComponentSetDefaults(t *testing.T) {
	cc := CustomComponent{
		Name:  "test-app",
		Image: "nginx:latest",
	}
	cc.SetDefaults()

	if cc.Namespace != "default" {
		t.Errorf("expected namespace 'default', got '%s'", cc.Namespace)
	}
	if cc.Replicas == nil || *cc.Replicas != 1 {
		t.Errorf("expected replicas 1, got %v", cc.Replicas)
	}
	if cc.Enabled == nil || !*cc.Enabled {
		t.Errorf("expected enabled true, got %v", cc.Enabled)
	}
	if cc.Resources == nil {
		t.Fatal("expected resources to be set")
	}
}
