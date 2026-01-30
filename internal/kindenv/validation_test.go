package kindenv

import (
	"testing"
)

// validation_test.go provides table-driven tests for custom component validation

func TestCustomComponentValidate(t *testing.T) {
	tests := []struct {
		name        string
		component   CustomComponent
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid minimal config",
			component: CustomComponent{
				Name:  "test-app",
				Image: "nginx:latest",
			},
			expectError: false,
		},
		{
			name: "missing name",
			component: CustomComponent{
				Image: "nginx:latest",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing image",
			component: CustomComponent{
				Name: "test-app",
			},
			expectError: true,
			errorMsg:    "image is required",
		},
		{
			name: "invalid DNS label name",
			component: CustomComponent{
				Name:  "Test_App",
				Image: "nginx:latest",
			},
			expectError: true,
			errorMsg:    "DNS label",
		},
		{
			name: "invalid replicas",
			component: CustomComponent{
				Name:     "test-app",
				Image:    "nginx:latest",
				Replicas: intPtr(0),
			},
			expectError: true,
			errorMsg:    "replicas must be >= 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.component.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEnvVarValidate(t *testing.T) {
	tests := []struct {
		name        string
		envVar      EnvVar
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid direct value",
			envVar: EnvVar{
				Name:  "APP_ENV",
				Value: "production",
			},
			expectError: false,
		},
		{
			name: "valid secret reference",
			envVar: EnvVar{
				Name: "DB_PASSWORD",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "mysql-secret",
						Key:  "password",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			envVar: EnvVar{
				Value: "value",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing value and valueFrom",
			envVar: EnvVar{
				Name: "VAR_NAME",
			},
			expectError: true,
			errorMsg:    "either 'value' or 'valueFrom'",
		},
		{
			name: "both value and valueFrom",
			envVar: EnvVar{
				Name:  "VAR_NAME",
				Value: "value",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "secret",
						Key:  "key",
					},
				},
			},
			expectError: true,
			errorMsg:    "mutually exclusive",
		},
		{
			name: "invalid env var name",
			envVar: EnvVar{
				Name:  "invalid-name",
				Value: "value",
			},
			expectError: true,
			errorMsg:    "invalid environment variable name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.envVar.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPortMappingValidate(t *testing.T) {
	tests := []struct {
		name        string
		port        PortMapping
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid port mapping",
			port: PortMapping{
				ContainerPort: 8080,
				HostPort:      8080,
				Protocol:      "TCP",
			},
			expectError: false,
		},
		{
			name: "invalid container port too low",
			port: PortMapping{
				ContainerPort: 0,
			},
			expectError: true,
			errorMsg:    "between 1 and 65535",
		},
		{
			name: "invalid container port too high",
			port: PortMapping{
				ContainerPort: 65536,
			},
			expectError: true,
			errorMsg:    "between 1 and 65535",
		},
		{
			name: "invalid nodePort too low",
			port: PortMapping{
				ContainerPort: 8080,
				NodePort:      29999,
			},
			expectError: true,
			errorMsg:    "between 30000 and 32767",
		},
		{
			name: "invalid protocol",
			port: PortMapping{
				ContainerPort: 8080,
				Protocol:      "HTTP",
			},
			expectError: true,
			errorMsg:    "TCP or UDP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.port.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigFileValidate(t *testing.T) {
	tests := []struct {
		name        string
		configFile  ConfigFile
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config file",
			configFile: ConfigFile{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "key: value",
			},
			expectError: false,
		},
		{
			name: "missing name",
			configFile: ConfigFile{
				Path:     "/config/app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing path",
			configFile: ConfigFile{
				Name:     "app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "path is required",
		},
		{
			name: "relative path",
			configFile: ConfigFile{
				Name:     "app.yaml",
				Path:     "config/app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "must be absolute",
		},
		{
			name: "name with directory separator",
			configFile: ConfigFile{
				Name:     "dir/app.yaml",
				Path:     "/config/app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "cannot contain directory separators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.configFile.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
