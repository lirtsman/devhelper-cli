package kindenv

import (
	"fmt"
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
